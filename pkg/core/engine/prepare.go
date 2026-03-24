package engine

import (
	"context"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/telemetry"
	"github.com/updatecli/updatecli/pkg/core/tmp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Prepare runs all preparation phases under the provided context,
// emitting an OTel span for the overall prepare phase and one child span per sub-phase.
func (e *Engine) Prepare(ctx context.Context) (err error) {
	PrintTitle("Prepare")

	tracer := telemetry.Tracer("updatecli")
	ctx, span := tracer.Start(ctx, "updatecli.prepare")
	defer span.End()

	var defaultCrawlersEnabled bool

	err = tmp.Create()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	{
		_, loadSpan := tracer.Start(ctx, "updatecli.load_configurations")
		err = e.LoadConfigurations()
		if !errors.Is(err, ErrNoManifestDetected) && err != nil {
			logrus.Errorln(err)
			logrus.Infof("\n%d pipeline(s) successfully loaded\n", len(e.Pipelines))
			loadSpan.RecordError(err)
			loadSpan.SetStatus(codes.Error, err.Error())
		}
		if errors.Is(err, ErrNoManifestDetected) {
			defaultCrawlersEnabled = true
		}
		loadSpan.SetAttributes(attribute.Int("updatecli.pipelines_loaded", len(e.Pipelines)))
		loadSpan.End()
	}

	// SCM initialization must happen before autodiscovery so that git repository
	// directories are available for crawlers to analyse.
	{
		_, scmSpan := tracer.Start(ctx, "updatecli.init_scm")
		err = e.InitSCM()
		if err != nil {
			scmSpan.RecordError(err)
			scmSpan.SetStatus(codes.Error, err.Error())
			scmSpan.End()
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}
		scmSpan.End()
	}

	{
		_, adSpan := tracer.Start(ctx, "updatecli.autodiscovery",
			trace.WithAttributes(
				attribute.Bool("updatecli.autodiscovery.default_crawlers_enabled", defaultCrawlersEnabled),
			),
		)
		err = e.LoadAutoDiscovery(ctx, defaultCrawlersEnabled)
		if err != nil {
			adSpan.RecordError(err)
			adSpan.SetStatus(codes.Error, err.Error())
			adSpan.End()
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}
		adSpan.End()
	}

	if len(e.Pipelines) == 0 {
		err = fmt.Errorf("no valid pipeline found")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}
