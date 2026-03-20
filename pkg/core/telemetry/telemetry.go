package telemetry

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// noopShutdown is returned on init failure so callers can always defer shutdown safely.
var noopShutdown = func(context.Context) error { return nil }

// Init configures the global OpenTelemetry tracer provider.
// The exporter is selected via OTEL_TRACES_EXPORTER / OTEL_EXPORTER_OTLP_ENDPOINT.
// On any error, tracing is silently disabled — it must never block updatecli.
func Init(ctx context.Context, serviceName, serviceVersion string) func(context.Context) error {
	hasConfig := os.Getenv("OTEL_TRACES_EXPORTER") != "" ||
		os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "" ||
		os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT") != ""
	if !hasConfig {
		return noopShutdown
	}

	exporter, err := autoexport.NewSpanExporter(ctx)
	if err != nil {
		logrus.Warnf("telemetry: failed to create span exporter (OTEL_TRACES_EXPORTER=%q), tracing disabled: %v",
			os.Getenv("OTEL_TRACES_EXPORTER"), err)
		return noopShutdown
	}

	res, err := resource.New(ctx,
		resource.WithSchemaURL(semconv.SchemaURL),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(serviceVersion),
		),
	)
	if err != nil && res == nil {
		logrus.Warnf("telemetry: failed to create resource, tracing disabled: %v", err)
		return noopShutdown
	}
	if err != nil {
		logrus.Warnf("telemetry: partial resource created: %v", err)
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(provider)

	return func(ctx context.Context) error {
		return provider.Shutdown(ctx)
	}
}

// Tracer returns a named tracer from the global provider.
func Tracer(name string) trace.Tracer {
	return otel.Tracer(name)
}
