package engine

import (
	"context"
	"fmt"
	"strings"

	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// RunActions runs all actions defined in the configuration.
func (e *Engine) runActions(ctx context.Context) error {

	errs := []string{}

	log.Heading("Actions")

	for id := range e.Pipelines {
		pipeline := e.Pipelines[id]
		if len(pipeline.Actions) > 0 {
			stopCapture := pipeline.CaptureConsoleOutput()
			if err := pipeline.RunActions(ctx); err != nil {
				errs = append(errs, err.Error())
				pipeline.Report.Result = result.FAILURE
				log.Errorf("action stage:\t%q", err.Error())
			}
			stopCapture()
		}
	}

	log.Infof("Cleaning up actions published by previous executions")

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
			stopCapture := pipeline.CaptureConsoleOutput()
			if err := pipeline.RunCleanActions(ctx, handled); err != nil {
				errs = append(errs, "cleaning: "+err.Error())
				pipeline.Report.Result = result.FAILURE
				log.Errorf("cleaning action stage:\t%q", err.Error())
			}
			stopCapture()
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf(
			"errors occurred while running actions:\n\t* %s",
			strings.Join(errs, "\n\t* "))
	}

	return nil
}
