package engine

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// RunActions runs all actions defined in the configuration.
// To avoid the situation where a pipeline close a pullrequest even thought the next pipeline need
// to update the same pullrequest, we need to run pipeline actions once all pipelines' targets have been executed.
// The goal is to respect an order where we first handle pipelines in ATTENTION state, then FAILURE, then SUCCESS and finally SKIPPED.
// cfr https://github.com/updatecli/updatecli/issues/2039

// 1. ATTENTION: to update existing pull request
// 2. FAILURE: which may clean up existing pull request
// 3. SUCCESS: which may clean up existing pull request
// 4. SKIPPED: which may clean up existing pull request
//
// It worth reminding that a pipeline can contain multiple actions.
// If at least one action is in an attention state,
// then the pipeline is considered in attention state as well.
// So the same logic must be applied differently based on the different actions state.
func (e *Engine) runActions() error {

	errs := []string{}

	logrus.Infof("\n\n%s\n", strings.ToTitle("Actions"))
	logrus.Infof("%s\n\n", strings.Repeat("=", len("Actions")+1))

	for _, pipelineState := range []string{result.ATTENTION, result.FAILURE, result.SUCCESS, result.SKIPPED} {
		for id := range e.Pipelines {
			pipeline := e.Pipelines[id]
			if len(pipeline.Actions) > 0 {
				if err := pipeline.RunActions(pipelineState); err != nil {
					errs = append(errs, err.Error())
					pipeline.Report.Result = result.FAILURE
					logrus.Errorf("action stage:\t%q", err.Error())
					continue
				}
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
