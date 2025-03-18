package httpclient

import (
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

type ThrottledTransport struct {
	roundTripperWrapper http.RoundTripper
	rateLimiter         *rate.Limiter
}

func (c *ThrottledTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	err := c.rateLimiter.Wait(req.Context()) // This is a blocking call. Honors the rate limit
	if err != nil {
		return nil, err
	}
	return c.roundTripperWrapper.RoundTrip(req)
}

func NewThrottledTransport(limitPeriod time.Duration, requestCount int, transportWrap http.RoundTripper) http.RoundTripper {
	return &ThrottledTransport{
		roundTripperWrapper: transportWrap,
		rateLimiter:         rate.NewLimiter(rate.Every(limitPeriod), requestCount),
	}
}

func NewThrottledClient(limitPeriod time.Duration, requestCount int, transportWrap http.RoundTripper) HTTPClient {
	client := &http.Client{}
	client.Transport = NewThrottledTransport(limitPeriod, requestCount, transportWrap)
	return client
}
