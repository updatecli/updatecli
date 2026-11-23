package engine

import (
	"context"
	"fmt"
	"strings"

	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/core/pipeline"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// RunActions runs all actions defined in the configuration.
func (e *Engine) runActions(ctx context.Context) error {

	errs := []string{}

	log.Title("Actions")

	for id := range e.Pipelines {
		pipeline := e.Pipelines[id]
		// A pipeline rejected because its scm failed to clone, or waiting for an earlier version filter,
		// never ran, so it must not publish or merge anything.
		if len(pipeline.Actions) > 0 && !e.skipsActions(pipeline) {
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

	handled := e.handledCleanupKeys()

	for id := range e.Pipelines {
		pipeline := e.Pipelines[id]
		if len(pipeline.Actions) > 0 && !e.skipsActions(pipeline) {
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

/*
handledCleanupKeys returns the cleanup keys of the actions which must not be cleaned up.

Pipelines sharing a working branch share one pull request, so it's cleaned once,
and never when any of them published to it during this execution.
It's also kept when one of them was rejected because its scm failed to clone,
or waited for an earlier version filter, as Updatecli doesn't know what that pipeline would have published.
*/
func (e *Engine) handledCleanupKeys() map[string]bool {
	handled := map[string]bool{}
	for _, pipeline := range e.Pipelines {
		rejected := e.skipsActions(pipeline)
		for _, a := range pipeline.Actions {
			if key := a.CleanupKey(); key != "" && (a.Published || rejected) {
				handled[key] = true
			}
		}
	}
	return handled
}

// skipsActions reports whether a pipeline never ran during this execution,
// because its scm failed to clone or it waited for an earlier version filter,
// so its actions must be neither run nor cleaned up.
func (e *Engine) skipsActions(p *pipeline.Pipeline) bool {
	return p.IsWaiting() || e.failedSCMError(p) != nil
}
