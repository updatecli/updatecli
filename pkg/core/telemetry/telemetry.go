package telemetry

import (
	"context"
	"fmt"
	"os"
	"regexp"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
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
	exporter, err := buildExporter(ctx)
	if err != nil {
		logrus.Warnf("telemetry: failed to create span exporter (OTEL_TRACES_EXPORTER=%q), tracing disabled: %v",
			os.Getenv("OTEL_TRACES_EXPORTER"), err)
		return noopShutdown
	}
	if exporter == nil {
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

var urlUserInfoRe = regexp.MustCompile(`(https?://)([^@]+)@`)

// SanitizeError strips URL-embedded credentials from an error message
// to prevent tokens from leaking into span status descriptions.
func SanitizeError(err error) string {
	return urlUserInfoRe.ReplaceAllString(err.Error(), "${1}****:****@")
}

// RecordSpanError records a sanitized error on a span and sets its status to Error.
// Use this instead of raw span.RecordError + span.SetStatus to prevent credential leaks.
func RecordSpanError(span trace.Span, err error) {
	sanitized := SanitizeError(err)
	span.RecordError(fmt.Errorf("%s", sanitized))
	span.SetStatus(codes.Error, sanitized)
}

// buildExporter selects a span exporter based on OTEL_TRACES_EXPORTER.
// Returns nil, nil when tracing is intentionally disabled (no config present).
func buildExporter(ctx context.Context) (sdktrace.SpanExporter, error) {
	exporterName := os.Getenv("OTEL_TRACES_EXPORTER")
	hasOTLPEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "" ||
		os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT") != ""

	// Infer "otlp" when an endpoint is configured but no exporter name is set.
	if exporterName == "" && hasOTLPEndpoint {
		exporterName = "otlp"
	}

	switch exporterName {
	case "otlp":
		return otlptracegrpc.New(ctx)
	case "otlphttp":
		return otlptracehttp.New(ctx)
	case "console", "stdout":
		return stdouttrace.New()
	case "":
		// No configuration — tracing disabled.
		return nil, nil
	default:
		logrus.Warnf("telemetry: unknown OTEL_TRACES_EXPORTER=%q, tracing disabled", exporterName)
		return nil, nil
	}
}
