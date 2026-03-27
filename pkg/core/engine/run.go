package engine

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/cache"
	"github.com/updatecli/updatecli/pkg/core/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Run runs the full process under the provided context, emitting
// OpenTelemetry spans for each major operation so failures are observable.
func (e *Engine) Run(ctx context.Context) (err error) {
	PrintTitle("Pipeline")

	tracer := e.tracer
	ctx, span := tracer.Start(ctx, "updatecli.run",
		trace.WithAttributes(
			attribute.Int("updatecli.pipeline_count", len(e.Pipelines)),
			attribute.Bool("updatecli.dry_run", e.Options.Pipeline.Target.DryRun),
		),
	)
	defer span.End()

	errs := []error{}

	if e.sourceCache == nil {
		e.sourceCache = cache.NewSourceCache()
	}

	for i := range e.Pipelines {
		pipeline := e.Pipelines[i]
		pipeline.SourceCache = e.sourceCache

		err := pipeline.Run(ctx)
		if err != nil {
			errs = append(errs, fmt.Errorf("pipeline %q failed: %w", pipeline.Name, err))
			span.AddEvent("pipeline.failed", trace.WithAttributes(
				attribute.String("pipeline.name", pipeline.Name),
				attribute.String("error", telemetry.SanitizeError(err)),
			))
			logrus.Printf("Pipeline %q failed\n", pipeline.Name)
			logrus.Printf("Skipping due to:\n\t%s\n", err)
			continue
		}
	}

	if !e.Options.Pipeline.Target.DryRun && e.Options.Pipeline.Target.Push {
		_, pushSpan := tracer.Start(ctx, "updatecli.push_commits")
		if err = e.pushSCMCommits(); err != nil {
			errs = append(errs, fmt.Errorf("pushing commits failed: %w", err))
			telemetry.RecordSpanError(pushSpan, err)
		}
		pushSpan.End()
	}

	actionsCtx, actionsSpan := tracer.Start(ctx, "updatecli.run_actions")
	if err = e.runActions(actionsCtx); err != nil {
		errs = append(errs, fmt.Errorf("running actions failed: %w", err))
		telemetry.RecordSpanError(actionsSpan, err)
	}
	actionsSpan.End()

	if !e.Options.Pipeline.Target.DryRun && e.Options.Pipeline.Target.Push && e.Options.Pipeline.Target.CleanGitBranches {
		_, pruneSpan := tracer.Start(ctx, "updatecli.prune_scm_branches")
		if err = e.pruneSCMBranches(); err != nil {
			errs = append(errs, fmt.Errorf("cleaning git branches failed: %w", err))
			telemetry.RecordSpanError(pruneSpan, err)
		}
		pruneSpan.End()
	}

	for i := range e.Pipelines {
		pipeline := e.Pipelines[i]
		err = pipeline.Report.UpdateID()
		if err != nil {
			errs = append(errs, fmt.Errorf("updating report ID failed: %w", err))
		}
		e.Reports = append(e.Reports, pipeline.Report)
	}

	if err = e.publishToUdash(); err != nil {
		errs = append(errs, fmt.Errorf("publishing to Udash failed: %w", err))
	}

	if err = e.exportReportToYAML(false); err != nil {
		errs = append(errs, fmt.Errorf("exporting report to YAML failed: %w", err))
	}

	if err = e.showReports(); err != nil {
		errs = append(errs, fmt.Errorf("showing reports failed: %w", err))
	}

	if len(errs) > 0 {
		for _, e := range errs {
			logrus.Error(e)
		}
		span.SetStatus(codes.Error, fmt.Sprintf("%d pipeline(s) failed", len(errs)))
		return fmt.Errorf("%d pipeline(s) failed during execution", len(errs))
	}

	return nil
}
