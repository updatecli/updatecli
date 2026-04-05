package pipeline

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/heimdalr/dag"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/cache"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/action"
	"github.com/updatecli/updatecli/pkg/core/pipeline/condition"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/pipeline/source"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
	"github.com/updatecli/updatecli/pkg/core/reports"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
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
	// CrawlerKind identifies the autodiscovery crawler that generated this pipeline (empty for user-defined pipelines).
	CrawlerKind string
	// SourceCache is a shared in-memory cache for source execution results,
	// injected by the engine before the pipeline runs.
	SourceCache *cache.SourceCache
	tracer      trace.Tracer
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
	p.Report.Labels = config.Spec.Labels

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
			Result: &result.Source{
				Result: result.SKIPPED,
			},
			Scm: scmPointer,
		}

		r := p.Sources[id].Result

		if scmPointer != nil {
			scm := *scmPointer
			r.Scm.URL = scm.GetURL()
			r.Scm.Branch.Source, r.Scm.Branch.Working, r.Scm.Branch.Target = scm.GetBranches()
		}

		p.Report.Sources[id] = r
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
			Result: &result.Condition{
				Result: result.SKIPPED,
			},
			Scm: scmPointer,
		}

		r := p.Conditions[id].Result

		if scmPointer != nil {
			scm := *scmPointer
			r.Scm.URL = scm.GetURL()
			r.Scm.Branch.Source, r.Scm.Branch.Working, r.Scm.Branch.Target = scm.GetBranches()
		}

		p.Report.Conditions[id] = r

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
			Result: &result.Target{
				Result: result.SKIPPED,
			},
			Scm: scmPointer,
		}

		r := p.Targets[id].Result

		if scmPointer != nil {
			scm := *scmPointer
			r.Scm.URL = scm.GetURL()
			r.Scm.Branch.Source, r.Scm.Branch.Working, r.Scm.Branch.Target = scm.GetBranches()
		}

		p.Report.Targets[id] = r

		p.Report.Targets[id].DryRun = r.DryRun
	}

	p.tracer = telemetry.Tracer("updatecli")

	// Graph must be generated after all resources have been initialized !
	graph, err := p.Graph(GraphFlavorMermaid)
	switch err {
	case nil:
		p.Report.Graph = graph
	default:
		logrus.Errorf("generating pipeline graph:\n%s", err)
	}

	return nil

}

// Run executes a single pipeline, wrapping the entire run in an OpenTelemetry span.
func (p *Pipeline) Run(ctx context.Context) error {
	logrus.Infof("\n%s\n", strings.Repeat("#", len(p.Name)+4))
	logrus.Infof("# %s #\n", strings.ToTitle(p.Name))
	logrus.Infof("%s\n\n", strings.Repeat("#", len(p.Name)+4))

	tracer := p.tracer

	opts := []trace.SpanStartOption{
		trace.WithAttributes(
			attribute.String("updatecli.pipeline.name", p.Name),
			attribute.String("updatecli.pipeline.id", p.ID),
			attribute.Int("updatecli.pipeline.sources_count", len(p.Sources)),
			attribute.Int("updatecli.pipeline.conditions_count", len(p.Conditions)),
			attribute.Int("updatecli.pipeline.targets_count", len(p.Targets)),
		),
	}

	if p.CrawlerKind != "" {
		opts = append(opts, trace.WithAttributes(attribute.String("updatecli.pipeline.crawler_kind", p.CrawlerKind)))
	}

	ctx, span := tracer.Start(ctx, "updatecli.pipeline", opts...)
	defer span.End()

	logrus.Infof("Pipeline ID\t: %s", p.ID)
	if p.Options.Target.DryRun {
		logrus.Infof("Dry Run\t\t: enabled\n")
		span.SetAttributes(attribute.Bool("updatecli.pipeline.dry_run", true))
	}

	p.Report.Result = result.SUCCESS

	resources, err := p.SortedResources()
	if err != nil {
		p.Report.Result = result.FAILURE
		span.RecordError(err)
		span.SetStatus(codes.Error, "dag creation failed")
		return fmt.Errorf("could not create dag from spec:\t%q", err.Error())
	}

	// Closure captures ctx so each DAG node callback can create a child span.
	callback := func(d *dag.DAG, id string, depsResults []dag.FlowResult) (interface{}, error) {
		return p.runFlowCallbackWithCtx(ctx, d, id, depsResults)
	}

	leaves, err := resources.DescendantsFlow(rootVertex, nil, callback)
	if err != nil {
		p.Report.Result = result.FAILURE
		span.RecordError(err)
		span.SetStatus(codes.Error, "dag execution failed")
		return fmt.Errorf("could not parse dag from spec:\t%q", err.Error())
	}

	hasError := false
	for _, leaf := range leaves {
		if leaf.ID == rootVertex {
			continue
		}
		if leaf.Error != nil {
			if !hasError {
				title := fmt.Sprintf("Error(s) found in pipeline %q execution", p.Name)
				logrus.Infof("\n%s\n%s\n", title, strings.Repeat("-", len(title)))
			}
			logrus.Infof("\n")
			logrus.Errorf("something went wrong in %q :\n\t%s", leaf.ID, strings.ReplaceAll(leaf.Error.Error(), "\n", "\n\t"))
			logrus.Infof("\n")
			hasError = true
		}
	}
	if hasError {
		p.Report.Result = result.FAILURE
		span.SetAttributes(attribute.String("updatecli.pipeline.result", result.FAILURE))
		span.SetStatus(codes.Error, "pipeline had errors")
		return ErrRunTargets
	}

	// Set pipeline report result based on target outcomes.
	if len(p.Targets) > 0 {
		successCounter := 0
		skippedCounter := 0
		attentionCounter := 0

		for id := range p.Targets {
			switch p.Targets[id].Result.Result {
			case result.FAILURE:
				p.Report.Result = result.FAILURE
				span.SetAttributes(attribute.String("updatecli.pipeline.result", result.FAILURE))
				span.SetStatus(codes.Error, "target reported failure")
				return nil
			case result.SUCCESS:
				successCounter++
			case result.SKIPPED:
				skippedCounter++
			case result.ATTENTION:
				attentionCounter++
			}
		}

		// No early return on SKIPPED — fall through so span attributes are always recorded.
		if len(p.Targets) == skippedCounter {
			p.Report.Result = result.SKIPPED
		} else if len(p.Targets) == successCounter+skippedCounter {
			p.Report.Result = result.SUCCESS
		} else if attentionCounter > 0 {
			p.Report.Result = result.ATTENTION
		}
	}

	span.SetAttributes(attribute.String("updatecli.pipeline.result", p.Report.Result))
	return nil
}

