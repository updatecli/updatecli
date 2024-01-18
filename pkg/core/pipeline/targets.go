package pipeline

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

var (
	// ErrRunTargets is return when at least one error happened during targets execution
	ErrRunTargets error = errors.New("something went wrong during target execution")
)

// RunTargets iterates on every target to update each of them.
func (p *Pipeline) RunTargets() error {
	logrus.Infof("\n\n%s\n", strings.ToTitle("Targets"))
	logrus.Infof("%s\n", strings.Repeat("=", len("Targets")+1))

	// Sort targets keys by building a dependency graph
	sortedTargetsKeys, err := SortedTargetsKeys(&p.Targets)
	if err != nil {
		p.Report.Result = result.FAILURE
		return err
	}

	isResultChanged := false
	p.Report.Result = result.SUCCESS

	errs := []error{}

	for _, id := range sortedTargetsKeys {
		// Update pipeline before each target run
		err = p.Update()
		if err != nil {
			return err
		}

		logrus.Infof("\n%s\n", id)
		logrus.Infof("%s\n", strings.Repeat("-", len(id)))

		target := p.Targets[id]
		target.Config = p.Config.Spec.Targets[id]

		// Ensure the result named contains the up to date target name after templating
		target.Result.Name = target.Config.ResourceConfig.Name
		target.Result.DryRun = target.DryRun

		shouldSkipTarget := false

		for _, parentTarget := range target.Config.DependsOn {
			if p.Targets[parentTarget].Result.Result == result.FAILURE {
				logrus.Warningf("Parent target[%q] did not succeed. Skipping execution of the target[%q]", parentTarget, id)
				shouldSkipTarget = true
				target.Result.Result = result.SKIPPED
			} else if target.Config.DependsOnChange && !p.Targets[parentTarget].Result.Changed {
				logrus.Infof("Parent target[%q] did not make a change. Skipping execution of the target[%q]", parentTarget, id)
				shouldSkipTarget = true
				target.Result.Result = result.SKIPPED
			}
		}

		// No need to run this target as one of its dependency failed
		if shouldSkipTarget {
			p.Targets[id] = target
			p.Report.Targets[id] = &target.Result
			continue
		}

		err = target.Run(p.Sources[target.Config.SourceID].Output, &p.Options.Target)

		if err != nil {
			p.Report.Result = result.FAILURE
			target.Result.Result = result.FAILURE

			errs = append(errs, fmt.Errorf("something went wrong in target %q : %q", id, err))
		}

		p.Targets[id] = target
		p.Report.Targets[id] = &target.Result

		if target.Result.Changed {
			isResultChanged = target.Result.Changed
		}
	}

	if len(errs) > 0 {
		logrus.Infof("\n")
		for _, e := range errs {
			logrus.Errorln(e)
		}
		logrus.Infof("\n")

		p.Report.Result = result.FAILURE
		return ErrRunTargets
	}

	if isResultChanged {
		p.Report.Result = result.ATTENTION
	}

	return nil
}
