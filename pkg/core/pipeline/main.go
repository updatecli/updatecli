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
	// Name defines a pipeline name, used to improve human visualization
	Name string
	// ID allows to identify a full pipeline run, this value is propagated into each target if not defined at that level
	ID string
	// Sources contains all sources defined in the configuration
	Sources map[string]source.Source
	// Conditions contains all conditions defined in the configuration
	Conditions map[string]condition.Condition
	// Targets contains all targets defined in the configuration
	Targets map[string]target.Target
	// SCMs contains all scms defined in the configuration
	SCMs map[string]scm.Scm
	// Actions contains all actions defined in the configuration
	Actions map[string]action.Action
	// Report contains the pipeline report
	Report reports.Report
	// Options contains all updatecli options for this specific pipeline
	Options Options
	// Config contains the pipeline configuration defined by the user
	Config *config.Config
}

// Init initialize an updatecli context based on its configuration
func (p *Pipeline) Init(config *config.Config, options Options) error {

	p.Name = config.Spec.Name
	if len(config.Spec.Title) > 0 && p.Name == "" {
		p.Name = config.Spec.Title
	}

	p.Options = options

	p.ID = config.Spec.PipelineID

	p.Config = config

	// Init context resource size
	p.SCMs = make(map[string]scm.Scm, len(config.Spec.SCMs))
	p.Sources = make(map[string]source.Source, len(config.Spec.Sources))
	p.Conditions = make(map[string]condition.Condition, len(config.Spec.Conditions))
	p.Targets = make(map[string]target.Target, len(config.Spec.Targets))
	p.Actions = make(map[string]action.Action, len(config.Spec.Actions))

	// Init context resource size
	p.Report.Sources = make(map[string]*result.Source, len(config.Spec.Sources))
	p.Report.Conditions = make(map[string]*result.Condition, len(config.Spec.Conditions))
	p.Report.Targets = make(map[string]*result.Target, len(config.Spec.Targets))
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
			Result: result.Source{
				Result: result.SKIPPED,
			},
			Scm: scmPointer,
		}

		r := p.Sources[id].Result
		p.Report.Sources[id] = &r
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
			Result: result.Condition{
				Result: result.SKIPPED,
			},
			Scm: scmPointer,
		}

		r := p.Conditions[id].Result
		p.Report.Conditions[id] = &r

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
			Result: result.Target{
				Result: result.SKIPPED,
			},
			Scm: scmPointer,
		}

		r := p.Targets[id].Result
		p.Report.Targets[id] = &r
	}
	return nil

}

// Run execute an single pipeline
func (p *Pipeline) Run() error {

	logrus.Infof("\n\n%s\n", strings.Repeat("#", len(p.Name)+4))
	logrus.Infof("# %s #\n", strings.ToTitle(p.Name))
	logrus.Infof("%s\n", strings.Repeat("#", len(p.Name)+4))

	if len(p.Sources) > 0 {
		if err := p.RunSources(); err != nil {
			p.Report.Result = result.FAILURE
			return fmt.Errorf("sources stage:\t%q", err.Error())
		}
	}

	if len(p.Conditions) > 0 {
		if err := p.RunConditions(); err != nil {
			p.Report.Result = result.FAILURE
			return fmt.Errorf("conditions stage:\t%q", err.Error())
		}
	}

	if len(p.Targets) > 0 {
		if err := p.RunTargets(); err != nil {
			p.Report.Result = result.FAILURE
			return fmt.Errorf("targets stage:\t%q", err.Error())
		}
	}

	return nil

}

func (p *Pipeline) String() string {

	result := fmt.Sprintf("%q: %q\n", "Name", p.Name)
	result = result + fmt.Sprintf("%q: %q\n", "ID", p.ID)

	result = result + fmt.Sprintf("%q:\n", "Sources")
	for key, value := range p.Sources {
		result = result + fmt.Sprintf("\t%q:\n", key)
		result = result + fmt.Sprintf("\t\t%q: %q\n", "Changelog", value.Changelog)
		result = result + fmt.Sprintf("\t\t%q: %q\n", "Output", value.Output)
		result = result + fmt.Sprintf("\t\t%q: %q\n", "Result", value.Result.Result)
	}
	result = result + fmt.Sprintf("%q:\n", "Conditions")
	for key, value := range p.Conditions {
		result = result + fmt.Sprintf("\t%q:\n", key)
		result = result + fmt.Sprintf("\t\t%q: %q\n", "Result", value.Result.Result)
	}
	result = result + fmt.Sprintf("%q:\n", "Targets")
	for key, value := range p.Targets {
		result = result + fmt.Sprintf("\t%q:\n", key)
		result = result + fmt.Sprintf("\t\t%q: %q\n", "Result", value.Result.Result)
	}

	return result
}

func (p *Pipeline) Update() error {
	err := p.Config.Update(p)
	if err != nil {
		return err
	}

	// Reset scm
	for id, scmConfig := range p.Config.Spec.SCMs {
		var err error

		// avoid gosec G601: Reassign the loop iteration variable to a local variable so the pointer address is correct
		scmConfig := scmConfig

		p.SCMs[id], err = scm.New(&scmConfig, p.Config.Spec.PipelineID)
		if err != nil {
			return err
		}
	}

	// Update scm pointer for each actions
	for id := range p.Config.Spec.Actions {
		action := p.Actions[id]

		if len(p.Config.Spec.Actions[id].ScmID) > 0 {
			sc, ok := p.SCMs[p.Config.Spec.Actions[id].ScmID]
			if !ok {
				return fmt.Errorf("scm id %q doesn't exist", p.Config.Spec.Actions[id].ScmID)
			}

			action.Scm = &sc
		}
		p.Actions[id] = action
	}
	// Update scm pointer for each condition
	for id := range p.Config.Spec.Conditions {
		condition := p.Conditions[id]

		if len(p.Config.Spec.Conditions[id].SCMID) > 0 {
			sc, ok := p.SCMs[p.Config.Spec.Conditions[id].SCMID]
			if !ok {
				return fmt.Errorf("scm id %q doesn't exist", p.Config.Spec.Conditions[id].SCMID)
			}

			condition.Scm = &sc.Handler
		}
		p.Conditions[id] = condition
	}

	// Update scm pointer for each sources
	for id := range p.Config.Spec.Sources {
		source := p.Sources[id]

		if len(p.Config.Spec.Sources[id].SCMID) > 0 {
			sc, ok := p.SCMs[p.Config.Spec.Sources[id].SCMID]
			if !ok {
				return fmt.Errorf("scm id %q doesn't exist", p.Config.Spec.Conditions[id].SCMID)
			}
			source.Scm = &sc.Handler
		}
		p.Sources[id] = source
	}

	// Update scm pointer for each target
	for id := range p.Config.Spec.Targets {
		target := p.Targets[id]

		if len(p.Config.Spec.Targets[id].SCMID) > 0 {
			sc, ok := p.SCMs[p.Config.Spec.Targets[id].SCMID]
			if !ok {
				return fmt.Errorf("scm id %q doesn't exist", p.Config.Spec.Targets[id].SCMID)
			}
			target.Scm = &sc.Handler
		}
		p.Targets[id] = target
	}

	return nil
}
