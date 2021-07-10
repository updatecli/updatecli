package engine

import (
	"fmt"
	"os"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/mitchellh/hashstructure"
	"github.com/mitchellh/mapstructure"
	"github.com/olblak/updateCli/pkg/core/config"
	"github.com/olblak/updateCli/pkg/core/context"
	"github.com/olblak/updateCli/pkg/core/engine/target"
	"github.com/olblak/updateCli/pkg/core/reports"
	"github.com/olblak/updateCli/pkg/core/result"
	"github.com/olblak/updateCli/pkg/core/scm"
	"github.com/olblak/updateCli/pkg/core/tmp"
	"github.com/olblak/updateCli/pkg/plugins/github"

	"path/filepath"
	"strings"
)

// Engine defined parameters for a specific engine run.
type Engine struct {
	configurations []config.Config
	Contexts       []context.Context
	Options        Options
	Reports        reports.Reports
}

// Clean remove every traces from an updatecli run.
func (e *Engine) Clean() (err error) {
	err = tmp.Clean()
	return
}

// GetFiles return an array with every valid files.
func GetFiles(root string) (files []string) {
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logrus.Errorf("\n\u26A0 File %s: %s\n", path, err)
			os.Exit(1)
		}
		if info.Mode().IsRegular() {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		logrus.Errorf("err - %s", err)
	}

	return files
}

// InitSCM search and clone only once SCM configurations found.
func (e *Engine) InitSCM() (err error) {
	hashes := []uint64{}

	wg := sync.WaitGroup{}
	channel := make(chan int, 20)
	defer wg.Wait()

	for _, conf := range e.configurations {
		for _, source := range conf.Sources {

			if len(source.Scm) > 0 {
				err = Clone(&source.Scm, &hashes, channel, &wg)
				if err != nil {
					return err
				}
			}
		}
		for _, condition := range conf.Conditions {
			if len(condition.Scm) > 0 {

				err = Clone(&condition.Scm, &hashes, channel, &wg)
				if err != nil {
					return err
				}

			}
		}

		for _, target := range conf.Targets {
			if len(target.Scm) > 0 {

				err = Clone(&target.Scm, &hashes, channel, &wg)
				if err != nil {
					return err
				}
			}
		}
	}

	return err
}

// Clone parses a scm configuration then clone the git repository if needed.
func Clone(
	SCM *map[string]interface{},
	hashes *[]uint64,
	channel chan int,
	wg *sync.WaitGroup) error {

	hash, err := hashstructure.Hash(SCM, nil)
	if err != nil {
		return err
	}
	found := false

	for _, h := range *hashes {
		if h == hash {
			found = true
		}
	}

	if !found {
		s, _, err := scm.Unmarshal(*SCM)
		if err != nil {
			logrus.Errorf("err - %s", err)
		}
		*hashes = append(*hashes, hash)
		wg.Add(1)
		go func(s scm.Scm) {
			channel <- 1
			defer wg.Done()
			_, err := s.Clone()
			if err != nil {
				logrus.Errorf("err - %s", err)
			}
		}(s)
		<-channel

	}
	return nil
}

// Prepare run every actions needed before going further.
func (e *Engine) Prepare() (err error) {
	logrus.Infof("\n\n%s\n", strings.Repeat("+", len("Prepare")+4))
	logrus.Infof("+ %s +\n", strings.ToTitle("Prepare"))
	logrus.Infof("%s\n\n", strings.Repeat("+", len("Prepare")+4))

	err = tmp.Create()
	if err != nil {
		return err
	}

	err = e.ReadConfigurations()
	if err != nil {
		return err
	}

	err = e.InitSCM()
	if err != nil {
		return err
	}

	return err
}

// ReadConfigurations read every strategies configuration.
func (e *Engine) ReadConfigurations() error {
	// Read every strategy files
	for _, cfgFile := range GetFiles(e.Options.File) {

		c, err := config.New(cfgFile, e.Options.ValuesFiles, e.Options.SecretsFiles)

		if err != nil && err != config.ErrConfigFileTypeNotSupported {
			logrus.Errorf("%s\n\n", err)
			continue
		} else if err == config.ErrConfigFileTypeNotSupported {
			continue
		}

		e.configurations = append(e.configurations, c)
	}
	return nil

}

