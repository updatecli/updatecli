package engine

import (
	"fmt"

	"github.com/olblak/updateCli/pkg/config"
	"github.com/olblak/updateCli/pkg/reports"
	"github.com/olblak/updateCli/pkg/result"

	"path/filepath"
	"strings"
)

var engine Engine

// Engine defined parameters for a specific engine run
type Engine struct {
	conf    config.Config
	Options Options
	Report  reports.Report
}

// Run run the full process one yaml file
func (e *Engine) Run(cfgFile string) (report reports.Report, err error) {

	_, basename := filepath.Split(cfgFile)
	cfgFileName := strings.TrimSuffix(basename, filepath.Ext(basename))

	fmt.Printf("\n\n%s\n", strings.Repeat("#", len(cfgFileName)+4))
	fmt.Printf("# %s #\n", strings.ToTitle(cfgFileName))
	fmt.Printf("%s\n\n", strings.Repeat("#", len(cfgFileName)+4))

	err = e.conf.ReadFile(cfgFile, e.Options.ValuesFile)

	if err != nil {
		r := reports.Report{}
		r.Result = result.FAILURE
		r.Err = err.Error()
		r.Name = strings.ToTitle(cfgFileName)
		return r, err
	}

	e.Report = reports.New(&e.conf)
	e.Report.Name = strings.ToTitle(cfgFileName)

	err = e.conf.Source.Execute()

	if err != nil {
		return e.Report, err
	}

	if e.conf.Source.Output == "" {
		e.conf.Source.Result = result.FAILURE
		fmt.Printf("\n%s Something went wrong no value returned from Source", result.FAILURE)
		return e.Report, nil
	}

	e.conf.Source.Result = result.SUCCESS
	e.Report.Update(&e.conf)

	if len(e.conf.Conditions) > 0 {
		ok, err := e.conditions()
		e.Report.Update(&e.conf)
		if err != nil {
			return e.Report, err
		}

		if !ok {
			return e.Report, nil
		}
	}

	if len(e.conf.Targets) > 0 {
		changed, err := e.targets()

		if err != nil {
			return e.Report, err
		}
		if changed {
			e.Report.Result = result.CHANGED
		} else {
			e.Report.Result = result.SUCCESS
		}
		e.Report.Update(&e.conf)
	}

	return e.Report, nil
}

// conditions iterates on every conditions and test the result
func (e *Engine) conditions() (bool, error) {

	fmt.Printf("\n\n%s:\n", strings.ToTitle("conditions"))
	fmt.Printf("%s\n\n", strings.Repeat("=", len("conditions")+1))

	for k, c := range e.conf.Conditions {

		c.Result = result.FAILURE

		e.conf.Conditions[k] = c
		ok, err := c.Execute(
			e.conf.Source.Prefix + e.conf.Source.Output + e.conf.Source.Postfix)
		if err != nil {
			return false, err
		}

		if !ok {

			c.Result = result.FAILURE
			e.conf.Conditions[k] = c
			fmt.Printf("\n%s skipping: condition not met\n", result.FAILURE)
			ok = false
			return false, nil
		} else {
			c.Result = result.SUCCESS
			e.conf.Conditions[k] = c
		}
	}

	return true, nil
}

// targets iterate on every targets and then call target on each of them
func (e *Engine) targets() (targetsChanged bool, err error) {
	targetsChanged = false

	fmt.Printf("\n\n%s:\n", strings.ToTitle("Targets"))
	fmt.Printf("%s\n\n", strings.Repeat("=", len("Targets")+1))

	for id, t := range e.conf.Targets {
		targetChanged := false

		if t.Prefix == "" && e.conf.Source.Prefix != "" {
			t.Prefix = e.conf.Source.Prefix
		}

		if t.Postfix == "" && e.conf.Source.Postfix != "" {
			t.Postfix = e.conf.Source.Postfix
		}

		targetChanged, err = t.Execute(e.conf.Source.Output, &e.Options.Target)

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

		e.conf.Targets[id] = t
	}
	return targetsChanged, nil
}

// Show displays the configuration that should be apply
func (e *Engine) Show(cfgFile string) error {

	_, basename := filepath.Split(cfgFile)
	cfgFileName := strings.TrimSuffix(basename, filepath.Ext(basename))

	fmt.Printf("\n\n%s\n", strings.Repeat("#", len(cfgFileName)+4))
	fmt.Printf("# %s #\n", strings.ToTitle(cfgFileName))
	fmt.Printf("%s\n\n", strings.Repeat("#", len(cfgFileName)+4))

	e.conf.ReadFile(cfgFile, e.Options.ValuesFile)
	err := e.conf.Display()
	if err != nil {
		return err
	}

	return nil
}
