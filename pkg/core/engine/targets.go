package engine

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/context"
	"github.com/updatecli/updatecli/pkg/core/engine/target"
	"github.com/updatecli/updatecli/pkg/core/reports"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// RunTargets iterates on every target to update each them.
func RunTargets(
	options *target.Options,
	pipelineReport *reports.Report,
	pipelineContext *context.Context) error {

	logrus.Infof("\n\n%s:\n", strings.ToTitle("Targets"))
	logrus.Infof("%s\n\n", strings.Repeat("=", len("Targets")+1))

	sourceReport, err := pipelineReport.String("sources")

	if err != nil {
		logrus.Errorf("err - %s", err)
	}
	conditionReport, err := pipelineReport.String("conditions")

	if err != nil {
		logrus.Errorf("err - %s", err)
	}

	// Sort targets keys by building a dependency graph
	sortedTargetsKeys, err := SortedTargetsKeys(&pipelineContext.Targets)
	if err != nil {
		pipelineReport.Result = result.FAILURE
		return err
	}

	i := 0

	isResultIsChanged := false
	isResultIsFailed := false

	for _, id := range sortedTargetsKeys {
		// Update pipeline before each target run
		err = pipelineContext.Config.Update(pipelineContext)
		if err != nil {
			return err
		}

		target := pipelineContext.Targets[id]
		target.Config = pipelineContext.Config.Targets[id]

		rpt := pipelineReport.Targets[i]

		rpt.Name = target.Config.Name
		rpt.Result = result.FAILURE
		rpt.Kind = target.Config.Kind

		targetChanged := false

		// Init target reporting
		target.Changelog = pipelineContext.Sources[target.Config.SourceID].Changelog
		target.ReportBody = fmt.Sprintf("%s \n %s", sourceReport, conditionReport)
		target.ReportTitle = pipelineContext.Config.GetChangelogTitle(
			id,
			pipelineContext.Sources[target.Config.SourceID].Result)

		if target.Config.Prefix == "" && pipelineContext.Sources[target.Config.SourceID].Config.Prefix != "" {
			target.Config.Prefix = pipelineContext.Sources[target.Config.SourceID].Config.Prefix
		}

		if target.Config.Postfix == "" && pipelineContext.Sources[target.Config.SourceID].Config.Postfix != "" {
			target.Config.Postfix = pipelineContext.Sources[target.Config.SourceID].Config.Postfix
		}

		targetChanged, err = target.Run(
			pipelineContext.Sources[target.Config.SourceID].Output,
			options)

		if err != nil {
			logrus.Errorf("Something went wrong in target \"%v\" :\n", id)
			logrus.Errorf("%v\n\n", err)

			isResultIsFailed = true

			rpt.Result = result.FAILURE
			target.Result = result.FAILURE

			pipelineContext.Targets[id] = target
			pipelineReport.Targets[i] = rpt
			i++
			continue

		} else if targetChanged {
			isResultIsChanged = true

			target.Result = result.CHANGED
			rpt.Result = result.CHANGED

		} else {
			target.Result = result.SUCCESS
			rpt.Result = result.SUCCESS
		}

		pipelineContext.Targets[id] = target
		pipelineReport.Targets[i] = rpt

		i++
	}

	if isResultIsFailed {
		pipelineReport.Result = result.FAILURE
	} else if isResultIsChanged {
		pipelineReport.Result = result.CHANGED
	} else {
		pipelineReport.Result = result.SUCCESS
	}

	return nil
}
