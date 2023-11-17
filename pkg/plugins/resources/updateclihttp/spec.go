package updateclihttp

// Spec defines a specification for a "http" resource
// parsed from an updatecli manifest file
type Spec struct {
	// [S][C] Specifies the URL to use for the HTTP request of this resource
	Url string `yaml:",omitempty"`
	// [S] Specifies the header to return value instead of body
	ReturnResponseHeader string
	// [S][C] Specifies custom HTTP request parameters
	Request Request
	// [C] Specifies assertions on the HTTP response
	ResponseAsserts ResponseAsserts
}

type Request struct {
	// [S][C] Specifies the HTTP verb for the request to be used. Defaults to "GET".
	Verb string `yaml:",omitempty"`
	// [S][C] Specifies the HTTP body for the request to be used. Defaults to "" (empty string).
	Body string `yaml:",omitempty"`
	// [S][C] Specifies the HTTP headers for the request to be used. Defaults to an empty map.
	Headers map[string]string `yaml:",inline,omitempty"`
	// [S][C] Specifies whether or not to follow redirect. Default to false (e.g. follow), unless spec.returnresponseheader is true
	NoFollowRedirects bool `yaml:",omitempty"`
}

type ResponseAsserts struct {
	// [C] Specifies a set of assertions on the HTTP response headers.
	Headers map[string]string `yaml:",inline,omitempty"`
	// [C] Specifies an assertion on the HTTP response status code.
	StatusCode int `yaml:",omitempty"`
}
