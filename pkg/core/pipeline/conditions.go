package pipeline

import (
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/reports"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// RunConditions run every conditions for a given configuration config.
func (p *Pipeline) RunConditions(
	pipelineReport *reports.Report) (globalResult bool, err error) {

	logrus.Infof("\n\n%s:\n", strings.ToTitle("conditions"))
	logrus.Infof("%s\n\n", strings.Repeat("=", len("conditions")+1))

	// Sort conditions keys by building a dependency graph
	sortedConditionsKeys, err := SortedConditionsKeys(&p.Conditions)
	if err != nil {
		return false, err
	}

	i := 0
	globalResult = true

	for _, id := range sortedConditionsKeys {
		condition := p.Conditions[id]
		condition.Config = p.Config.Conditions[id]

		rpt := pipelineReport.Conditions[i]

		rpt.Name = condition.Config.Name
		rpt.Result = result.FAILURE
		rpt.Kind = condition.Config.Kind

		ok, err := condition.Run(
			p.Sources[condition.Config.SourceID].Config.Prefix +
				p.Sources[condition.Config.SourceID].Output +
				p.Sources[condition.Config.SourceID].Config.Postfix)

		if err != nil {
			globalResult = false
			p.Conditions[id] = condition
			pipelineReport.Conditions[i] = rpt
			i++
			continue
		}

		if !ok {
			globalResult = false
			p.Conditions[id] = condition
			pipelineReport.Conditions[i] = rpt
			i++
			continue
		}

		condition.Result = result.SUCCESS
		rpt.Result = result.SUCCESS

		p.Conditions[id] = condition
		pipelineReport.Conditions[i] = rpt

		// Update pipeline after each condition run
		err = p.Config.Update(p)
		if err != nil {
			globalResult = false
			return globalResult, err
		}

		i++
	}

	return globalResult, nil
}
