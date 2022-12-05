package pipeline

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/action"
	"github.com/updatecli/updatecli/pkg/core/pipeline/condition"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/pipeline/source"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
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
	SCMs       map[string]scm.Scm
	Actions    map[string]action.Action

	Report reports.Report

	Options Options

	Config *config.Config
}

// Init initialize an updatecli context based on its configuration
func (p *Pipeline) Init(config *config.Config, options Options) error {

	if len(config.Spec.Title) > 0 {
		p.Title = config.Spec.Title
	} else {
		p.Title = config.Spec.Name
	}

	p.Options = options

	p.Name = config.Spec.Name
	p.ID = config.Spec.PipelineID

	p.Config = config

	// Init context resource size
	p.SCMs = make(map[string]scm.Scm, len(config.Spec.SCMs))
	p.Sources = make(map[string]source.Source, len(config.Spec.Sources))
	p.Conditions = make(map[string]condition.Condition, len(config.Spec.Conditions))
	p.Targets = make(map[string]target.Target, len(config.Spec.Targets))
	p.Actions = make(map[string]action.Action, len(config.Spec.Actions))

	// Init context resource size
	p.Report.Sources = make(map[string]reports.Stage, len(config.Spec.Sources))
	p.Report.Conditions = make(map[string]reports.Stage, len(config.Spec.Conditions))
	p.Report.Targets = make(map[string]reports.Stage, len(config.Spec.Targets))
	p.Report.Name = config.Spec.Name
	p.Report.Result = result.SKIPPED

	// Init scm
	for id, scmConfig := range config.Spec.SCMs {
		// Init Sources[id]
		var err error

		// avoid gosec G601: Reassign the loop iteration variable to a local variable so the pointer address is correct
		scmConfig := scmConfig

		p.SCMs[id], err = scm.New(&scmConfig, config.Spec.PipelineID)
		if err != nil {
			return err
		}

	}

	// Init actions
	for id, actionConfig := range config.Spec.Actions {
		var err error

		// avoid gosec G601: Reassign the loop iteration variable to a local variable so the pointer address is correct
		actionConfig := actionConfig

		SCM, ok := p.SCMs[actionConfig.ScmID]

		// Validate that scm ID exists
		if !ok {
			return fmt.Errorf("scms ID %q referenced by the action id %q does not exist",
				actionConfig.ScmID,
				id)
		}

		p.Actions[id], err = action.New(
			&actionConfig,
			&SCM)

		if err != nil {
			return err
		}

	}

	// Init sources report
	for id := range config.Spec.Sources {
		// Set scm pointer
		var scmPointer *scm.ScmHandler
		if len(config.Spec.Sources[id].SCMID) > 0 {
			sc, ok := p.SCMs[config.Spec.Sources[id].SCMID]
			if !ok {
				return fmt.Errorf("scm ID %q from source ID %q doesn't exist",
					config.Spec.Sources[id].SCMID,
					id)
			}

			scmPointer = &sc.Handler
		}

		// Init Sources[id]
		p.Sources[id] = source.Source{
			Config: config.Spec.Sources[id],
			Result: result.SKIPPED,
			Scm:    scmPointer,
		}

		p.Report.Sources[id] = reports.Stage{
			Name:   config.Spec.Sources[id].Name,
			Kind:   config.Spec.Sources[id].Kind,
			Result: result.SKIPPED,
		}

	}

	// Init conditions report
	for id := range config.Spec.Conditions {

		// Set scm pointer
		var scmPointer *scm.ScmHandler
		if len(config.Spec.Conditions[id].SCMID) > 0 {
			sc, ok := p.SCMs[config.Spec.Conditions[id].SCMID]
			if !ok {
				return fmt.Errorf("scm id %q doesn't exist", config.Spec.Conditions[id].SCMID)
			}

			scmPointer = &sc.Handler
		}

		p.Conditions[id] = condition.Condition{
			Config: config.Spec.Conditions[id],
			Result: result.SKIPPED,
			Scm:    scmPointer,
		}

		p.Report.Conditions[id] = reports.Stage{
			Name:   config.Spec.Conditions[id].Name,
			Kind:   config.Spec.Conditions[id].Kind,
			Result: result.SKIPPED,
		}
	}

	// Init target report
	for id := range config.Spec.Targets {

		var scmPointer *scm.ScmHandler
		if len(config.Spec.Targets[id].SCMID) > 0 {
			sc, ok := p.SCMs[config.Spec.Targets[id].SCMID]
			if !ok {
				return fmt.Errorf("scm id %q doesn't exist", config.Spec.Targets[id].SCMID)
			}

			scmPointer = &sc.Handler
		}

		p.Targets[id] = target.Target{
			Config: config.Spec.Targets[id],
			Result: result.SKIPPED,
			Scm:    scmPointer,
		}

		p.Report.Targets[id] = reports.Stage{
			Name:   config.Spec.Targets[id].Name,
			Kind:   config.Spec.Targets[id].Kind,
			Result: result.SKIPPED,
		}
	}
	return nil

}

// Run execute an single pipeline
func (p *Pipeline) Run() error {

	logrus.Infof("\n\n%s\n", strings.Repeat("#", len(p.Title)+4))
	logrus.Infof("# %s #\n", strings.ToTitle(p.Title))
	logrus.Infof("%s\n", strings.Repeat("#", len(p.Title)+4))

	if len(p.Sources) > 0 {
		err := p.RunSources()

		if err != nil {
			p.Report.Result = result.FAILURE
			return fmt.Errorf("sources stage:\t%q", err.Error())
		}
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

	if len(p.Actions) > 0 {
		err := p.RunActions()

		if err != nil {
			p.Report.Result = result.FAILURE
			return fmt.Errorf("action stage:\t%q", err.Error())
		}

	}

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
		result = result + fmt.Sprintf("\t%q:\n", key)
		result = result + fmt.Sprintf("\t\t%q: %q\n", "Result", value.Result)
	}

	return result
}
