package engine

import (
	"fmt"
	"os"

	"github.com/mitchellh/hashstructure"
	"github.com/mitchellh/mapstructure"
	"github.com/olblak/updateCli/pkg/config"
	"github.com/olblak/updateCli/pkg/engine/target"
	"github.com/olblak/updateCli/pkg/github"
	"github.com/olblak/updateCli/pkg/reports"
	"github.com/olblak/updateCli/pkg/result"
	"github.com/olblak/updateCli/pkg/scm"
	"github.com/olblak/updateCli/pkg/tmp"

	"path/filepath"
	"strings"
)

var engine Engine

// Engine defined parameters for a specific engine run
type Engine struct {
	configurations []config.Config
	Options        Options
	Reports        reports.Reports
}

// Clean remove every traces from an updatecli run
func (e *Engine) Clean() (err error) {
	tmp.Clean()
	return err
}

// GetFiles return an array with every valid files
func GetFiles(root string) (files []string) {

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("\n\u26A0 File %s: %s\n", path, err)
			os.Exit(1)
		}
		if info.Mode().IsRegular() {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		fmt.Println(err)
	}

	return files
}

// InitSCM search and clone only once SCM configurations found

func (e *Engine) InitSCM() (err error) {

	hashes := []uint64{}

	for _, conf := range e.configurations {
		for _, condition := range conf.Conditions {
			if len(condition.Scm) > 0 {
				hash, err := hashstructure.Hash(condition.Scm, nil)
				if err != nil {
					fmt.Println(hash)
				}
				found := false

				for _, h := range hashes {
					if h == hash {
						found = true
					}
				}

				if !found {
					s, err := scm.Unmarshal(condition.Scm)

					if err != nil {
						fmt.Println(err)
					}
					hashes = append(hashes, hash)
					s.Clone()
				}
			}
		}
		for _, target := range conf.Targets {
			if len(target.Scm) > 0 {
				hash, err := hashstructure.Hash(target.Scm, nil)
				if err != nil {
					fmt.Println(hash)
				}
				found := false

				for _, h := range hashes {
					if h == hash {
						found = true
					}
				}

				if !found {
					s, err := scm.Unmarshal(target.Scm)

					if err != nil {
					}
					fmt.Println(err)
					hashes = append(hashes, hash)
					s.Clone()
				}
			}
		}
	}

	return err
}

// Prepare run every actions needed before going further
func (e *Engine) Prepare() (err error) {
	err = tmp.Create()
	if err != nil {
		fmt.Printf("\n\u26A0 %s\n", err)
		os.Exit(1)
	}

	err = e.ReadConfigurations()

	if err != nil {
		fmt.Printf("\n\u26A0 %s\n", err)
	}

	err = e.InitSCM()

	if err != nil {
		fmt.Printf("\n\u26A0 %s\n", err)
	}

	return err
}

// ReadConfigurations read every strategies configuration
func (e *Engine) ReadConfigurations() error {
	// Read every strategy files
	for _, cfgFile := range GetFiles(e.Options.File) {

		c := config.Config{}

		_, basename := filepath.Split(cfgFile)
		cfgFileName := strings.TrimSuffix(basename, filepath.Ext(basename))

		// fmt.Printf("\n\n%s\n", strings.Repeat("#", len(cfgFileName)+4))
		// fmt.Printf("# %s #\n", strings.ToTitle(cfgFileName))
		// fmt.Printf("%s\n\n", strings.Repeat("#", len(cfgFileName)+4))

		c.Name = strings.ToTitle(cfgFileName)

		err := c.ReadFile(cfgFile, e.Options.ValuesFile)
		if err != nil {
			fmt.Println(err)
			continue
		}
		e.configurations = append(e.configurations, c)
	}
	return nil

}

