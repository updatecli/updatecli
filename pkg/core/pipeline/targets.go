package pipeline

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
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

	for _, id := range sortedTargetsKeys {
		// Update pipeline before each target run
		err = p.Config.Update(p)
		if err != nil {
			return err
		}

		logrus.Infof("\n%s\n", id)
		logrus.Infof("%s\n", strings.Repeat("-", len(id)))

		target := p.Targets[id]
		target.Config = p.Config.Targets[id]

		report := p.Report.Targets[id]
		// Update report name as the target configuration might has been updated (templated values)
		report.Name = target.Config.Name

		err = target.Run(p.Sources[target.Config.SourceID].Output, &p.Options.Target)
		if err != nil {
			p.Report.Result = result.FAILURE
			p.Targets[id] = target
			return fmt.Errorf("something went wrong in target \"%v\" :\n%v\n\n", id, err)
		}

		report.Result = target.Result
		p.Targets[id] = target
		p.Report.Targets[id] = report

		if strings.Compare(target.Result, result.ATTENTION) == 0 {
			isResultChanged = true
		}
	}

	if isResultChanged {
		p.Report.Result = result.ATTENTION
	}

	return nil
}
