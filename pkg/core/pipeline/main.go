package pipeline

import (
	"fmt"
	"strings"
	"sync"

	"github.com/heimdalr/dag"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/action"
	"github.com/updatecli/updatecli/pkg/core/pipeline/condition"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
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
	mu     sync.Mutex
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
	p.Report.PipelineID = config.Spec.PipelineID

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

		p.Report.Targets[id].DryRun = r.DryRun
	}
	return nil

}

func (p *Pipeline) runFlowCallback(d *dag.DAG, id string, depsResults []dag.FlowResult) (v interface{}, err error) {
	p.mu.Lock()         // Acquire lock at the start
	defer p.mu.Unlock() // Release lock when function exits
	if id == rootVertex {
		return nil, nil
	}
	err = p.Update()
	if err != nil {
		return nil, fmt.Errorf("update pipeline: %w", err)
	}
	v, err = d.GetVertex(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaf from dag: %w", err) // Should never happens
	}
	leaf, ok := v.(Node)
	if !ok {
		return nil, fmt.Errorf("failed to reconstruct leaf from interface: %s", id) // Should never happens
	}
	logrus.Infof("\n%s: %s\n", leaf.Category, id)
	logrus.Infof("%s\n", strings.Repeat("-", len(id)))

	depsSourceIDs := []string{}
	deps := map[string]*Node{}
	for _, r := range depsResults {
		if r.ID == rootVertex {
			continue
		}

		p, _ := r.Result.(Node)
		deps[r.ID] = &p

		if p.Category == sourceCategory {
			// source id order, is not guaranteed as the information is coming from a map
			depsSourceIDs = append(depsSourceIDs, strings.TrimPrefix(r.ID, "source#"))
		}
	}

	updateSourceResult := func(id string) {
		var err error

		source := p.Sources[id]

		source.Result.Name = source.Config.Name
		source.Result.Config, err = resource.GetReportConfig(p.Config.Spec.Sources[id].ResourceConfig)

		if err != nil {
			logrus.Errorf("error while cleaning config: %v", err)
			logrus.Errorf("Config: %+v", p.Config.Spec.Sources[id].ResourceConfig)
		}

		p.Report.Sources[id] = &source.Result
	}

	updateConditionResult := func(id string) {
		var err error

		condition := p.Conditions[id]

		condition.Result.SourceID = condition.Config.SourceID
		if condition.Config.SourceID == "" && len(depsSourceIDs) > 0 {
			condition.Result.SourceID = depsSourceIDs[0]
		}

		condition.Result.Name = p.Config.Spec.Conditions[id].Name
		condition.Result.Config, err = resource.GetReportConfig(p.Config.Spec.Conditions[id].ResourceConfig)
		if err != nil {
			logrus.Errorf("error while cleaning config: %v", err)
			logrus.Errorf("Config: %+v", p.Config.Spec.Conditions[id].ResourceConfig)
		}

		p.Report.Conditions[id] = &condition.Result
	}

	updateTargetResult := func(id string) {
		var err error
		target := p.Targets[id]

		target.Result.SourceID = target.Config.SourceID
		if target.Config.SourceID == "" && len(depsSourceIDs) > 0 {
			target.Result.SourceID = depsSourceIDs[0]
		}

		target.Result.Name = p.Config.Spec.Targets[id].Name
		target.Result.Config, err = resource.GetReportConfig(p.Config.Spec.Targets[id].ResourceConfig)
		target.Result.DryRun = target.DryRun
		if err != nil {
			logrus.Errorf("error while cleaning config: %v", err)
			logrus.Errorf("Config: %+v", p.Config.Spec.Targets[id].ResourceConfig)
		}

		p.Report.Targets[id] = &target.Result
	}

	shouldSkip := p.shouldSkipResource(&leaf, deps)
	if shouldSkip {
		logrus.Debugf("Skipping %s[%q] because of dependsOn conditions", leaf.Category, id)
		leaf.Result = result.SKIPPED

		switch leaf.Category {
		case sourceCategory:
			sourceId := strings.ReplaceAll(id, "source#", "")
			updateSourceResult(sourceId)

		case conditionCategory:
			conditionId := strings.ReplaceAll(id, "condition#", "")
			updateConditionResult(conditionId)
		case targetCategory:
			targetId := strings.ReplaceAll(id, "target#", "")
			updateTargetResult(targetId)
		}
	}

	if leaf.Result != result.SKIPPED {
		// Run the resource
		switch leaf.Category {
		case sourceCategory:
			sourceId := strings.ReplaceAll(id, "source#", "")
			r, e := p.RunSource(sourceId)
			if e != nil {
				err = e
			}

			updateSourceResult(sourceId)

			leaf.Result = r
			p.updateSource(sourceId, leaf.Result)
		case conditionCategory:
			conditionId := strings.ReplaceAll(id, "condition#", "")
			r, e := p.RunCondition(conditionId)
			if e != nil {
				err = e
			}

			updateConditionResult(conditionId)

			leaf.Result = r
			p.updateCondition(conditionId, leaf.Result)
		case targetCategory:
			targetId := strings.ReplaceAll(id, "target#", "")
			r, changed, e := p.RunTarget(targetId, depsSourceIDs)
			if e != nil {
				err = e
			}

			updateTargetResult(targetId)

			leaf.Result = r
			leaf.Changed = changed
			p.updateTarget(targetId, leaf.Result)
		}
	}
	if leaf.Changed {
		p.Report.Result = result.ATTENTION
	}
	return leaf, err
}

// Run execute an single pipeline
func (p *Pipeline) Run() error {

	logrus.Infof("\n\n%s\n", strings.Repeat("#", len(p.Name)+4))
	logrus.Infof("# %s #\n", strings.ToTitle(p.Name))
	logrus.Infof("%s\n", strings.Repeat("#", len(p.Name)+4))

	p.Report.Result = result.SUCCESS

	resources, err := p.SortedResources()
	if err != nil {
		p.Report.Result = result.FAILURE
		return fmt.Errorf("could not create dag from spec:\t%q", err.Error())
	}
	leaves, err := resources.DescendantsFlow(rootVertex, nil, p.runFlowCallback)
	if err != nil {
		p.Report.Result = result.FAILURE
		return fmt.Errorf("could not parse dag from spec:\t%q", err.Error())
	}

	hasError := false
	for _, leaf := range leaves {
		if leaf.ID == rootVertex {
			// ignore
			continue
		}
		if leaf.Error != nil {
			logrus.Infof("\n")
			logrus.Errorf("something went wrong in %q : %s", leaf.ID, leaf.Error)
			logrus.Infof("\n")
			hasError = true
		}
	}
	if hasError {
		p.Report.Result = result.FAILURE
		return ErrRunTargets
	}

	// set pipeline report result
	if len(p.Targets) > 0 {
		successCounter := 0
		skippedCounter := 0
		attentionCounter := 0

		for id := range p.Targets {
			switch p.Targets[id].Result.Result {
			case result.FAILURE:
				p.Report.Result = result.FAILURE
				return nil
			case result.SUCCESS:
				successCounter++
			case result.SKIPPED:
				skippedCounter++
			case result.ATTENTION:
				attentionCounter++
			}
		}

		if len(p.Targets) == skippedCounter {
			p.Report.Result = result.SKIPPED
			return nil
		} else if len(p.Targets) == successCounter+skippedCounter {
			p.Report.Result = result.SUCCESS
		} else if attentionCounter > 0 {
			p.Report.Result = result.ATTENTION
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
		result = result + fmt.Sprintf("\t\t%q: %q\n", "Changelog", value.Result.Changelogs)
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
