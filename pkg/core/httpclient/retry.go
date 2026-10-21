package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/updatecli/updatecli/pkg/core/log"
	version "github.com/updatecli/updatecli/pkg/core/version"
)

const RetryCount = 3

type retryTransport struct {
	transport http.RoundTripper
}

// MaxRetryWait bounds the total time spent waiting between the retries of one
// request. A wait that would go beyond it is skipped, and the last response is
// returned as is.
const MaxRetryWait = 2 * time.Minute

func backoff(retries int) time.Duration {
	return time.Duration(math.Pow(2, float64(retries))) * time.Second
}

// serverDelay returns how long the server asks the client to wait before
// sending the request again, read from the Retry-After header, either seconds
// or an HTTP date. On a 429 only, it falls back to a rate limit reset header:
// GitHub and GitLab send one on every response, where it marks the end of the
// rate limit window rather than a delay to wait for. GitLab sends
// Ratelimit-Reset and GitHub X-RateLimit-Reset, both as a Unix timestamp,
// while the IETF draft RateLimit-Reset sends seconds, so a small value is
// read as seconds.
func serverDelay(resp *http.Response, now time.Time) (time.Duration, bool) {
	header := resp.Header

	if v := strings.TrimSpace(header.Get("Retry-After")); v != "" {
		if seconds, err := strconv.ParseInt(v, 10, 64); err == nil && seconds >= 0 {
			return time.Duration(seconds) * time.Second, true
		}
		if date, err := http.ParseTime(v); err == nil {
			return max(date.Sub(now), 0), true
		}
	}

	if resp.StatusCode != http.StatusTooManyRequests {
		return 0, false
	}

	for _, name := range []string{"Ratelimit-Reset", "X-Ratelimit-Reset"} {
		v, err := strconv.ParseInt(strings.TrimSpace(header.Get(name)), 10, 64)
		if err != nil || v < 0 {
			continue
		}
		// A Unix timestamp is far larger than any delay in seconds.
		if v > 1_000_000_000 {
			return max(time.Unix(v, 0).Sub(now), 0), true
		}
		return time.Duration(v) * time.Second, true
	}

	return 0, false
}

// retryDelay returns how long to wait before retry number retries: as long as
// the server asks, or else the backoff period.
func retryDelay(retries int, resp *http.Response, now time.Time) time.Duration {
	if resp != nil {
		if delay, ok := serverDelay(resp, now); ok {
			return delay
		}
	}
	return backoff(retries)
}

// sleep waits for d, or until ctx is done, in which case it returns false.
func sleep(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
		return true
	case <-ctx.Done():
		return false
	}
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
		req.Header.Set("User-Agent", UserAgent())
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
	var waited time.Duration
	for shouldRetry(err, resp) && retries < RetryCount {
		delay := retryDelay(retries, resp, time.Now())
		if waited+delay > MaxRetryWait {
			break
		}

		// We're going to retry, consume any response to reuse the connection.
		// Do it before waiting, so the connection isn't held for the whole wait.
		if err = drainBody(resp); err != nil {
			return resp, err
		}

		log.Debugf("retrying %s %s in %s", req.Method, req.URL.Redacted(), delay)
		if !sleep(req.Context(), delay) {
			return nil, req.Context().Err()
		}
		waited += delay

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

// UserAgent returns the User-Agent identifying updatecli traffic.
func UserAgent() string {
	if version.Version != "" {
		return "Updatecli/" + version.Version
	}
	return "Updatecli"
}

func NewRetryClient() *http.Client {
	return &http.Client{Transport: DefaultTransport()}
}

// NewUncachedRetryClient returns an HTTP client with retry and proxy support
// that bypasses the HTTP cache. See UncachedTransport for when to use it.
func NewUncachedRetryClient() *http.Client {
	return &http.Client{Transport: UncachedTransport()}
}
