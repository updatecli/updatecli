package temurin

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

/*
Condition tests if the response of the specified HTTP request meets assertion.
If no assertion is specified, it only checks for successful HTTP response code (HTTP/1xx, HTTP/2xx or HTTP/3xx).
*/
func (t *Temurin) Condition(source string, scm scm.ScmHandler) (pass bool, message string, err error) {
	message = "Condition not supported for the resources of kind 'temurin'"
	return false, message, fmt.Errorf(message)
}
