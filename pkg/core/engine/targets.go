package engine

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/context"
	"github.com/updatecli/updatecli/pkg/core/engine/target"
	"github.com/updatecli/updatecli/pkg/core/reports"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// RunTargets iterates on every target to update each them.
func RunTargets(
	cfg *config.Config,
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
	sortedTargetsKeys, err := SortedTargetsKeys(&cfg.Targets)
	if err != nil {
		pipelineReport.Result = result.FAILURE
		return err
	}

	i := 0

	isResultIsChanged := false
	isResultIsFailed := false

	for _, id := range sortedTargetsKeys {
		target := cfg.Targets[id]
		ctx := pipelineContext.Targets[id]
		rpt := pipelineReport.Targets[i]

		rpt.Name = target.Name
		rpt.Result = result.FAILURE
		rpt.Kind = target.Kind

		targetChanged := false

		// Update pipeline before each target run
		err = cfg.Update(pipelineContext)
		if err != nil {
			return err
		}

		// Init target reporting
		target.Changelog = pipelineContext.Sources[target.SourceID].Changelog
		target.ReportBody = fmt.Sprintf("%s \n %s", sourceReport, conditionReport)
		target.ReportTitle = cfg.GetChangelogTitle(
			id,
			pipelineContext.Sources[target.SourceID].Output)

		if target.Prefix == "" && cfg.Sources[target.SourceID].Prefix != "" {
			target.Prefix = cfg.Sources[target.SourceID].Prefix
		}

		if target.Postfix == "" && cfg.Sources[target.SourceID].Postfix != "" {
			target.Postfix = cfg.Sources[target.SourceID].Postfix
		}

		targetChanged, err = target.Run(
			pipelineContext.Sources[target.SourceID].Output,
			options)

		if err != nil {
			logrus.Errorf("Something went wrong in target \"%v\" :\n", id)
			logrus.Errorf("%v\n\n", err)

			isResultIsFailed = true

			rpt.Result = result.FAILURE
			ctx.Result = result.FAILURE

			cfg.Targets[id] = target
			pipelineContext.Targets[id] = ctx
			pipelineReport.Targets[i] = rpt
			i++
			continue

		} else if targetChanged {
			isResultIsChanged = true

			ctx.Result = result.CHANGED
			rpt.Result = result.CHANGED

		} else {
			ctx.Result = result.SUCCESS
			rpt.Result = result.SUCCESS
		}

		cfg.Targets[id] = target
		pipelineContext.Targets[id] = ctx
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
