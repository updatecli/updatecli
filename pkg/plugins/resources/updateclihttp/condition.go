package updateclihttp

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

/*
Condition tests if the response of the specified HTTP request meets assertion.
If no assertion is specified, it only checks for successful HTTP response code (HTTP/1xx, HTTP/2xx or HTTP/3xx).
*/
func (h *Http) Condition(source string, scm scm.ScmHandler, resultCondition *result.Condition) error {
	resultCondition.Result = result.FAILURE
	resultCondition.Pass = false
	var failureMessages []string

	conditionResult := true

	httpRes, err := h.performHttpRequest()
	if err != nil {
		return err
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
			return &ErrHttpError{resStatusCode: httpRes.StatusCode}
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

	resultCondition.Pass = conditionResult
	if conditionResult {
		resultCondition.Result = result.SUCCESS
		resultCondition.Description = fmt.Sprintf("[http] condition with URL: %q passed", h.spec.Url)
	} else {
		resultCondition.Result = result.FAILURE
		resultCondition.Description = fmt.Sprintf("[http] condition with URL: %q did NOT pass with the following errors: %s", h.spec.Url, strings.Join(failureMessages, "\n"))
	}

	return nil
}
