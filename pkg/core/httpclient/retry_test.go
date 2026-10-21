package httpclient

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// roundTripFunc adapts a function into an http.RoundTripper.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestRetryDelay(t *testing.T) {
	now := time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC)
	inTwenty := strconv.FormatInt(now.Add(20*time.Second).Unix(), 10)

	tests := []struct {
		name   string
		status int
		header map[string]string
		want   time.Duration
	}{
		{name: "no header falls back to backoff", want: time.Second},
		{name: "Retry-After seconds", header: map[string]string{"Retry-After": "30"}, want: 30 * time.Second},
		{name: "Retry-After date", header: map[string]string{"Retry-After": now.Add(45 * time.Second).Format(http.TimeFormat)}, want: 45 * time.Second},
		{name: "Retry-After date in the past", header: map[string]string{"Retry-After": now.Add(-time.Minute).Format(http.TimeFormat)}, want: 0},
		{name: "Retry-After invalid falls back to backoff", header: map[string]string{"Retry-After": "soon"}, want: time.Second},
		{name: "Retry-After on 503", status: http.StatusServiceUnavailable, header: map[string]string{"Retry-After": "30"}, want: 30 * time.Second},
		{name: "GitLab Ratelimit-Reset timestamp", header: map[string]string{"Ratelimit-Reset": inTwenty}, want: 20 * time.Second},
		{name: "GitHub X-RateLimit-Reset timestamp", header: map[string]string{"X-RateLimit-Reset": inTwenty}, want: 20 * time.Second},
		{name: "RateLimit-Reset seconds", header: map[string]string{"RateLimit-Reset": "15"}, want: 15 * time.Second},
		{name: "Retry-After wins over Ratelimit-Reset", header: map[string]string{"Retry-After": "5", "Ratelimit-Reset": "50"}, want: 5 * time.Second},
		{name: "Ratelimit-Reset ignored on 502", status: http.StatusBadGateway, header: map[string]string{"Ratelimit-Reset": inTwenty}, want: time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := tt.status
			if status == 0 {
				status = http.StatusTooManyRequests
			}
			resp := &http.Response{StatusCode: status, Header: http.Header{}}
			for k, v := range tt.header {
				resp.Header.Set(k, v)
			}

			assert.Equal(t, tt.want, retryDelay(0, resp, now))
		})
	}
}

func TestRetryTransport_SkipsWaitBeyondMaxRetryWait(t *testing.T) {
	// Arrange — a server that throttles and asks for more than MaxRetryWait.
	calls := 0
	transport := &retryTransport{transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		calls++
		return &http.Response{
			StatusCode: http.StatusTooManyRequests,
			Header:     http.Header{"Retry-After": []string{strconv.Itoa(int((MaxRetryWait + time.Second).Seconds()))}},
			Body:       http.NoBody,
		}, nil
	})}

	req, err := http.NewRequest(http.MethodGet, "http://example.invalid", nil)
	require.NoError(t, err)

	// Act
	start := time.Now()
	resp, err := transport.RoundTrip(req)
	require.NoError(t, err)

	// Assert — the 429 comes back at once, without waiting or retrying.
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	assert.Equal(t, 1, calls)
	assert.Less(t, time.Since(start), time.Second)
}

func TestRetryTransport_HonoursRetryAfter(t *testing.T) {
	// Arrange — a server that throttles the first request and asks for no wait.
	srv, hits := newCountingServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	calls := 0
	transport := &retryTransport{transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			return &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Header:     http.Header{"Retry-After": []string{"0"}},
				Body:       http.NoBody,
			}, nil
		}
		return http.DefaultTransport.RoundTrip(req)
	})}

	req, err := http.NewRequest(http.MethodPost, srv.URL, strings.NewReader("{}"))
	require.NoError(t, err)

	// Act
	start := time.Now()
	resp, err := transport.RoundTrip(req)
	require.NoError(t, err)
	resp.Body.Close()

	// Assert — retried at once instead of after the one second backoff.
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int64(1), hits.Load())
	assert.Less(t, time.Since(start), time.Second)
}

func TestRetryTransport_StopsWaitingWhenContextIsDone(t *testing.T) {
	// Arrange — a server that always throttles and asks for a long wait.
	transport := &retryTransport{transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusTooManyRequests,
			Header:     http.Header{"Retry-After": []string{"60"}},
			Body:       http.NoBody,
		}, nil
	})}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://example.invalid", nil)
	require.NoError(t, err)

	// Act
	start := time.Now()
	_, err = transport.RoundTrip(req)

	// Assert — the wait ends once the context is done, not after a minute.
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Less(t, time.Since(start), time.Second)
}
