package pipeline

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// RunTargets iterates on every target to update each of them.
func (p *Pipeline) RunTargets() error {
	var errorMessages strings.Builder
	logrus.Infof("\n\n%s\n", strings.ToTitle("Targets"))
	logrus.Infof("%s\n", strings.Repeat("=", len("Targets")+1))

	// Sort targets keys by building a dependency graph
	sortedTargetsKeys, err := SortedTargetsKeys(&p.Targets)
	if err != nil {
		p.Report.Result = result.FAILURE
		return err
	}

	i := 0

	isResultIsChanged := false
	isResultIsFailed := false

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

		rpt := p.Report.Targets[id]
		// Update report name as the target configuration might has been updated (templated values)
		rpt.Name = target.Config.Name

		err = target.Run(p.Sources[target.Config.SourceID].Output, &p.Options.Target)

		rpt.Result = target.Result

		p.Targets[id] = target
		p.Report.Targets[id] = rpt

		if err != nil {
			errorMessages.WriteString(fmt.Sprintf("Something went wrong in target \"%v\" :\n", id))
			errorMessages.WriteString(fmt.Sprintf("%v\n\n", err))

			isResultIsFailed = true

			i++
			continue
		}

		if strings.Compare(target.Result, result.ATTENTION) == 0 {
			isResultIsChanged = true
		}
	}

	if isResultIsFailed {
		p.Report.Result = result.FAILURE
		return errors.New(errorMessages.String())
	} else if isResultIsChanged {
		p.Report.Result = result.ATTENTION
	} else {
		p.Report.Result = result.SUCCESS
	}

	return nil
}
