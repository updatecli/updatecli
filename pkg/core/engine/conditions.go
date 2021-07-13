package engine

import (
	"strings"

	"github.com/olblak/updateCli/pkg/core/config"
	"github.com/olblak/updateCli/pkg/core/context"
	"github.com/olblak/updateCli/pkg/core/reports"
	"github.com/olblak/updateCli/pkg/core/result"
	"github.com/sirupsen/logrus"
)

// RunConditions run every conditions for a given configuration config.
func RunConditions(
	config *config.Config,
	pipelineContext *context.Context,
	pipelineReport *reports.Report) (globalResult bool, err error) {

	logrus.Infof("\n\n%s:\n", strings.ToTitle("conditions"))
	logrus.Infof("%s\n\n", strings.Repeat("=", len("conditions")+1))

	// Sort conditions keys by building a dependency graph
	sortedConditionsKeys, err := SortedConditionsKeys(&config.Conditions)
	if err != nil {
		return false, err
	}

	i := 0
	globalResult = true

	for _, id := range sortedConditionsKeys {
		condition := config.Conditions[id]
		ctx := pipelineContext.Conditions[id]
		rpt := pipelineReport.Conditions[i]

		rpt.Name = condition.Name
		rpt.Result = result.FAILURE
		rpt.Kind = condition.Kind

		ok, err := condition.Run(
			config.Sources[condition.SourceID].Prefix +
				pipelineContext.Sources[condition.SourceID].Output +
				config.Sources[condition.SourceID].Postfix)

		if err != nil {
			globalResult = false
			pipelineContext.Conditions[id] = ctx
			pipelineReport.Conditions[i] = rpt
			i++
			continue
		}

		if !ok {
			globalResult = false
			pipelineContext.Conditions[id] = ctx
			pipelineReport.Conditions[i] = rpt
			i++
			continue
		}

		ctx.Result = result.SUCCESS
		rpt.Result = result.SUCCESS

		pipelineContext.Conditions[id] = ctx
		pipelineReport.Conditions[i] = rpt

		// Update pipeline after each condition run
		err = config.Update(pipelineContext)
		if err != nil {
			globalResult = false
			return globalResult, err
		}
		i++
	}

	return globalResult, nil
}
