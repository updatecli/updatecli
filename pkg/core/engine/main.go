package engine

import (
	"fmt"
	"os"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/mitchellh/hashstructure"
	"github.com/mitchellh/mapstructure"
	"github.com/olblak/updateCli/pkg/core/config"
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
		if len(conf.Source.Scm) > 0 {
			err = Clone(&conf.Source.Scm, &hashes, channel, &wg)
			if err != nil {
				return err
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

		c := config.Config{}

		_, basename := filepath.Split(cfgFile)
		cfgFileName := strings.TrimSuffix(basename, filepath.Ext(basename))

		c.Name = strings.ToTitle(cfgFileName)

		err := c.ReadFile(cfgFile, e.Options.ValuesFile)
		if err != nil {
			logrus.Errorf("%s - %s\n\n", basename, err)
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

	for _, conf := range e.configurations {
		if len(conf.Title) > 0 {
			logrus.Infof("\n\n%s\n", strings.Repeat("#", len(conf.Title)+4))
			logrus.Infof("# %s #\n", strings.ToTitle(conf.Title))
			logrus.Infof("%s\n\n", strings.Repeat("#", len(conf.Title)+4))

		} else {
			logrus.Infof("\n\n%s\n", strings.Repeat("#", len(conf.Name)+4))
			logrus.Infof("# %s #\n", strings.ToTitle(conf.Name))
			logrus.Infof("%s\n\n", strings.Repeat("#", len(conf.Name)+4))
		}

		conditionsStageReport := []reports.Stage{}
		targetsStageReport := []reports.Stage{}

		for _, c := range conf.Conditions {
			s := reports.Stage{
				Name:   c.Name,
				Kind:   c.Kind,
				Result: result.FAILURE,
			}
			conditionsStageReport = append(conditionsStageReport, s)
		}

		for _, t := range conf.Targets {
			s := reports.Stage{
				Name:   t.Name,
				Kind:   t.Kind,
				Result: result.FAILURE,
			}
			targetsStageReport = append(targetsStageReport, s)
		}

		report := reports.Init(
			conf.Name,
			reports.Stage{
				Name:   conf.Source.Name,
				Kind:   conf.Source.Kind,
				Result: result.FAILURE,
			},
			conditionsStageReport,
			targetsStageReport,
		)

		report.Name = strings.ToTitle(conf.Name)

		err = conf.Source.Execute()

		if err != nil {
			logrus.Errorf("%s %v\n", result.FAILURE, err)
			e.Reports = append(e.Reports, report)
			continue
		}

		if conf.Source.Output == "" {
			conf.Source.Result = result.FAILURE
			report.Source.Result = result.FAILURE
			logrus.Infof("\n%s Something went wrong no value returned from Source", result.FAILURE)
			e.Reports = append(e.Reports, report)
			continue
		}

		conf.Source.Result = result.SUCCESS
		report.Source.Result = result.SUCCESS

		if len(conf.Conditions) > 0 {
			c := conf
			ok, err := RunConditions(&c)

			i := 0

			for _, c := range conf.Conditions {
				conditionsStageReport[i].Result = c.Result
				report.Conditions[i].Result = c.Result
				i++
			}

			if err != nil || !ok {
				logrus.Infof("%s %v\n", result.FAILURE, err)
				e.Reports = append(e.Reports, report)
				continue
			}
		}

		if len(conf.Targets) > 0 {
			c := conf
			changed, err := RunTargets(&c, &e.Options.Target, &report)
			if err != nil {
				logrus.Errorf("%s %v\n", result.FAILURE, err)
				e.Reports = append(e.Reports, report)
				continue
			}
			if changed {
				report.Result = result.CHANGED
			} else {
				report.Result = result.SUCCESS
			}

			i := 0
			for _, t := range conf.Targets {
				targetsStageReport[i].Result = t.Result
				report.Targets[i].Result = t.Result
				i++
			}
		}

		if err != nil {
			logrus.Errorf("\n%s %s \n\n", result.FAILURE, err)
		}

		e.Reports = append(e.Reports, report)
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
func RunConditions(conf *config.Config) (bool, error) {
	logrus.Infof("\n\n%s:\n", strings.ToTitle("conditions"))
	logrus.Infof("%s\n\n", strings.Repeat("=", len("conditions")+1))

	for k, c := range conf.Conditions {
		c.Result = result.FAILURE

		conf.Conditions[k] = c
		ok, err := c.Run(conf.Source.Prefix + conf.Source.Output + conf.Source.Postfix)
		if err != nil {
			return false, err
		}

		if !ok {
			c.Result = result.FAILURE
			conf.Conditions[k] = c
			logrus.Infof("\n%s skipping: condition not met\n", result.FAILURE)
			return false, nil
		}

		c.Result = result.SUCCESS
		conf.Conditions[k] = c
	}

	return true, nil
}

// RunTargets iterate on every targets then call target on each of them.
func RunTargets(config *config.Config, options *target.Options, report *reports.Report) (targetsChanged bool, err error) {
	targetsChanged = false

	logrus.Infof("\n\n%s:\n", strings.ToTitle("Targets"))
	logrus.Infof("%s\n\n", strings.Repeat("=", len("Targets")+1))

	sourceReport, err := report.String("source")

	if err != nil {
		logrus.Errorf("err - %s", err)
	}
	conditionReport, err := report.String("conditions")

	if err != nil {
		logrus.Errorf("err - %s", err)
	}

	for id, t := range config.Targets {
		targetChanged := false

		t.Changelog = config.Source.Changelog

		if _, ok := t.Scm["github"]; ok {
			var g github.Github

			err := mapstructure.Decode(t.Scm["github"], &g)

			if err != nil {
				return false, err
			}

			g.PullRequestDescription.Description = t.Changelog
			g.PullRequestDescription.Report = fmt.Sprintf("%s \n %s", sourceReport, conditionReport)

			if len(config.Title) > 0 {
				// If a pipeline title has been defined, then use it for pull request title
				g.PullRequestDescription.Title = fmt.Sprintf("[updatecli] %s",
					config.Title)

			} else if len(config.Targets) == 1 && len(t.Name) > 0 {
				// If we only have one target then we can use it as fallback.
				// Reminder, map in golang are not sorted so the order can't be kept between updatecli run
				g.PullRequestDescription.Title = fmt.Sprintf("[updatecli] %s", t.Name)
			} else {
				// At the moment, we don't have an easy way to describe what changed
				// I am still thinking to a better solution.
				logrus.Warning("**Fallback** Please add a title to you configuration using the field 'title: <your pipeline>'")
				g.PullRequestDescription.Title = fmt.Sprintf("[updatecli][%s] Bump version to %s",
					config.Source.Kind,
					config.Source.Output)
			}

			t.Scm["github"] = g

		}

		if t.Prefix == "" && config.Source.Prefix != "" {
			t.Prefix = config.Source.Prefix
		}

		if t.Postfix == "" && config.Source.Postfix != "" {
			t.Postfix = config.Source.Postfix
		}

		targetChanged, err = t.Run(config.Source.Output, options)

		if err != nil {
			logrus.Errorf("Something went wrong in target \"%v\" :\n", id)
			logrus.Errorf("%v\n\n", err)
			t.Result = result.FAILURE
			return targetChanged, err
		} else if targetChanged {
			t.Result = result.CHANGED
			targetsChanged = true
		} else {
			t.Result = result.SUCCESS
		}

		config.Targets[id] = t
	}
	return targetsChanged, nil
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
