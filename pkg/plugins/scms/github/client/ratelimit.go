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

// MaxRateLimitHold bounds how long requests are held after a refusal. GitHub
// resets its rate limit every hour, so a longer delay comes from a broken or
// hostile server and would otherwise stall the run, or overflow.
const MaxRateLimitHold = time.Hour

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

// rateLimitHeaders holds the rate limit headers of one response.
type rateLimitHeaders struct {
	// retryAfter is when the response allows the next request through its
	// Retry-After header, zero if it had none.
	retryAfter time.Time
	// remaining and resetAt come from the X-RateLimit headers, which GitHub
	// sends on every response, a refused request included. ok is false when
	// the response had none.
	remaining int
	resetAt   time.Time
	ok        bool
}

func parseRateLimitHeaders(header http.Header, now time.Time) rateLimitHeaders {
	var h rateLimitHeaders

	if v := strings.TrimSpace(header.Get("Retry-After")); v != "" {
		if seconds, err := strconv.ParseInt(v, 10, 64); err == nil && seconds >= 0 {
			h.retryAfter = now.Add(time.Duration(min(seconds, int64(MaxRateLimitHold/time.Second))) * time.Second)
		} else if date, err := http.ParseTime(v); err == nil {
			h.retryAfter = date
		}
	}

	remaining, err := strconv.Atoi(strings.TrimSpace(header.Get("X-RateLimit-Remaining")))
	if err != nil {
		return h
	}
	reset, err := strconv.ParseInt(strings.TrimSpace(header.Get("X-RateLimit-Reset")), 10, 64)
	if err != nil {
		return h
	}

	h.remaining = remaining
	h.resetAt = time.Unix(reset, 0)
	h.ok = true
	return h
}

// responseSlotKey is the context key of the *rateLimitHeaders that receives
// the rate limit headers of the response to a single call, so a refusal is
// judged by its own response even when calls run in parallel.
type responseSlotKey struct{}

// rateLimitRecorder holds the rate limit of the last response, and the time
// before which no request may be sent after GitHub refused one for exceeding
// a rate limit.
type rateLimitRecorder struct {
	mu sync.Mutex

	// last holds the headers of the last response with rate limit headers,
	// whatever call it answered, for queryRateLimit to fall back on.
	last rateLimitHeaders

	// holdUntil is when requests may be sent again after a refusal.
	holdUntil time.Time
	// lastRefusal is when GitHub last refused a request for a rate limit.
	lastRefusal time.Time
	// secondaryRefusals counts the secondary rate limit refusals in a row.
	secondaryRefusals int
}

func (r *rateLimitRecorder) record(h rateLimitHeaders) {
	if !h.ok {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.last = h
}

func (r *rateLimitRecorder) lastRateLimit() (int, time.Time, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.last.remaining, r.last.resetAt, r.last.ok
}

// observe sets how long to hold the next requests after a call sent at sent
// ended with err and a response carrying h, following GitHub's guidance: wait
// as long as Retry-After says, or else until the rate limit reset when no
// credit is left, or else, after a secondary rate limit, at least a minute,
// doubled with each refusal in a row.
func (r *rateLimitRecorder) observe(err error, h rateLimitHeaders, sent, now time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !IsRateLimitError(err) {
		// Only a success sent after the last refusal shows the limit is lifted.
		if err == nil && !sent.Before(r.lastRefusal) {
			r.secondaryRefusals = 0
		}
		return
	}

	var until time.Time
	switch {
	case h.retryAfter.After(now):
		until = h.retryAfter
	case h.ok && h.remaining == 0 && h.resetAt.After(now):
		until = h.resetAt
	case isSecondaryRateLimitError(err):
		until = now.Add(SecondaryRateLimitWait << min(r.secondaryRefusals, 5))
	}

	if isSecondaryRateLimitError(err) {
		r.secondaryRefusals++
	}
	r.lastRefusal = now

	until = minTime(until, now.Add(MaxRateLimitHold))
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

// wait blocks until requests may be sent again, or until ctx is done. It
// waits again if the hold was extended in the meantime.
func (r *rateLimitRecorder) wait(ctx context.Context) error {
	for {
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
		select {
		case <-timer.C:
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		}
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
		h := parseRateLimitHeaders(resp.Header, time.Now())
		t.recorder.record(h)
		if slot, ok := req.Context().Value(responseSlotKey{}).(*rateLimitHeaders); ok {
			*slot = h
		}
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
	return c.do(ctx, func(ctx context.Context) error {
		return c.client.Query(ctx, q, variables)
	})
}

// Mutate implements Client.
func (c *rateLimitClient) Mutate(ctx context.Context, m interface{}, input githubv4.Input, variables map[string]interface{}) error {
	return c.do(ctx, func(ctx context.Context) error {
		return c.client.Mutate(ctx, m, input, variables)
	})
}

// do waits until requests may be sent, sends one with send, and sets how long
// to hold the next ones from its outcome and its own response headers.
func (c *rateLimitClient) do(ctx context.Context, send func(context.Context) error) error {
	if err := c.recorder.wait(ctx); err != nil {
		return err
	}

	var h rateLimitHeaders
	sent := time.Now()
	err := send(context.WithValue(ctx, responseSlotKey{}, &h))
	c.recorder.observe(err, h, sent, time.Now())
	return err
}

// LastRateLimit implements RateLimitReporter.
func (c *rateLimitClient) LastRateLimit() (int, time.Time, bool) {
	return c.recorder.lastRateLimit()
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
