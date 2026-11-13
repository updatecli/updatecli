package github

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github/client"
)

// newRateLimitedServer returns a fake GitHub GraphQL API refusing every
// request, the rateLimit query included, until reset, as GitHub does once the
// rate limit is exceeded. refuse writes the refused response.
func newRateLimitedServer(t *testing.T, reset time.Time, refuse func(w http.ResponseWriter)) (*httptest.Server, *int32) {
	t.Helper()

	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")

		if !time.Now().Before(reset) {
			w.Header().Set("X-RateLimit-Remaining", "4999")
			w.Header().Set("X-RateLimit-Reset", fmt.Sprint(reset.Add(time.Hour).Unix()))
			fmt.Fprint(w, `{"data":{"rateLimit":{"cost":1,"remaining":4999,"resetAt":"2099-01-01T00:00:00Z"},`+
				`"repository":{"refs":{"totalCount":1,"pageInfo":{"hasPreviousPage":false},"edges":[{"node":{"name":"v1.0.0"}}]}}}}`)
			return
		}

		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("X-RateLimit-Reset", fmt.Sprint(reset.Unix()))
		refuse(w)
	}))
	t.Cleanup(srv.Close)

	return srv, &calls
}

func TestSearchTagsWaitsWhenRateLimitQueryIsRefused(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("UPDATECLI_GITHUB_TOKEN", "")

	const message = "API rate limit already exceeded for user ID 1."

	tests := []struct {
		name   string
		refuse func(w http.ResponseWriter)
	}{
		{
			name: "GraphQL error",
			refuse: func(w http.ResponseWriter) {
				fmt.Fprintf(w, `{"data":null,"errors":[{"type":"RATE_LIMITED","message":%q}]}`, message)
			},
		},
		{
			name: "403 Forbidden",
			refuse: func(w http.ResponseWriter) {
				w.WriteHeader(http.StatusForbidden)
				fmt.Fprintf(w, `{"message":%q}`, message)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// GitHub sends the reset time in whole seconds.
			reset := time.Now().Add(2 * time.Second).Truncate(time.Second)
			srv, calls := newRateLimitedServer(t, reset, tt.refuse)

			c, err := client.New("user", "token", nil, srv.URL)
			require.NoError(t, err)

			g := &Github{client: c.Client, Spec: Spec{Owner: "owner", Repository: "repo"}}

			tags, err := g.SearchTags(context.Background(), 0)
			require.NoError(t, err)
			assert.Equal(t, []string{"v1.0.0"}, tags)
			assert.False(t, time.Now().Before(reset), "SearchTags returned before the rate limit reset")
			// The refused query, the refused rateLimit query, then the retry.
			assert.Equal(t, int32(3), atomic.LoadInt32(calls))
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
