package updateclihttp

import (
	"fmt"
	"io"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source returns content from the response of the specified HTTP request (defaults to the body).
func (h *Http) Source(workingDir string, resultSource *result.Source) error {
	resultSource.Result = result.FAILURE

	httpRes, err := h.performHttpRequest()
	if err != nil {
		return err
	}
	defer httpRes.Body.Close()

	if httpRes.StatusCode >= http.StatusBadRequest {
		return &ErrHttpError{resStatusCode: httpRes.StatusCode}
	}

	resultSource.Result = result.SUCCESS
	resultSource.Description = fmt.Sprintf("[http] response received from %q.", h.spec.Url)

	if h.spec.ReturnResponseHeader != "" {
		resultSource.Information = httpRes.Header.Get(h.spec.ReturnResponseHeader)
		logrus.Debugf("[http] source: header %q found with %d characters", h.spec.ReturnResponseHeader, len(resultSource.Information))
	} else {
		b, err := io.ReadAll(httpRes.Body)
		if err != nil {
			return err
		}
		bodyContent := string(b)

		logrus.Debugf("[http] source: response body received (with %d characters)", len(bodyContent))

		resultSource.Information = bodyContent
	}

	return nil
}
