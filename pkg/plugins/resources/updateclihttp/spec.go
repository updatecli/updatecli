package updateclihttp

/*
Spec defines a specification for a "http" resource
parsed from an updatecli manifest file.
*/
type Spec struct {
	/*
		[S][C] Specifies the URL of the HTTP request for this resource.
	*/
	Url string `yaml:",omitempty"`
	/*
		[S] Specifies the header to return as source value (instead of the body).
	*/
	ReturnResponseHeader string
	/*
		[S][C] Customizes the HTTP request to emit.
	*/
	Request Request
	/*
		[C] Specifies a set of custom assertions on the HTTP response for the condition.
	*/
	ResponseAsserts ResponseAsserts
}

type Request struct {
	/*
		[S][C] Specifies a custom HTTP request verb. Defaults to "GET".
	*/
	Verb string `yaml:",omitempty"`
	/*
		[S][C] Specifies a custom HTTP request body. Required with POST, PUT, PATCH.
	*/
	Body string `yaml:",omitempty"`
	/*
		[S][C] Specifies custom HTTP request headers. Defaults to an empty map.
	*/
	Headers map[string]string `yaml:",inline,omitempty"`
	/*
		[S][C] Specifies whether or not to follow redirects. Default to false (e.g. follow HTTP redirections) unless spec.returnresponseheader is set to true (source only).
	*/
	NoFollowRedirects bool `yaml:",omitempty"`
}

type ResponseAsserts struct {
	/*
		[C] Specifies a set of assertions on the HTTP response headers.
	*/
	Headers map[string]string `yaml:",inline,omitempty"`
	/*
		[C] Specifies a custom assertion on the HTTP response status code.
	*/
	StatusCode int `yaml:",omitempty"`
}
