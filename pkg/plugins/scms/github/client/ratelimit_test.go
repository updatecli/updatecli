package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/shurcooL/githubv4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
)

func TestIsRateLimitError(t *testing.T) {
	tests := []struct {
		err  error
		want bool
	}{
		{nil, false},
		{errors.New("API rate limit already exceeded for user ID 1."), true},
		{errors.New("API rate limit exceeded for user ID 1."), true},
		{errors.New(`non-200 OK status code: 403 Forbidden body: "{\"message\":\"You have exceeded a secondary rate limit. Please wait a few minutes before you try again.\"}"`), true},
		{fmt.Errorf("querying: %w", errors.New(ErrRateLimitFinalAttempt)), false},
		{errors.New("Could not resolve to a Repository"), false},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, IsRateLimitError(tt.err), "%v", tt.err)
	}
}

func TestObserve(t *testing.T) {
	now := time.Date(2026, 11, 13, 10, 0, 0, 0, time.UTC)
	primary := errors.New("API rate limit exceeded for user ID 1.")
	secondary := errors.New("You have exceeded a secondary rate limit.")

	header := func(values map[string]string) http.Header {
		h := http.Header{}
		for k, v := range values {
			h.Set(k, v)
		}
		return h
	}

	tests := []struct {
		name   string
		header http.Header
		errs   []error
		want   time.Time
	}{
		{
			name:   "Retry-After wins over the reset",
			header: header(map[string]string{"Retry-After": "30", "X-RateLimit-Remaining": "0", "X-RateLimit-Reset": fmt.Sprint(now.Add(time.Hour).Unix())}),
			errs:   []error{secondary},
			want:   now.Add(30 * time.Second),
		},
		{
			name:   "no credit left waits for the reset",
			header: header(map[string]string{"X-RateLimit-Remaining": "0", "X-RateLimit-Reset": fmt.Sprint(now.Add(time.Hour).Unix())}),
			errs:   []error{primary},
			want:   now.Add(time.Hour),
		},
		{
			name:   "secondary limit without a delay waits a minute",
			header: header(map[string]string{"X-RateLimit-Remaining": "4000", "X-RateLimit-Reset": fmt.Sprint(now.Add(time.Hour).Unix())}),
			errs:   []error{secondary},
			want:   now.Add(time.Minute),
		},
		{
			name:   "secondary limit refused again doubles the wait",
			header: header(map[string]string{"X-RateLimit-Remaining": "4000", "X-RateLimit-Reset": fmt.Sprint(now.Add(time.Hour).Unix())}),
			errs:   []error{secondary, secondary, secondary},
			want:   now.Add(4 * time.Minute),
		},
		{
			name:   "a success resets the backoff",
			header: header(map[string]string{"X-RateLimit-Remaining": "4000", "X-RateLimit-Reset": fmt.Sprint(now.Add(time.Hour).Unix())}),
			errs:   []error{secondary, nil, secondary},
			want:   now.Add(time.Minute),
		},
		{
			name:   "a delay beyond an hour is capped",
			header: header(map[string]string{"Retry-After": "99999999999999"}),
			errs:   []error{secondary},
			want:   now.Add(time.Hour),
		},
		{
			name:   "a reset beyond an hour is capped",
			header: header(map[string]string{"X-RateLimit-Remaining": "0", "X-RateLimit-Reset": fmt.Sprint(now.Add(48 * time.Hour).Unix())}),
			errs:   []error{primary},
			want:   now.Add(time.Hour),
		},
		{
			name:   "a past Retry-After falls back to the reset",
			header: header(map[string]string{"Retry-After": now.Add(-time.Minute).Format(http.TimeFormat), "X-RateLimit-Remaining": "0", "X-RateLimit-Reset": fmt.Sprint(now.Add(time.Hour).Unix())}),
			errs:   []error{primary},
			want:   now.Add(time.Hour),
		},
		{
			name:   "other errors don't hold requests",
			header: header(map[string]string{"X-RateLimit-Remaining": "0", "X-RateLimit-Reset": fmt.Sprint(now.Add(time.Hour).Unix())}),
			errs:   []error{errors.New("Could not resolve to a Repository")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &rateLimitRecorder{}
			h := parseRateLimitHeaders(tt.header, now)
			for _, err := range tt.errs {
				r.holdUntil = time.Time{}
				r.observe(err, h, now, now)
			}
			assert.True(t, tt.want.Equal(r.holdUntil), "want %s, got %s", tt.want, r.holdUntil)
		})
	}
}

