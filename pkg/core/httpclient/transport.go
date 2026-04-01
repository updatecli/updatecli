package httpclient

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// DefaultTransport returns a RoundTripper with retry and proxy support.
// Use when a third-party library needs a transport rather than a full http.Client.
func DefaultTransport() http.RoundTripper {
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
