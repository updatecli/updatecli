package engine

import (
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/context"
	"github.com/updatecli/updatecli/pkg/core/reports"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// RunConditions run every conditions for a given configuration config.
func RunConditions(
	pipelineContext *context.Context,
	pipelineReport *reports.Report) (globalResult bool, err error) {

	logrus.Infof("\n\n%s:\n", strings.ToTitle("conditions"))
	logrus.Infof("%s\n\n", strings.Repeat("=", len("conditions")+1))

	// Sort conditions keys by building a dependency graph
	sortedConditionsKeys, err := SortedConditionsKeys(&pipelineContext.Conditions)
	if err != nil {
		return false, err
	}

	i := 0
	globalResult = true

	for _, id := range sortedConditionsKeys {
		condition := pipelineContext.Conditions[id]
		condition.Config = pipelineContext.Config.Conditions[id]

		rpt := pipelineReport.Conditions[i]

		rpt.Name = condition.Config.Name
		rpt.Result = result.FAILURE
		rpt.Kind = condition.Config.Kind

		ok, err := condition.Run(
			pipelineContext.Sources[condition.Config.SourceID].Config.Prefix +
				pipelineContext.Sources[condition.Config.SourceID].Output +
				pipelineContext.Sources[condition.Config.SourceID].Config.Postfix)

		if err != nil {
			globalResult = false
			pipelineContext.Conditions[id] = condition
			pipelineReport.Conditions[i] = rpt
			i++
			continue
		}

		if !ok {
			globalResult = false
			pipelineContext.Conditions[id] = condition
			pipelineReport.Conditions[i] = rpt
			i++
			continue
		}

		condition.Result = result.SUCCESS
		rpt.Result = result.SUCCESS

		pipelineContext.Conditions[id] = condition
		pipelineReport.Conditions[i] = rpt

		// Update pipeline after each condition run
		err = pipelineContext.Config.Update(pipelineContext)
		if err != nil {
			globalResult = false
			return globalResult, err
		}

		i++
	}

	return globalResult, nil
}
