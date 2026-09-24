package updateclihttp

/*
"http" defines the specification for sending an HTTP request and using its response.
It can be used as a "source" or a "condition".
*/
type Spec struct {
	// "url" defines the URL of the HTTP request.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// remark:
	//   * "url" is required.
	//
	// example:
	//   * url: https://www.updatecli.io
	//
	Url string `yaml:",omitempty"`
	// "returnresponseheader" defines the response header to return as the source value instead of the body.
	//
	// compatible:
	//   * source
	//
	// default:
	//   empty, so the source returns the response body.
	//
	// remark:
	//   * when set, the HTTP client does not follow redirects.
	//   * "returnresponseheader" and "responseasserts" are mutually exclusive.
	//
	// example:
	//   * returnresponseheader: Location
	//
	ReturnResponseHeader string
	// "request" defines the HTTP request to send.
	//
	// compatible:
	//   * source
	//   * condition
	//
	Request Request
	// "responseasserts" defines a set of assertions on the HTTP response.
	//
	// compatible:
	//   * condition
	//
	// default:
	//   empty, so the condition only fails when the response status code is 404 or higher.
	//
	// remark:
	//   * when set, the HTTP client does not follow redirects.
	//   * "returnresponseheader" and "responseasserts" are mutually exclusive.
	//
	ResponseAsserts ResponseAsserts
}

// Request defines the HTTP request sent by an "http" resource.
type Request struct {
	// "verb" defines the HTTP request method.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// default:
	//   GET
	//
	// example:
	//   * verb: POST
	//
	Verb string `yaml:",omitempty"`
	// "body" defines the HTTP request body.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// remark:
	//   * required with the methods POST, PUT and PATCH.
	//
	Body string `yaml:",omitempty"`
	// "headers" defines custom HTTP request headers.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// default:
	//   empty
	//
	// example:
	//   * headers:
	//       Accept: application/json
	//
	Headers map[string]string `yaml:",inline,omitempty"`
	// "nofollowredirects" defines whether the HTTP client stops following redirects.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// default:
	//   false, so redirects are followed.
	//
	// remark:
	//   * redirects are never followed when "returnresponseheader" or "responseasserts" is set.
	//
	NoFollowRedirects bool `yaml:",omitempty"`
}

// ResponseAsserts defines the assertions an "http" condition checks on the HTTP response.
type ResponseAsserts struct {
	// "headers" defines the expected values of HTTP response headers.
	//
	// compatible:
	//   * condition
	//
	// example:
	//   * headers:
	//       Content-Type: application/json
	//
	Headers map[string]string `yaml:",inline,omitempty"`
	// "statuscode" defines the expected HTTP response status code.
	//
	// compatible:
	//   * condition
	//
	// example:
	//   * statuscode: 200
	//
	StatusCode int `yaml:",omitempty"`
}