func (p *Pipeline) runFlowCallbackWithCtx(ctx context.Context, d *dag.DAG, id string, depsResults []dag.FlowResult) (v interface{}, err error) {
	if id == rootVertex {
		return nil, nil
	}

	// Start span before lock so timing reflects actual work, not mutex wait.
	_, span := p.tracer.Start(ctx, "updatecli.resource",
		trace.WithAttributes(
			attribute.String("updatecli.resource.id", id),
		),
	)
	defer span.End()

	p.mu.Lock()
	defer p.mu.Unlock()

	err = p.Update()
	if err != nil {
		return nil, fmt.Errorf("update pipeline: %w", err)
	}

	v, err = d.GetVertex(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaf from dag: %w", err) // unreachable: DAG guarantees vertex exists
	}
	leaf, ok := v.(Node)
	if !ok {
		return nil, fmt.Errorf("failed to reconstruct leaf from interface: %s", id) // unreachable: all vertices are Node
	}

	span.SetAttributes(attribute.String("updatecli.resource.category", leaf.Category))

	resourceTitle := strings.TrimPrefix(id, leaf.Category+"#")
	logrus.Infof("\n%s: %s\n", leaf.Category, resourceTitle)
	logrus.Infof("%s\n\n", strings.Repeat("-", len(resourceTitle)+len(leaf.Category)+2))

	// Display repository header when the resource is bound to an SCM.
	p.logRepositoryHeader(id, leaf.Category)

	depsSourceIDs := []string{}
	deps := map[string]*Node{}
	for _, r := range depsResults {
		if r.ID == rootVertex {
			continue
		}
		p, _ := r.Result.(Node)
		deps[r.ID] = &p
		if p.Category == sourceCategory {
			// source id order is not guaranteed as the information is coming from a map
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
		}
		p.Report.Sources[id] = source.Result
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
		}
		p.Report.Conditions[id] = condition.Result
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
		}
		p.Report.Targets[id] = target.Result
	}

	shouldSkip := p.shouldSkipResource(&leaf, deps)
	if shouldSkip {
		logrus.Infof("%s Skipping %q because of dependsOn conditions", result.SKIPPED, id)
		leaf.Result = result.SKIPPED
		span.SetAttributes(attribute.String("updatecli.resource.result", result.SKIPPED))

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

	displayError := func(err error) {
		logrus.Infof("%s Something went wrong:\n\t%s\n\n",
			result.FAILURE,
			strings.ReplaceAll(err.Error(), "\n", "\n\t"),
		)
	}

	if leaf.Result != result.SKIPPED {
		switch leaf.Category {
		case sourceCategory:
			sourceId := strings.ReplaceAll(id, "source#", "")
			r, e := p.RunSource(ctx, sourceId)
			if e != nil {
				displayError(e)
				err = e
				telemetry.RecordSpanError(span, e)
			}
			updateSourceResult(sourceId)
			leaf.Result = r
			p.updateSource(sourceId, leaf.Result)
			p.recordSourceSpan(span, sourceId, leaf.Result)

		case conditionCategory:
			conditionId := strings.ReplaceAll(id, "condition#", "")
			r, e := p.RunCondition(ctx, conditionId)
			if e != nil {
				displayError(e)
				err = e
				telemetry.RecordSpanError(span, e)
			}
			updateConditionResult(conditionId)
			leaf.Result = r
			p.updateCondition(conditionId, leaf.Result)
			p.recordConditionSpan(span, conditionId, leaf.Result)

		case targetCategory:
			targetId := strings.ReplaceAll(id, "target#", "")
			r, changed, e := p.RunTarget(ctx, targetId, depsSourceIDs)
			if e != nil {
				displayError(e)
				err = e
				telemetry.RecordSpanError(span, e)
			}
			updateTargetResult(targetId)
			leaf.Result = r
			leaf.Changed = changed
			p.updateTarget(targetId, leaf.Result)
			p.recordTargetSpan(span, targetId, leaf.Result, changed)
		}
	}

	if leaf.Changed {
		p.Report.Result = result.ATTENTION
	}
	return leaf, err
}