// TestMutateWaitsAfterSecondaryRateLimit checks that a mutation refused for a
// secondary rate limit isn't sent again before Retry-After elapsed.
func TestMutateWaitsAfterSecondaryRateLimit(t *testing.T) {
	var calls int32
	var first time.Time
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if atomic.AddInt32(&calls, 1) == 1 {
			first = time.Now()
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, `{"message":"You have exceeded a secondary rate limit."}`)
			return
		}
		fmt.Fprint(w, `{"data":{"addStar":{"clientMutationId":"1"}}}`)
	}))
	defer srv.Close()

	c := newRateLimitClient(httpclient.NewUncachedGraphQLClient(), srv.URL)

	var m struct {
		AddStar struct {
			ClientMutationID string
		} `graphql:"addStar(input: $input)"`
	}
	input := githubv4.AddStarInput{StarrableID: githubv4.ID("1")}

	err := c.Mutate(context.Background(), &m, input, nil)
	require.True(t, IsRateLimitError(err), "%v", err)

	require.NoError(t, c.Mutate(context.Background(), &m, input, nil))
	assert.Equal(t, int32(2), atomic.LoadInt32(&calls))
	assert.GreaterOrEqual(t, time.Since(first), time.Second)
}

func TestWaitStopsWhenContextIsDone(t *testing.T) {
	r := &rateLimitRecorder{holdUntil: time.Now().Add(time.Hour)}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	assert.ErrorIs(t, r.wait(ctx), context.DeadlineExceeded)
}

// TestObserveKeepsBackoffAfterStaleSuccess checks that a success sent before
// the last refusal, by a call running in parallel, doesn't reset the backoff.
func TestObserveKeepsBackoffAfterStaleSuccess(t *testing.T) {
	now := time.Date(2026, 11, 13, 10, 0, 0, 0, time.UTC)
	secondary := errors.New("You have exceeded a secondary rate limit.")

	r := &rateLimitRecorder{}
	r.observe(secondary, rateLimitHeaders{}, now, now)
	r.observe(nil, rateLimitHeaders{}, now.Add(-time.Second), now)
	r.observe(secondary, rateLimitHeaders{}, now, now)

	assert.True(t, now.Add(2*time.Minute).Equal(r.holdUntil), "got %s", r.holdUntil)
}

// TestTransportFillsTheSlotOfItsOwnCall checks that each call gets the
// headers of its own response, whatever response the recorder saw last.
func TestTransportFillsTheSlotOfItsOwnCall(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/refused" {
			w.Header().Set("Retry-After", "300")
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.Header().Set("X-RateLimit-Remaining", "4000")
		w.Header().Set("X-RateLimit-Reset", "2000000000")
	}))
	defer srv.Close()

	recorder := &rateLimitRecorder{}
	transport := &rateLimitTransport{transport: httpclient.UncachedTransport(), recorder: recorder}

	send := func(ctx context.Context, path string) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+path, nil)
		require.NoError(t, err)
		resp, err := transport.RoundTrip(req)
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())
	}

	var refused rateLimitHeaders
	start := time.Now()
	send(context.WithValue(context.Background(), responseSlotKey{}, &refused), "/refused")
	// A parallel call answered after the refusal.
	send(context.Background(), "/ok")

	assert.WithinDuration(t, start.Add(300*time.Second), refused.retryAfter, 5*time.Second)
	remaining, _, ok := recorder.lastRateLimit()
	assert.True(t, ok)
	assert.Equal(t, 4000, remaining)
}

func TestWaitHonoursAnExtendedHold(t *testing.T) {
	r := &rateLimitRecorder{holdUntil: time.Now().Add(100 * time.Millisecond)}
	extended := time.Now().Add(400 * time.Millisecond)

	go func() {
		time.Sleep(20 * time.Millisecond)
		r.mu.Lock()
		r.holdUntil = extended
		r.mu.Unlock()
	}()

	require.NoError(t, r.wait(context.Background()))
	assert.False(t, time.Now().Before(extended))
}
