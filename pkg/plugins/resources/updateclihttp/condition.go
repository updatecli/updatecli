package updateclihttp

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

/*
Condition tests if the response of the specified HTTP request meets assertion.
If no assertion is specified, it only checks for successful HTTP response code (HTTP/1xx, HTTP/2xx or HTTP/3xx).
*/
func (h *Http) Condition(source string, scm scm.ScmHandler) (pass bool, message string, err error) {
	var failureMessages []string
	conditionResult := true

	httpRes, err := h.performHttpRequest()
	if err != nil {
		return false, "", err
	}
	defer httpRes.Body.Close()

	if h.spec.ResponseAsserts.StatusCode > 0 {
		if httpRes.StatusCode != h.spec.ResponseAsserts.StatusCode {
			failureMessages = append(failureMessages, fmt.Sprintf("Received status code %d while expecting %d.", httpRes.StatusCode, h.spec.ResponseAsserts.StatusCode))
			conditionResult = false
		}
	} else {
		// Only return an error if the HTTP status code is a server-side error: not found is not an error (but a condition failure)
		if httpRes.StatusCode >= http.StatusInternalServerError {
			return false, "", &ErrHttpError{resStatusCode: httpRes.StatusCode}
		}
		if httpRes.StatusCode >= http.StatusNotFound {
			failureMessages = append(failureMessages, fmt.Sprintf("Received status code %d which is >= 400, e.g. client or server HTTP error", httpRes.StatusCode))
			conditionResult = false
		}
	}

	if len(h.spec.ResponseAsserts.Headers) > 0 {
		for key, value := range h.spec.ResponseAsserts.Headers {
			foundValue := httpRes.Header.Get(key)

			if foundValue != value {
				failureMessages = append(failureMessages, fmt.Sprintf("Found value %q for header %q while expecting %q", foundValue, key, value))
				conditionResult = false
			}
		}
	}

	if !conditionResult {
		return false, fmt.Sprintf("[http] condition with URL: %q did NOT pass with the following errors: %s", h.spec.Url, strings.Join(failureMessages, "\n")), nil
	}

	return true, fmt.Sprintf("[http] condition with URL: %q passed", h.spec.Url), nil
}