// Run run the full process one yaml file
func (e *Engine) Run() (err error) {

	for _, conf := range e.configurations {

		fmt.Printf("\n\n%s\n", strings.Repeat("#", len(conf.Name)+4))
		fmt.Printf("# %s #\n", strings.ToTitle(conf.Name))
		fmt.Printf("%s\n\n", strings.Repeat("#", len(conf.Name)+4))

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
			e.Reports = append(e.Reports, report)
			continue
		}

		if conf.Source.Output == "" {
			conf.Source.Result = result.FAILURE
			report.Source.Result = result.FAILURE
			fmt.Printf("\n%s Something went wrong no value returned from Source", result.FAILURE)
			e.Reports = append(e.Reports, report)
			continue
		}
		conf.Source.Result = result.SUCCESS
		report.Source.Result = result.SUCCESS

		if len(conf.Conditions) > 0 {
			ok, err := RunConditions(&conf)

			i := 0
			for _, c := range conf.Conditions {
				conditionsStageReport[i].Result = c.Result
				report.Conditions[i].Result = c.Result
				i++
			}

			if err != nil || !ok {
				e.Reports = append(e.Reports, report)
				continue
			}

		}

		if len(conf.Targets) > 0 {
			changed, err := RunTargets(&conf, &e.Options.Target, &report)
			if err != nil {
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
			fmt.Printf("\n%s %s \n\n", result.FAILURE, err)
		}
		e.Reports = append(e.Reports, report)
	}

	e.Reports.Show()
	e.Reports.Summary()
	fmt.Printf("\n")

	return err
}

// RunConditions run every conditions for a given configuration config
func RunConditions(conf *config.Config) (bool, error) {

	fmt.Printf("\n\n%s:\n", strings.ToTitle("conditions"))
	fmt.Printf("%s\n\n", strings.Repeat("=", len("conditions")+1))

	for k, c := range conf.Conditions {

		c.Result = result.FAILURE

		conf.Conditions[k] = c
		ok, err := c.Run(
			conf.Source.Prefix + conf.Source.Output + conf.Source.Postfix)
		if err != nil {
			return false, err
		}

		if !ok {

			c.Result = result.FAILURE
			conf.Conditions[k] = c
			fmt.Printf("\n%s skipping: condition not met\n", result.FAILURE)
			ok = false
			return false, nil
		}
		c.Result = result.SUCCESS

		conf.Conditions[k] = c
	}

	return true, nil
}

// RunTargets iterate on every targets then call target on each of them
func RunTargets(config *config.Config, options *target.Options, report *reports.Report) (targetsChanged bool, err error) {
	targetsChanged = false

	fmt.Printf("\n\n%s:\n", strings.ToTitle("Targets"))
	fmt.Printf("%s\n\n", strings.Repeat("=", len("Targets")+1))

	sourceReport, err := report.String("source")

	if err != nil {
		fmt.Println(err)
	}
	conditionReport, err := report.String("conditions")

	if err != nil {
		fmt.Println(err)
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

			if err != nil {
				fmt.Println(err)
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
			fmt.Printf("Something went wrong in target \"%v\" :\n", id)
			fmt.Printf("%v\n\n", err)
		}

		if err != nil {
			fmt.Println(err)
			t.Result = result.FAILURE

		} else if targetChanged && err == nil {
			t.Result = result.CHANGED
			targetsChanged = true

		} else if !targetChanged && err == nil {
			t.Result = result.SUCCESS
		} else {
			t.Result = result.FAILURE
			err = fmt.Errorf("Unplanned target result")
			fmt.Println(err)
		}

		config.Targets[id] = t
	}
	return targetsChanged, nil
}

// Show displays configurations that should be apply
func (e *Engine) Show() error {

	for _, conf := range e.configurations {

		fmt.Printf("\n\n%s\n", strings.Repeat("#", len(conf.Name)+4))
		fmt.Printf("# %s #\n", strings.ToTitle(conf.Name))
		fmt.Printf("%s\n\n", strings.Repeat("#", len(conf.Name)+4))

		err := conf.Display()
		if err != nil {
			return err
		}

	}
	return nil
}
