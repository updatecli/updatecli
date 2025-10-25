package engine

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// RunActions runs all actions defined in the configuration.
func (e *Engine) runActions() error {

	errs := []string{}

	logrus.Infof("\n\n%s\n", strings.ToTitle("Actions"))
	logrus.Infof("%s\n", strings.Repeat("=", len("Actions")+1))

	for id := range e.Pipelines {
		pipeline := e.Pipelines[id]
		if len(pipeline.Actions) > 0 {
			if err := pipeline.RunActions(); err != nil {
				errs = append(errs, err.Error())
				pipeline.Report.Result = result.FAILURE
				logrus.Errorf("action stage:\t%q", err.Error())
				continue
			}
		}
	}

	for id := range e.Pipelines {
		pipeline := e.Pipelines[id]
		if len(pipeline.Actions) > 0 {
			if err := pipeline.RunCleanActions(); err != nil {
				errs = append(errs, "cleaning: "+err.Error())
				pipeline.Report.Result = result.FAILURE
				logrus.Errorf("cleaning action stage:\t%q", err.Error())
				continue
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf(
			"errors occurred while running actions:\n\t* %s",
			strings.Join(errs, "\n\t* "))
	}

	return nil
}
