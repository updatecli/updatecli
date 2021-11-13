package pipeline

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/engine/target"
	"github.com/updatecli/updatecli/pkg/core/pipeline/condition"
	"github.com/updatecli/updatecli/pkg/core/pipeline/source"
	"github.com/updatecli/updatecli/pkg/core/reports"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Pipeline represent an updatecli run for a specific configuration
type Pipeline struct {
	Name  string // Name defines a pipeline name, used to improve human visualization
	ID    string // ID allows to identify a full pipeline run, this value is propagated into each target if not defined at that level
	Title string // Title is used for the full pipelin

	Sources    map[string]source.Source
	Conditions map[string]condition.Condition
	Targets    map[string]target.Target

	Report reports.Report

	Options Options

	Config *config.Config
}

// Init initialize an updatecli context based on its configuration
func (p *Pipeline) Init(config *config.Config, options Options) {

	if len(config.Title) > 0 {
		p.Title = config.Title
	} else {
		p.Title = config.Name
	}

	p.Options = options

	p.Name = config.Name
	p.ID = config.PipelineID

	p.Config = config

	// Init context resource size
	p.Sources = make(map[string]source.Source, len(config.Sources))
	p.Conditions = make(map[string]condition.Condition, len(config.Conditions))
	p.Targets = make(map[string]target.Target, len(config.Targets))

	// Init context resource size
	p.Report.Sources = make(map[string]reports.Stage, len(config.Sources))
	p.Report.Conditions = make(map[string]reports.Stage, len(config.Conditions))
	p.Report.Targets = make(map[string]reports.Stage, len(config.Targets))
	p.Report.Name = config.Name
	p.Report.Result = result.SKIPPED

	// Init sources report
	for id := range config.Sources {
		// Init Sources[id]
		p.Sources[id] = source.Source{
			Config: config.Sources[id],
			Result: result.SKIPPED,
		}

		p.Report.Sources[id] = reports.Stage{
			Name:   config.Sources[id].Name,
			Kind:   config.Sources[id].Kind,
			Result: result.SKIPPED,
		}

	}

	// Init conditions report
	for id := range config.Conditions {

		p.Conditions[id] = condition.Condition{
			Config: config.Conditions[id],
			Result: result.SKIPPED,
		}

		p.Report.Conditions[id] = reports.Stage{
			Name:   config.Conditions[id].Name,
			Kind:   config.Conditions[id].Kind,
			Result: result.SKIPPED,
		}
	}

	// Init target report
	for id := range config.Targets {

		p.Targets[id] = target.Target{
			Config: config.Targets[id],
			Result: result.SKIPPED,
		}

		p.Report.Targets[id] = reports.Stage{
			Name:   config.Targets[id].Name,
			Kind:   config.Targets[id].Kind,
			Result: result.SKIPPED,
		}
	}

}

// Run execute an single pipeline
func (p *Pipeline) Run() error {

	logrus.Infof("\n\n%s\n", strings.Repeat("#", len(p.Title)+4))
	logrus.Infof("# %s #\n", strings.ToTitle(p.Title))
	logrus.Infof("%s\n", strings.Repeat("#", len(p.Title)+4))

	err := p.RunSources()

	if err != nil {
		p.Report.Result = result.FAILURE
		return fmt.Errorf("sources stage:\t%q", err.Error())
	}

	if len(p.Conditions) > 0 {

		ok, err := p.RunConditions()

		if err != nil {
			p.Report.Result = result.FAILURE
			return fmt.Errorf("conditions stage:\t%q", err.Error())
		} else if !ok {
			logrus.Infof("\n%s condition not met, skipping pipeline\n", result.FAILURE)
			return nil
		}

	}

	if len(p.Targets) > 0 {
		err := p.RunTargets()

		if err != nil {
			p.Report.Result = result.FAILURE
			return fmt.Errorf("targets stage:\t%q", err.Error())
		}
	}
	p.Report.Result = result.SUCCESS
	return nil

}

func (p *Pipeline) String() string {

	result := fmt.Sprintf("%q: %q\n", "Name", p.Name)
	result = result + fmt.Sprintf("%q: %q\n", "Title", p.Title)
	result = result + fmt.Sprintf("%q: %q\n", "ID", p.ID)

	result = result + fmt.Sprintf("%q:\n", "Sources")
	for key, value := range p.Sources {
		result = result + fmt.Sprintf("\t%q:\n", key)
		result = result + fmt.Sprintf("\t\t%q: %q\n", "Changelog", value.Changelog)
		result = result + fmt.Sprintf("\t\t%q: %q\n", "Output", value.Output)
		result = result + fmt.Sprintf("\t\t%q: %q\n", "Result", value.Result)
	}
	result = result + fmt.Sprintf("%q:\n", "Conditions")
	for key, value := range p.Conditions {
		result = result + fmt.Sprintf("\t%q:\n", key)
		result = result + fmt.Sprintf("\t\t%q: %q\n", "Result", value.Result)
	}
	result = result + fmt.Sprintf("%q:\n", "Targets")
	for key, value := range p.Targets {
		result = result + fmt.Sprintf("\t%q: %q\n", key, value.ReportTitle)
		result = result + fmt.Sprintf("\t\t%q: %q\n", "Result", value.Result)
	}

	return result
}
