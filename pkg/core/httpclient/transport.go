package httpclient

import (
	"net/http"
	"sync"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var (
	cacheMu     sync.Mutex
	activeCache *cachingTransport
)

// EnableHTTPCache activates in-memory caching of GET responses for all
// HTTP clients created via this package after this call.
// Call once before pipeline execution begins.
func EnableHTTPCache() {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	activeCache = newCachingTransport(otelhttp.NewTransport(&http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}))
}

// DisableHTTPCache deactivates the HTTP cache and releases cached data.
// Safe to call even if caching was never enabled.
func DisableHTTPCache() {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	activeCache = nil
}

// DefaultTransport returns a RoundTripper with retry and proxy support.
// When HTTP caching is enabled via EnableHTTPCache, GET responses are served
// from cache: retryTransport -> cachingTransport -> otelhttp -> http.Transport.
func DefaultTransport() http.RoundTripper {
	cacheMu.Lock()
	inner := activeCache
	cacheMu.Unlock()

	if inner != nil {
		return &retryTransport{transport: inner}
	}

	return &retryTransport{
		transport: otelhttp.NewTransport(&http.Transport{
			Proxy: http.ProxyFromEnvironment,
		}),
	}
}

// ProxyOnlyTransport returns a RoundTripper with proxy support but no retry.
// Use when the caller handles its own retry (e.g. go-containerregistry) or
// for non-idempotent operations (e.g. OAuth token exchange) where retry is unsafe.
func ProxyOnlyTransport() http.RoundTripper {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}
}

// NewPlainClient returns an HTTP client with proxy support but no retry.
func NewPlainClient() *http.Client {
	return &http.Client{Transport: ProxyOnlyTransport()}
}
