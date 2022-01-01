package httpclient

import "net/http"

// HTTPClient interface to define the contract of ALL http client to be used (http package or mocks)
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}