// logRepositoryHeader logs the SCM repository summary for the resource bound to id,
// stripping the category prefix to resolve the resource configuration.
func (p *Pipeline) logRepositoryHeader(id string, category string) {
	switch category {
	case sourceCategory:
		sourceId := strings.ReplaceAll(id, "source#", "")
		if scm, ok := p.SCMs[p.Sources[sourceId].Config.SCMID]; ok {
			logrus.Infof("Repository\t: %s\n\n", scm.Handler.Summary())
		}
	case conditionCategory:
		conditionId := strings.ReplaceAll(id, "condition#", "")
		if scm, ok := p.SCMs[p.Conditions[conditionId].Config.SCMID]; ok {
			logrus.Infof("Repository\t: %s\n\n", scm.Handler.Summary())
		}
	case targetCategory:
		targetId := strings.ReplaceAll(id, "target#", "")
		if scm, ok := p.SCMs[p.Targets[targetId].Config.SCMID]; ok {
			logrus.Infof("Repository\t: %s\n\n", scm.Handler.Summary())
		}
	}
}

// recordSourceSpan sets OpenTelemetry span attributes for a completed source run.
func (p *Pipeline) recordSourceSpan(span trace.Span, sourceId string, leafResult string) {
	src, ok := p.Sources[sourceId]
	if !ok {
		return
	}
	span.SetAttributes(
		attribute.String("updatecli.resource.name", src.Config.Name),
		attribute.String("updatecli.resource.kind", src.Config.Kind),
		attribute.String("updatecli.resource.result", leafResult),
	)
	if src.Result != nil && src.Result.Description != "" {
		span.SetAttributes(attribute.String("updatecli.resource.description", src.Result.Description))
	}
}

// recordConditionSpan sets OpenTelemetry span attributes for a completed condition run.
func (p *Pipeline) recordConditionSpan(span trace.Span, conditionId string, leafResult string) {
	cond, ok := p.Conditions[conditionId]
	if !ok {
		return
	}
	span.SetAttributes(
		attribute.String("updatecli.resource.name", cond.Config.Name),
		attribute.String("updatecli.resource.kind", cond.Config.Kind),
		attribute.String("updatecli.resource.result", leafResult),
	)
	if cond.Result != nil {
		span.SetAttributes(
			attribute.Bool("updatecli.condition.pass", cond.Result.Pass),
			attribute.String("updatecli.condition.source_id", cond.Result.SourceID),
		)
		if cond.Result.Description != "" {
			span.SetAttributes(attribute.String("updatecli.resource.description", cond.Result.Description))
		}
	}
}

// recordTargetSpan sets OpenTelemetry span attributes for a completed target run.
func (p *Pipeline) recordTargetSpan(span trace.Span, targetId string, leafResult string, changed bool) {
	tgt, ok := p.Targets[targetId]
	if !ok {
		return
	}
	span.SetAttributes(
		attribute.String("updatecli.resource.name", tgt.Config.Name),
		attribute.String("updatecli.resource.kind", tgt.Config.Kind),
		attribute.String("updatecli.resource.result", leafResult),
		attribute.Bool("updatecli.target.changed", changed),
	)
	if tgt.Result != nil {
		span.SetAttributes(
			attribute.Bool("updatecli.target.dry_run", tgt.Result.DryRun),
			attribute.String("updatecli.target.source_id", tgt.Result.SourceID),
		)
		if tgt.Result.Description != "" {
			span.SetAttributes(attribute.String("updatecli.resource.description", tgt.Result.Description))
		}
		if len(tgt.Result.Files) > 0 {
			span.SetAttributes(attribute.StringSlice("updatecli.target.files", tgt.Result.Files))
		}
	}
	if changed {
		span.AddEvent("target.changed")
	}
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

// Update updates the pipeline based on the latest configuration
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

		action.Title = p.Config.Spec.Actions[id].Title
		action.Config = p.Config.Spec.Actions[id]

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

		condition.Config = p.Config.Spec.Conditions[id]

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

		source.Config = p.Config.Spec.Sources[id]

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

		target.Config = p.Config.Spec.Targets[id]

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
