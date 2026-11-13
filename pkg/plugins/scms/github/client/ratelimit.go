package client

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shurcooL/githubv4"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
	"github.com/updatecli/updatecli/pkg/core/log"
)

const (
	// ErrRateLimitFinalAttempt is the message of the error returned once a
	// request was refused for exceeding a rate limit more than MaxRetry times.
	ErrRateLimitFinalAttempt = "API rate limit exceeded, final attempt failed"
)

var (
	// SecondaryRateLimitWait is how long to wait after GitHub refused a request
	// for exceeding a secondary rate limit without saying for how long. It
	// doubles with each refusal in a row, as GitHub asks.
	SecondaryRateLimitWait = time.Minute
)

// maxRateLimitHold bounds how long requests are held after a refusal. GitHub
// resets its rate limit every hour, so a longer delay comes from a broken or
// hostile server and would otherwise stall the run, or overflow.
const maxRateLimitHold = time.Hour

// IsRateLimitError reports whether err means GitHub refused a request for
// exceeding its primary or a secondary rate limit.
func IsRateLimitError(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())

	// Our own error once every retry failed must not start a new round of retries.
	if strings.Contains(msg, strings.ToLower(ErrRateLimitFinalAttempt)) {
		return false
	}

	return strings.Contains(msg, "api rate limit exceeded") ||
		strings.Contains(msg, "api rate limit already exceeded") ||
		strings.Contains(msg, "secondary rate limit")
}

func isSecondaryRateLimitError(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "secondary rate limit")
}

// RateLimitReporter is implemented by a Client that remembers the rate limit
// GitHub reported in the headers of its last response.
type RateLimitReporter interface {
	// LastRateLimit returns the remaining credit and the reset time GitHub
	// sent with the last response, and false if no response carried them yet.
	LastRateLimit() (remaining int, resetAt time.Time, ok bool)
}

// rateLimitRecorder holds the rate limit read from the last response headers,
// and the time before which no request may be sent after GitHub refused one
// for exceeding a rate limit.
type rateLimitRecorder struct {
	mu sync.Mutex

	// remaining and resetAt come from the X-RateLimit headers, which GitHub
	// sends on every response, a refused request included, when a rateLimit
	// query can't be answered either. ok is false until a response carried them.
	remaining int
	resetAt   time.Time
	ok        bool

	// retryAfter is when the last response allowed the next request through
	// its Retry-After header, zero if it had none.
	retryAfter time.Time

	// holdUntil is when requests may be sent again after a refusal.
	holdUntil time.Time
	// secondaryRefusals counts the secondary rate limit refusals in a row.
	secondaryRefusals int
}

func (r *rateLimitRecorder) record(header http.Header, now time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.retryAfter = time.Time{}
	if v := strings.TrimSpace(header.Get("Retry-After")); v != "" {
		if seconds, err := strconv.ParseInt(v, 10, 64); err == nil && seconds >= 0 {
			r.retryAfter = now.Add(time.Duration(min(seconds, int64(maxRateLimitHold/time.Second))) * time.Second)
		} else if date, err := http.ParseTime(v); err == nil {
			r.retryAfter = date
		}
	}

	remaining, err := strconv.Atoi(strings.TrimSpace(header.Get("X-RateLimit-Remaining")))
	if err != nil {
		return
	}
	reset, err := strconv.ParseInt(strings.TrimSpace(header.Get("X-RateLimit-Reset")), 10, 64)
	if err != nil {
		return
	}

	r.remaining = remaining
	r.resetAt = time.Unix(reset, 0)
	r.ok = true
}

func (r *rateLimitRecorder) last() (int, time.Time, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.remaining, r.resetAt, r.ok
}

// observe sets how long to hold the next request after one ended with err,
// following GitHub's guidance: wait as long as Retry-After says, or else until
// the rate limit reset when no credit is left, or else, after a secondary
// rate limit, at least a minute, doubled with each refusal in a row.
func (r *rateLimitRecorder) observe(err error, now time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !IsRateLimitError(err) {
		if err == nil {
			r.secondaryRefusals = 0
		}
		return
	}

	var until time.Time
	switch {
	case !r.retryAfter.IsZero():
		until = r.retryAfter
	case r.ok && r.remaining == 0 && r.resetAt.After(now):
		until = r.resetAt
	case isSecondaryRateLimitError(err):
		until = now.Add(SecondaryRateLimitWait << min(r.secondaryRefusals, 5))
	}

	if isSecondaryRateLimitError(err) {
		r.secondaryRefusals++
	}

	until = minTime(until, now.Add(maxRateLimitHold))
	if until.After(r.holdUntil) {
		r.holdUntil = until
	}
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

// wait blocks until requests may be sent again, or until ctx is done.
func (r *rateLimitRecorder) wait(ctx context.Context) error {
	r.mu.Lock()
	until := r.holdUntil
	r.mu.Unlock()

	d := time.Until(until)
	if d <= 0 {
		return nil
	}

	log.Warningf(
		"GitHub API rate limit reached, on hold for %s until %s. The process will resume automatically.\n",
		d.Round(time.Second),
		until.UTC().Format("2006-01-02 15:04:05 UTC"))

	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// rateLimitTransport records the rate limit headers of each response.
type rateLimitTransport struct {
	transport http.RoundTripper
	recorder  *rateLimitRecorder
}

func (t *rateLimitTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.transport.RoundTrip(req)
	if resp != nil {
		t.recorder.record(resp.Header, time.Now())
	}
	return resp, err
}

// rateLimitClient is a GitHub GraphQL client that remembers the rate limit
// headers of its last response and, after GitHub refused a request for
// exceeding a rate limit, holds the next ones until GitHub allows them.
type rateLimitClient struct {
	client   *githubv4.Client
	recorder *rateLimitRecorder
}

// Query implements Client.
func (c *rateLimitClient) Query(ctx context.Context, q interface{}, variables map[string]interface{}) error {
	if err := c.recorder.wait(ctx); err != nil {
		return err
	}
	err := c.client.Query(ctx, q, variables)
	c.recorder.observe(err, time.Now())
	return err
}

// Mutate implements Client.
func (c *rateLimitClient) Mutate(ctx context.Context, m interface{}, input githubv4.Input, variables map[string]interface{}) error {
	if err := c.recorder.wait(ctx); err != nil {
		return err
	}
	err := c.client.Mutate(ctx, m, input, variables)
	c.recorder.observe(err, time.Now())
	return err
}

// LastRateLimit implements RateLimitReporter.
func (c *rateLimitClient) LastRateLimit() (int, time.Time, bool) {
	return c.recorder.last()
}

// newRateLimitClient returns a GitHub GraphQL client sending its requests
// with httpClient to graphqlURL, or to github.com when graphqlURL is empty.
func newRateLimitClient(httpClient *http.Client, graphqlURL string) *rateLimitClient {
	recorder := &rateLimitRecorder{}

	transport := httpClient.Transport
	if transport == nil {
		transport = httpclient.UncachedTransport()
	}
	httpClient.Transport = &rateLimitTransport{transport: transport, recorder: recorder}

	var c *githubv4.Client
	if graphqlURL == "" {
		c = githubv4.NewClient(httpClient)
	} else {
		c = githubv4.NewEnterpriseClient(graphqlURL, httpClient)
	}

	return &rateLimitClient{client: c, recorder: recorder}
}
