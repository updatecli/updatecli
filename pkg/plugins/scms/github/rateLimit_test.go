package github

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github/client"
)

// newRateLimitedServer returns a fake GitHub GraphQL API refusing every
// request, the rateLimit query included, until reset, as GitHub does once the
// rate limit is exceeded. refuse writes the refused response, headers included.
func newRateLimitedServer(t *testing.T, reset time.Time, refuse func(w http.ResponseWriter)) (*httptest.Server, *int32) {
	t.Helper()

	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("reading request body: %s", err)
		}
		w.Header().Set("Content-Type", "application/json")

		if !time.Now().Before(reset) {
			w.Header().Set("X-RateLimit-Remaining", "4999")
			w.Header().Set("X-RateLimit-Reset", fmt.Sprint(reset.Add(time.Hour).Unix()))
			if !strings.Contains(string(body), "repository") {
				fmt.Fprint(w, `{"data":{"rateLimit":{"cost":1,"remaining":4999,"resetAt":"2099-01-01T00:00:00Z"}}}`)
				return
			}
			fmt.Fprint(w, `{"data":{"rateLimit":{"cost":1,"remaining":4999,"resetAt":"2099-01-01T00:00:00Z"},`+
				`"repository":{"refs":{"totalCount":1,"pageInfo":{"hasPreviousPage":false},"edges":[{"node":{"name":"v1.0.0"}}]}}}}`)
			return
		}

		refuse(w)
	}))
	t.Cleanup(srv.Close)

	return srv, &calls
}

func TestSearchTagsWaitsForTheRateLimit(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("UPDATECLI_GITHUB_TOKEN", "")

	defaultWait := client.SecondaryRateLimitWait
	client.SecondaryRateLimitWait = time.Second
	t.Cleanup(func() { client.SecondaryRateLimitWait = defaultWait })

	const (
		primary   = "API rate limit already exceeded for user ID 1."
		secondary = "You have exceeded a secondary rate limit. Please wait a few minutes before you try again."
	)

	tests := []struct {
		name   string
		refuse func(w http.ResponseWriter, reset time.Time)
	}{
		{
			name: "primary limit as a GraphQL error",
			refuse: func(w http.ResponseWriter, reset time.Time) {
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("X-RateLimit-Reset", fmt.Sprint(reset.Unix()))
				fmt.Fprintf(w, `{"data":null,"errors":[{"type":"RATE_LIMITED","message":%q}]}`, primary)
			},
		},
		{
			name: "primary limit as a 403",
			refuse: func(w http.ResponseWriter, reset time.Time) {
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("X-RateLimit-Reset", fmt.Sprint(reset.Unix()))
				w.WriteHeader(http.StatusForbidden)
				fmt.Fprintf(w, `{"message":%q}`, primary)
			},
		},
		{
			name: "secondary limit as a 403 with Retry-After",
			refuse: func(w http.ResponseWriter, reset time.Time) {
				w.Header().Set("X-RateLimit-Remaining", "4000")
				w.Header().Set("X-RateLimit-Reset", fmt.Sprint(reset.Add(time.Hour).Unix()))
				w.Header().Set("Retry-After", fmt.Sprint(int(time.Until(reset).Seconds()+1)))
				w.WriteHeader(http.StatusForbidden)
				fmt.Fprintf(w, `{"message":%q}`, secondary)
			},
		},
		{
			name: "secondary limit as a GraphQL error without delay",
			refuse: func(w http.ResponseWriter, reset time.Time) {
				w.Header().Set("X-RateLimit-Remaining", "4000")
				w.Header().Set("X-RateLimit-Reset", fmt.Sprint(reset.Add(time.Hour).Unix()))
				fmt.Fprintf(w, `{"data":null,"errors":[{"type":"RATE_LIMITED","message":%q}]}`, secondary)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// GitHub sends the reset time in whole seconds.
			reset := time.Now().Add(2 * time.Second).Truncate(time.Second)
			srv, calls := newRateLimitedServer(t, reset, func(w http.ResponseWriter) { tt.refuse(w, reset) })

			c, err := client.New("user", "token", nil, srv.URL)
			require.NoError(t, err)

			g := &Github{client: c.Client, Spec: Spec{Owner: "owner", Repository: "repo"}}

			tags, err := g.SearchTags(context.Background(), 0)
			require.NoError(t, err)
			assert.Equal(t, []string{"v1.0.0"}, tags)
			assert.False(t, time.Now().Before(reset), "SearchTags returned before the rate limit reset")
			// No request is sent while on hold, so the refused query is only
			// followed by the rateLimit query and the retry.
			assert.LessOrEqual(t, atomic.LoadInt32(calls), int32(4))
		})
	}
}

func TestPauseWithoutRateLimit(t *testing.T) {
	var rateLimit *RateLimit
	assert.NotPanics(t, func() { rateLimit.Pause(context.Background()) })
}

func TestPauseStopsWhenContextIsDone(t *testing.T) {
	rateLimit := &RateLimit{
		Remaining: 0,
		ResetAt:   time.Now().Add(time.Hour).UTC().Format(time.RFC3339),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	rateLimit.Pause(ctx)
	assert.Less(t, time.Since(start), 10*time.Second)
}
