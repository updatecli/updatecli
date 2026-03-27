package httpclient

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newCountingServer returns a test server that increments a hit counter on every
// request and responds with the provided handler. The counter pointer is safe to
// read after the server is closed.
func newCountingServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *atomic.Int64) {
	t.Helper()
	var hits atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		handler(w, r)
	}))
	t.Cleanup(srv.Close)
	return srv, &hits
}

func TestCachingTransport_CachesGetRequests(t *testing.T) {
	// Arrange
	srv, hits := newCountingServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello cache"))
	})

	ct := newCachingTransport(http.DefaultTransport)
	client := &http.Client{Transport: ct}

	// Act
	resp1, err := client.Get(srv.URL + "/resource")
	require.NoError(t, err)
	body1, err := io.ReadAll(resp1.Body)
	require.NoError(t, err)
	resp1.Body.Close()

	resp2, err := client.Get(srv.URL + "/resource")
	require.NoError(t, err)
	body2, err := io.ReadAll(resp2.Body)
	require.NoError(t, err)
	resp2.Body.Close()

	// Assert
	assert.Equal(t, int64(1), hits.Load(), "server should be hit exactly once")
	assert.Equal(t, "hello cache", string(body1))
	assert.Equal(t, "hello cache", string(body2))
	assert.Equal(t, 1, ct.Len())
}

func TestCachingTransport_SkipsNonGetMethods(t *testing.T) {
	// Arrange
	srv, hits := newCountingServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	ct := newCachingTransport(http.DefaultTransport)
	client := &http.Client{Transport: ct}

	// Act
	for i := 0; i < 2; i++ {
		resp, err := client.Post(srv.URL+"/submit", "application/json", nil)
		require.NoError(t, err)
		resp.Body.Close()
	}

	// Assert
	assert.Equal(t, int64(2), hits.Load(), "non-GET requests must never be served from cache")
	assert.Equal(t, 0, ct.Len())
}

func TestCachingTransport_SkipsNon2xxResponses(t *testing.T) {
	// Arrange
	srv, hits := newCountingServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	ct := newCachingTransport(http.DefaultTransport)
	client := &http.Client{Transport: ct}

	// Act
	for i := 0; i < 2; i++ {
		resp, err := client.Get(srv.URL + "/missing")
		require.NoError(t, err)
		resp.Body.Close()
	}

	// Assert
	assert.Equal(t, int64(2), hits.Load(), "non-2xx responses must not be cached")
	assert.Equal(t, 0, ct.Len())
}

func TestCachingTransport_DifferentURLsDifferentEntries(t *testing.T) {
	// Arrange
	srv, _ := newCountingServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, "body for %s", r.URL.Path)
	})

	ct := newCachingTransport(http.DefaultTransport)
	client := &http.Client{Transport: ct}

	// Act
	resp1, err := client.Get(srv.URL + "/alpha")
	require.NoError(t, err)
	body1, err := io.ReadAll(resp1.Body)
	require.NoError(t, err)
	resp1.Body.Close()

	resp2, err := client.Get(srv.URL + "/beta")
	require.NoError(t, err)
	body2, err := io.ReadAll(resp2.Body)
	require.NoError(t, err)
	resp2.Body.Close()

	// Assert
	assert.Equal(t, 2, ct.Len())
	assert.Equal(t, "body for /alpha", string(body1))
	assert.Equal(t, "body for /beta", string(body2))
}

func TestCachingTransport_CachedResponseHasCorrectHeaders(t *testing.T) {
	// Arrange
	srv, hits := newCountingServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Custom", "my-value")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	})

	ct := newCachingTransport(http.DefaultTransport)
	client := &http.Client{Transport: ct}

	// Act — first request populates the cache
	resp1, err := client.Get(srv.URL + "/data")
	require.NoError(t, err)
	io.ReadAll(resp1.Body) //nolint:errcheck
	resp1.Body.Close()

	// Act — second request is served from cache
	resp2, err := client.Get(srv.URL + "/data")
	require.NoError(t, err)
	io.ReadAll(resp2.Body) //nolint:errcheck
	resp2.Body.Close()

	// Assert
	assert.Equal(t, int64(1), hits.Load())
	assert.Equal(t, "application/json", resp2.Header.Get("Content-Type"))
	assert.Equal(t, "my-value", resp2.Header.Get("X-Custom"))
}

func TestCachingTransport_CachedBodyIsReReadable(t *testing.T) {
	// Arrange
	srv, hits := newCountingServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("rereadable body"))
	})

	ct := newCachingTransport(http.DefaultTransport)
	client := &http.Client{Transport: ct}

	var bodies [3]string

	// Act — three reads of the same URL
	for i := 0; i < 3; i++ {
		resp, err := client.Get(srv.URL + "/item")
		require.NoError(t, err)
		b, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		resp.Body.Close()
		bodies[i] = string(b)
	}

	// Assert
	assert.Equal(t, int64(1), hits.Load(), "server should be hit exactly once across three reads")
	assert.Equal(t, "rereadable body", bodies[0])
	assert.Equal(t, bodies[0], bodies[1])
	assert.Equal(t, bodies[1], bodies[2])
}

func TestEnableDisableHTTPCache(t *testing.T) {
	// Restore default state after the test regardless of outcome.
	t.Cleanup(DisableHTTPCache)

	// Arrange — a server that counts every request it receives.
	srv, hits := newCountingServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Act — client created while caching is enabled.
	EnableHTTPCache()
	cachedClient := NewRetryClient()

	// Hit the endpoint twice; the second call must be served from cache.
	for i := 0; i < 2; i++ {
		resp, err := cachedClient.Get(srv.URL + "/ping")
		require.NoError(t, err)
		io.ReadAll(resp.Body) //nolint:errcheck
		resp.Body.Close()
	}

	// Assert — only one real network request expected.
	assert.Equal(t, int64(1), hits.Load(), "caching client should hit the server exactly once")

	// Act — disable caching and create a new client; it must not cache.
	DisableHTTPCache()
	plainClient := NewRetryClient()

	for i := 0; i < 2; i++ {
		resp, err := plainClient.Get(srv.URL + "/ping")
		require.NoError(t, err)
		io.ReadAll(resp.Body) //nolint:errcheck
		resp.Body.Close()
	}

	// Assert — two additional hits means no caching.
	assert.Equal(t, int64(3), hits.Load(), "plain client should not use cache, expecting 3 total hits")
}

func TestCachingTransport_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	// Arrange
	srv, hits := newCountingServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("concurrent"))
	})

	ct := newCachingTransport(http.DefaultTransport)
	client := &http.Client{Transport: ct}

	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)

	// Act — all goroutines race to GET the same URL.
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			resp, err := client.Get(srv.URL + "/shared")
			if err != nil {
				return
			}
			io.ReadAll(resp.Body) //nolint:errcheck
			resp.Body.Close()
		}()
	}
	wg.Wait()

	// Assert — at most one real request (cache wins the race), never more than
	// the number of goroutines (safety upper bound), and no data race detected
	// by the Go race detector.
	serverHits := hits.Load()
	assert.GreaterOrEqual(t, serverHits, int64(1))
	assert.LessOrEqual(t, serverHits, int64(goroutines))
	assert.Equal(t, 1, ct.Len())
}
