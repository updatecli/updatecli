package client

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shurcooL/githubv4"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
)

// RateLimitReporter is implemented by a Client that remembers the rate limit
// GitHub reported in the headers of its last response.
type RateLimitReporter interface {
	// LastRateLimit returns the remaining credit and the reset time GitHub
	// sent with the last response, and false if no response carried them yet.
	LastRateLimit() (remaining int, resetAt time.Time, ok bool)
}

// rateLimitRecorder holds the rate limit read from the last response headers.
// GitHub sends them on every response, including a request it refuses because
// the rate limit is exceeded, when a rateLimit query can't be answered either.
type rateLimitRecorder struct {
	mu        sync.Mutex
	remaining int
	resetAt   time.Time
	ok        bool
}

func (r *rateLimitRecorder) record(header http.Header) {
	remaining, err := strconv.Atoi(strings.TrimSpace(header.Get("X-RateLimit-Remaining")))
	if err != nil {
		return
	}
	reset, err := strconv.ParseInt(strings.TrimSpace(header.Get("X-RateLimit-Reset")), 10, 64)
	if err != nil {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.remaining = remaining
	r.resetAt = time.Unix(reset, 0)
	r.ok = true
}

func (r *rateLimitRecorder) last() (int, time.Time, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.remaining, r.resetAt, r.ok
}

// rateLimitTransport records the rate limit headers of each response.
type rateLimitTransport struct {
	transport http.RoundTripper
	recorder  *rateLimitRecorder
}

func (t *rateLimitTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.transport.RoundTrip(req)
	if resp != nil {
		t.recorder.record(resp.Header)
	}
	return resp, err
}

// rateLimitClient is a GitHub GraphQL client that remembers the rate limit
// headers of its last response.
type rateLimitClient struct {
	*githubv4.Client
	recorder *rateLimitRecorder
}

// LastRateLimit implements RateLimitReporter.
func (c *rateLimitClient) LastRateLimit() (int, time.Time, bool) {
	return c.recorder.last()
}

// newRateLimitClient returns a GitHub GraphQL client sending its requests
// with httpClient to graphqlURL, or to github.com when graphqlURL is empty,
// and recording the rate limit headers of each response.
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

	return &rateLimitClient{Client: c, recorder: recorder}
}
