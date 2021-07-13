package engine

import (
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/olblak/updateCli/pkg/core/config"
	"github.com/olblak/updateCli/pkg/core/context"
	"github.com/olblak/updateCli/pkg/core/engine/target"
	"github.com/olblak/updateCli/pkg/core/reports"
	"github.com/olblak/updateCli/pkg/core/result"
	"github.com/olblak/updateCli/pkg/plugins/github"
	"github.com/sirupsen/logrus"
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

		target.Changelog = pipelineContext.Sources[target.SourceID].Changelog

		if _, ok := target.Scm["github"]; ok {
			var g github.Github

			err := mapstructure.Decode(target.Scm["github"], &g)

			if err != nil {
				continue
			}

			g.PullRequestDescription.Description = target.Changelog
			g.PullRequestDescription.Report = fmt.Sprintf("%s \n %s", sourceReport, conditionReport)

			if len(cfg.Title) > 0 {
				// If a pipeline title has been defined, then use it for pull request title
				g.PullRequestDescription.Title = fmt.Sprintf("[updatecli] %s",
					cfg.Title)

			} else if len(cfg.Targets) == 1 && len(target.Name) > 0 {
				// If we only have one target then we can use it as fallback.
				// Reminder, map in golang are not sorted so the order can't be kept between updatecli run
				g.PullRequestDescription.Title = fmt.Sprintf("[updatecli] %s", target.Name)
			} else {
				// At the moment, we don't have an easy way to describe what changed
				// I am still thinking to a better solution.
				logrus.Warning("**Fallback** Please add a title to you configuration using the field 'title: <your pipeline>'")
				g.PullRequestDescription.Title = fmt.Sprintf("[updatecli][%s] Bump version to %s",
					cfg.Sources[target.SourceID].Kind,
					pipelineContext.Sources[target.SourceID].Output)
			}

			target.Scm["github"] = g

		}

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
