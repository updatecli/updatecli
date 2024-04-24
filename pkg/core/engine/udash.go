package engine

import (
	"errors"
	"fmt"
	"strings"

	"github.com/updatecli/updatecli/pkg/core/cmdoptions"
	"github.com/updatecli/updatecli/pkg/core/udash"
)

// publishToUdash publish pipeline reports to the Udash service.
// This service is still experimental and should be used with caution.
// More information on https://github.com/updatecli/udash
func (e *Engine) publishToUdash() error {

	errs := []string{}

	if !cmdoptions.Experimental {
		return nil
	}

	for id := range e.Pipelines {
		pipeline := e.Pipelines[id]
		if err := udash.Publish(&pipeline.Report); err != nil &&
			!errors.Is(err, udash.ErrNoUdashBearerToken) &&
			!errors.Is(err, udash.ErrNoUdashAPIURL) {
			errs = append(errs, pipeline.Name+err.Error())
		}
		e.Pipelines[id] = pipeline
	}

	if len(errs) > 0 {
		return fmt.Errorf(
			"errors occurred while publishing to Udash:\n\t* %s",
			strings.Join(errs, "\n\t* "),
		)
	}

	return nil
}
