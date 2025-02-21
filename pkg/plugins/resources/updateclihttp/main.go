package updateclihttp

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Http defines a resource of type "http"
type Http struct {
	spec       Spec
	httpClient httpclient.HTTPClient
	httpReq    *http.Request
}

/*
*
New returns a reference to a newly initialized Http resource
or an error if the provided Spec triggers a validation error.
*
*/
func New(spec interface{}) (*Http, error) {
	newSpec := Spec{}

	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return nil, err
	}

	if newSpec.Url == "" {
		return nil, fmt.Errorf("spec.url is not set but required.")
	}

	if newSpec.ResponseAsserts.StatusCode != 0 || len(newSpec.ResponseAsserts.Headers) > 0 {
		// This resource is a condition as the asserts are specified
		if newSpec.ReturnResponseHeader != "" {
			// Cannot be both source and condition
			return nil, fmt.Errorf("Cannot define both spec.responseasserts (source only) and spec.responseasserts (condition only).")
		}
	}

	httpClient := http.DefaultClient
	httpClient.Transport = httpclient.NewThrottledTransport(1*time.Second, 1, http.DefaultTransport)

	// Do not follow redirect as per https://pkg.go.dev/net/http when we want to get a header from original request
	httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if newSpec.ReturnResponseHeader != "" || newSpec.Request.NoFollowRedirects || newSpec.ResponseAsserts.StatusCode > 0 || len(newSpec.ResponseAsserts.Headers) > 0 {
			logrus.Debugf("spec.returnresponseheader defined: HTTP client won't follow redirects")
			return http.ErrUseLastResponse
		}

		return nil
	}

	httpVerb := http.MethodGet
	if newSpec.Request.Verb != "" {
		httpVerb = newSpec.Request.Verb
	}
	httpReq, err := http.NewRequest(httpVerb, newSpec.Url, nil)
	if err != nil {
		return nil, err
	}

	newResource := &Http{
		spec:       newSpec,
		httpClient: httpClient,
		httpReq:    httpReq,
	}

	return newResource, nil
}

// Changelog returns the changelog for this resource, or an empty string if not supported
func (h *Http) Changelog(from, to string) *result.Changelogs {
	return &result.Changelogs{
		{
			Title: "Changelog",
			Body:  h.spec.Url,
			URL:   h.spec.Url,
		},
	}
}

func (h *Http) performHttpRequest() (*http.Response, error) {
	logrus.Debugf("[http] Request to execute: %v", h.httpReq)
	httpRes, err := h.httpClient.Do(h.httpReq)
	if err != nil {
		return nil, err
	}

	logrus.Debugf("[http] Response received: %v", httpRes)

	return httpRes, nil
}
