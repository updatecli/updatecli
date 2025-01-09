package httpclient

import (
	"bytes"
	"io"
	"math"
	"net/http"
	"time"
)

const RetryCount = 3

type retryTransport struct {
	transport http.RoundTripper
}

func backoff(retries int) time.Duration {
	return time.Duration(math.Pow(2, float64(retries))) * time.Second
}

func shouldRetry(err error, resp *http.Response) bool {
	if err != nil {
		return true
	}

	if resp.StatusCode == http.StatusBadGateway ||
		resp.StatusCode == http.StatusServiceUnavailable ||
		resp.StatusCode == http.StatusGatewayTimeout ||
		resp.StatusCode == http.StatusTooManyRequests {
		return true
	}

	return false
}

func drainBody(resp *http.Response) error {
	if resp != nil && resp.Body != nil {
		if _, err := io.Copy(io.Discard, resp.Body); err != nil {
			return err
		}
		if err := resp.Body.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request body
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, _ = io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	// Send the request
	resp, err := t.transport.RoundTrip(req)

	// Retry logic
	retries := 0
	for shouldRetry(err, resp) && retries < RetryCount {
		// Wait for the specified backoff period
		time.Sleep(backoff(retries))

		// We're going to retry, consume any response to reuse the connection.
		if err = drainBody(resp); err != nil {
			return resp, err
		}

		// Clone the request body again
		if req.Body != nil {
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Retry the request
		resp, err = t.transport.RoundTrip(req)

		retries++
	}

	// Return the response
	return resp, err
}

func NewRetryClient() HTTPClient {
	transport := &retryTransport{
		transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}

	client := http.DefaultClient
	client.Transport = transport
	return client
}
