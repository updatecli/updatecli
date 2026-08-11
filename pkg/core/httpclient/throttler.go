package httpclient

import (
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

type throttledTransport struct {
	roundTripperWrapper http.RoundTripper
	rateLimiter         *rate.Limiter
}

func (c *throttledTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	err := c.rateLimiter.Wait(req.Context()) // This is a blocking call. Honors the rate limit
	if err != nil {
		return nil, err
	}
	return c.roundTripperWrapper.RoundTrip(req)
}

func newThrottledTransport(limitPeriod time.Duration, requestCount int, transportWrap http.RoundTripper) http.RoundTripper {
	return &throttledTransport{
		roundTripperWrapper: transportWrap,
		rateLimiter:         rate.NewLimiter(rate.Every(limitPeriod), requestCount),
	}
}

// NewThrottledRetryClient returns an HTTP client with rate limiting wrapping retry + proxy.
func NewThrottledRetryClient(limitPeriod time.Duration, requestCount int) *http.Client {
	return &http.Client{
		Transport: newThrottledTransport(limitPeriod, requestCount, DefaultTransport()),
	}
}
