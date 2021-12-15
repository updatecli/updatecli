package pipeline

import (
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

		if target.Config.Prefix == "" && p.Sources[target.Config.SourceID].Config.Prefix != "" {
			target.Config.Prefix = p.Sources[target.Config.SourceID].Config.Prefix
		}

		if target.Config.Postfix == "" && p.Sources[target.Config.SourceID].Config.Postfix != "" {
			target.Config.Postfix = p.Sources[target.Config.SourceID].Config.Postfix
		}

		err = target.Run(
			p.Sources[target.Config.SourceID].Output,
			&p.Options.Target)

		rpt.Result = target.Result

		if err != nil {
			logrus.Errorf("Something went wrong in target \"%v\" :\n", id)
			logrus.Errorf("%v\n\n", err)

			isResultIsFailed = true

			p.Targets[id] = target
			p.Report.Targets[id] = rpt
			i++
			continue
		}

		if strings.Compare(target.Result, result.ATTENTION) == 0 {
			isResultIsChanged = true
		}

		p.Targets[id] = target
		p.Report.Targets[id] = rpt

	}

	if isResultIsFailed {
		p.Report.Result = result.FAILURE
	} else if isResultIsChanged {
		p.Report.Result = result.ATTENTION
	} else {
		p.Report.Result = result.SUCCESS
	}

	return nil
}
