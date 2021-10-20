package pipeline

import (
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// RunConditions run every conditions for a given configuration config.
func (p *Pipeline) RunConditions() (globalResult bool, err error) {

	logrus.Infof("\n\n%s:\n", strings.ToTitle("conditions"))
	logrus.Infof("%s\n", strings.Repeat("=", len("conditions")+1))

	// Sort conditions keys by building a dependency graph
	sortedConditionsKeys, err := SortedConditionsKeys(&p.Conditions)
	if err != nil {
		return false, err
	}

	globalResult = true

	for _, id := range sortedConditionsKeys {
		// Update pipeline before each condition run
		err = p.Config.Update(p)
		if err != nil {
			globalResult = false
		}

		condition := p.Conditions[id]
		condition.Config = p.Config.Conditions[id]

		rpt := p.Report.Conditions[id]

		logrus.Infof("\n%s\n", id)
		logrus.Infof("%s\n", strings.Repeat("-", len(id)))

		err := condition.Run(
			p.Sources[condition.Config.SourceID].Config.Prefix +
				p.Sources[condition.Config.SourceID].Output +
				p.Sources[condition.Config.SourceID].Config.Postfix)

		rpt.Result = condition.Result

		if err == nil {

			// Update pipeline after each condition run
			err = p.Config.Update(p)
			if err != nil {
				globalResult = false
				return globalResult, err
			}
		}

		if err != nil || strings.Compare(condition.Result, result.SUCCESS) != 0 {
			globalResult = false
		}

		p.Conditions[id] = condition
		p.Report.Conditions[id] = rpt

	}

	return globalResult, nil
}
