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

	for _, targetId := range sortedTargetsKeys {
		// Update pipeline before each target run
		err = p.Config.Update(p)
		if err != nil {
			return err
		}

		logrus.Infof("\n%s\n", targetId)
		logrus.Infof("%s\n", strings.Repeat("-", len(targetId)))

		target := p.Targets[targetId]
		target.Config = p.Config.Spec.Targets[targetId]

		report := p.Report.Targets[targetId]
		// Update report name as the target configuration might has been updated (templated values)
		report.Name = target.Config.Name

		shouldSkipTarget := false

		for _, parentTarget := range target.Config.DependsOn {
			if p.Targets[parentTarget].Result != result.SUCCESS {
				logrus.Warningf("Parent target[%q] did not succeed. Skipping execution of the target[%q]", parentTarget, targetId)
				shouldSkipTarget = true
				target.Result = result.SKIPPED
				report.Result = target.Result
			}
		}

		// No need to run this target as one of its dependency failed
		if shouldSkipTarget {
			p.Targets[targetId] = target
			p.Report.Targets[targetId] = report
			continue
		}

		err = target.Run(p.Sources[target.Config.SourceID].Output)
		// TODO: retrieve target diff/changed files
		if err != nil {
			p.Report.Result = result.FAILURE
			target.Result = result.FAILURE

			errs = append(errs, fmt.Errorf("something went wrong in target %q : %q", targetId, err))
		}

		report.Result = target.Result

		p.Targets[targetId] = target
		p.Report.Targets[targetId] = report

		if strings.Compare(target.Result, result.ATTENTION) == 0 {
			isResultChanged = true
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