// Run run the full process one yaml file.
func (e *Engine) Run() (err error) {
	logrus.Infof("\n\n%s\n", strings.Repeat("+", len("Run")+4))
	logrus.Infof("+ %s +\n", strings.ToTitle("Run"))
	logrus.Infof("%s\n\n", strings.Repeat("+", len("Run")+4))

	for id, conf := range e.configurations {

		currentContext := context.Context{}
		currentContext.Init(&conf)

		currentReport := reports.Report{}
		currentReport.Init(&conf)

		if len(conf.Title) > 0 {
			logrus.Infof("\n\n%s\n", strings.Repeat("#", len(conf.Title)+4))
			logrus.Infof("# %s #\n", strings.ToTitle(conf.Title))
			logrus.Infof("%s\n\n", strings.Repeat("#", len(conf.Title)+4))

		} else {
			logrus.Infof("\n\n%s\n", strings.Repeat("#", len(conf.Name)+4))
			logrus.Infof("# %s #\n", strings.ToTitle(conf.Name))
			logrus.Infof("%s\n\n", strings.Repeat("#", len(conf.Name)+4))
		}

		err = RunSources(&conf, &currentReport, &currentContext)
		if err != nil {
			logrus.Errorf("Error occurred while running sources - %q", err.Error())
			e.Reports = append(e.Reports, currentReport)
			e.Contexts = append(e.Contexts, currentContext)
			continue
		}

		if len(conf.Conditions) > 0 {

			ok, err := RunConditions(
				&e.configurations[id],
				&currentContext,
				&currentReport)

			if err != nil {
				logrus.Infof("\n%s error happened during condition evaluation\n\n", result.FAILURE)
				e.Reports = append(e.Reports, currentReport)
				e.Contexts = append(e.Contexts, currentContext)
				continue
			} else if !ok {
				logrus.Infof("\n%s condition not met, skipping pipeline\n", result.FAILURE)
				e.Reports = append(e.Reports, currentReport)
				e.Contexts = append(e.Contexts, currentContext)
				continue
			}

		}

		if len(conf.Targets) > 0 {
			err := RunTargets(
				&e.configurations[id],
				&e.Options.Target,
				&currentReport,
				&currentContext)

			if err != nil {
				logrus.Errorf("%s %v\n", result.FAILURE, err)
				e.Reports = append(e.Reports, currentReport)
				e.Contexts = append(e.Contexts, currentContext)
				continue
			}
		}

		if err != nil {
			logrus.Errorf("\n%s %s \n\n", result.FAILURE, err)
		}

		e.Reports = append(e.Reports, currentReport)
		e.Contexts = append(e.Contexts, currentContext)
	}

	err = e.Reports.Show()
	if err != nil {
		return err
	}
	successCounter, changedCounter, failedCounter, err := e.Reports.Summary()
	if err != nil {
		return err
	}

	logrus.Infof("Run Summary")
	logrus.Infof("===========")
	logrus.Infof("%d job run", successCounter+changedCounter+failedCounter)
	logrus.Infof("%d job succeed", successCounter)
	logrus.Infof("%d job failed", failedCounter)
	logrus.Infof("%d job applied changes", changedCounter)

	logrus.Infof("")

	return err
}

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

// RunSources execute every updatecli sources for a specific pipeline
func RunSources(
	conf *config.Config,
	pipelineReport *reports.Report,
	pipelineContext *context.Context) error {

	sortedSourcesKeys, err := SortedSourcesKeys(&conf.Sources)
	if err != nil {
		logrus.Errorf("%s %v\n", result.FAILURE, err)
		return err
	}

	i := 0

	for _, id := range sortedSourcesKeys {
		source := conf.Sources[id]
		ctx := pipelineContext.Sources[id]
		rpt := pipelineReport.Sources[i]

		ctx.Result = result.FAILURE
		ctx.Output, ctx.Changelog, err = source.Execute()

		if err != nil {
			logrus.Errorf("%s %v\n", result.FAILURE, err)
			pipelineContext.Sources[id] = ctx
			pipelineReport.Sources[i] = rpt
			i++
			continue
		}

		if len(ctx.Output) == 0 {
			logrus.Infof("\n%s Something went wrong no value returned from Source", result.FAILURE)
			pipelineContext.Sources[id] = ctx
			pipelineReport.Sources[i] = rpt
			i++
			continue
		}

		ctx.Result = result.SUCCESS
		rpt.Result = result.SUCCESS

		pipelineContext.Sources[id] = ctx
		pipelineReport.Sources[i] = rpt

		err = conf.Update(pipelineContext)
		if err != nil {
			return err
		}

		i++
	}
	return err
}

// RunTargets iterate on every targets then call target on each of them.
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
		return err
	}

	i := 0

	for _, id := range sortedTargetsKeys {
		target := cfg.Targets[id]
		ctx := pipelineContext.Targets[id]
		rpt := pipelineReport.Targets[i]

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

			rpt.Result = result.FAILURE
			ctx.Result = result.FAILURE

			cfg.Targets[id] = target
			pipelineContext.Targets[id] = ctx
			pipelineReport.Targets[i] = rpt
			i++
			continue

		} else if targetChanged {
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
	return nil
}

// Show displays configurations that should be apply.
func (e *Engine) Show() error {

	err := e.ReadConfigurations()

	if err != nil {
		return err
	}

	for _, conf := range e.configurations {

		logrus.Infof("\n\n%s\n", strings.Repeat("#", len(conf.Name)+4))
		logrus.Infof("# %s #\n", strings.ToTitle(conf.Name))
		logrus.Infof("%s\n\n", strings.Repeat("#", len(conf.Name)+4))

		err = conf.Display()
		if err != nil {
			return err
		}

	}
	return nil
}
