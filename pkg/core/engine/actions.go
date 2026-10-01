package engine

import (
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// RunActions runs all actions defined in the configuration.
func (e *Engine) runActions(ctx context.Context) error {

	errs := []string{}

	logrus.Infof("\n\n%s\n", strings.ToTitle("Actions"))
	logrus.Infof("%s\n", strings.Repeat("=", len("Actions")+1))

	for id := range e.Pipelines {
		pipeline := e.Pipelines[id]
		if len(pipeline.Actions) > 0 {
			if err := pipeline.RunActions(ctx); err != nil {
				errs = append(errs, err.Error())
				pipeline.Report.Result = result.FAILURE
				logrus.Errorf("action stage:\t%q", err.Error())
				continue
			}
		}
	}

	logrus.Infof("Cleaning up actions published by previous executions")

	// Pipelines sharing a working branch share one pull request, so it's cleaned once,
	// and never when any of them published to it during this execution.
	handled := map[string]bool{}
	for _, pipeline := range e.Pipelines {
		for _, a := range pipeline.Actions {
			if key := a.CleanupKey(); a.Published && key != "" {
				handled[key] = true
			}
		}
	}

	for id := range e.Pipelines {
		pipeline := e.Pipelines[id]
		if len(pipeline.Actions) > 0 {
			if err := pipeline.RunCleanActions(ctx, handled); err != nil {
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
