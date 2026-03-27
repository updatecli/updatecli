package httpclient

import (
	"bytes"
	"io"
	"net/http"
	"sync"

	"github.com/sirupsen/logrus"
)

// cachedResponse stores the essential parts of an HTTP response for replay.
type cachedResponse struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

// cachingTransport is an http.RoundTripper that caches successful GET responses
// in memory. Non-GET requests and non-2xx responses are passed through uncached.
type cachingTransport struct {
	transport http.RoundTripper
	mu        sync.RWMutex
	entries   map[string]cachedResponse
}

func newCachingTransport(transport http.RoundTripper) *cachingTransport {
	return &cachingTransport{
		transport: transport,
		entries:   make(map[string]cachedResponse),
	}
}

func (c *cachingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Method != http.MethodGet {
		return c.transport.RoundTrip(req)
	}

	key := req.URL.String()

	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()

	if ok {
		logrus.Debugf("http cache hit: %s", key)
		return &http.Response{
			StatusCode: entry.StatusCode,
			Header:     entry.Header.Clone(),
			Body:       io.NopCloser(bytes.NewReader(entry.Body)),
			Request:    req,
		}, nil
	}

	resp, err := c.transport.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	// Only cache 2xx responses so callers still see errors and redirects normally.
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			resp.Body.Close()
			return nil, readErr
		}
		resp.Body.Close()

		resp.Body = io.NopCloser(bytes.NewReader(body))

		c.mu.Lock()
		c.entries[key] = cachedResponse{
			StatusCode: resp.StatusCode,
			Header:     resp.Header.Clone(),
			Body:       body,
		}
		c.mu.Unlock()

		logrus.Debugf("http cache store: %s", key)
	}

	return resp, nil
}

// Len returns the number of cached responses (for testing).
func (c *cachingTransport) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}
