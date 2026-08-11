package engine

import (
	"errors"

	"github.com/updatecli/updatecli/pkg/core/cache"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline"
	"github.com/updatecli/updatecli/pkg/core/reports"
	"github.com/updatecli/updatecli/pkg/core/tmp"
	"go.opentelemetry.io/otel/trace"
)

var (
	// ErrNoManifestDetected is the error message returned by Updatecli if it can't find manifest
	ErrNoManifestDetected error = errors.New("no Updatecli manifest detected")
)

// Engine defined parameters for a specific engine run.
type Engine struct {
	configurations []*config.Config
	Pipelines      []*pipeline.Pipeline
	Options        Options
	Reports        reports.Reports
	tracer         trace.Tracer
	sourceCache    *cache.SourceCache
}

// SetTracer configures the tracer used for OTel instrumentation across all engine operations.
// Call this once after telemetry is initialized, before invoking Prepare or Run.
func (e *Engine) SetTracer(t trace.Tracer) {
	e.tracer = t
}

// Clean remove every traces from an updatecli run.
func (e *Engine) Clean() (err error) {
	err = tmp.Clean()
	return
}
