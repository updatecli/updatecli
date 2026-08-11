package httpclient

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	version "github.com/updatecli/updatecli/pkg/core/version"
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

	if resp == nil {
		return false
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
	req = req.Clone(req.Context())

	// Inject a recognizable User-Agent so servers can identify updatecli traffic.
	// We only override the header when it is absent or set to Go's default value.
	if ua := req.Header.Get("User-Agent"); ua == "" || ua == "Go-http-client/1.1" || ua == "Go-http-client/2.0" {
		ua = "Updatecli"
		if version.Version != "" {
			ua = "Updatecli/" + version.Version
		}
		req.Header.Set("User-Agent", ua)
	}

	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("reading request body: %w", err)
		}
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

	return resp, err
}

func NewRetryClient() *http.Client {
	return &http.Client{Transport: DefaultTransport()}
}
