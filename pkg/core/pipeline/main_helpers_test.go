package pipeline

import (
	"bytes"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/pipeline/condition"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/pipeline/source"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
	"github.com/updatecli/updatecli/pkg/core/result"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

// newRecordingProvider returns a TracerProvider backed by an in-memory
// SpanRecorder, making emitted spans available for assertion.
func newRecordingProvider() (*sdktrace.TracerProvider, *tracetest.SpanRecorder) {
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	return provider, recorder
}

// startTestSpan starts a span using the given provider and returns it.
func startTestSpan(t *testing.T, provider *sdktrace.TracerProvider) trace.Span {
	t.Helper()
	_, span := provider.Tracer("test").Start(t.Context(), "test-span")
	return span
}

// endedStub converts the first ended span from the recorder into a SpanStub,
// which exposes Attributes, Events, etc. as concrete fields.
func endedStub(recorder *tracetest.SpanRecorder) tracetest.SpanStub {
	spans := recorder.Ended()
	return tracetest.SpanStubFromReadOnlySpan(spans[0])
}

// spanStrAttrs returns a map of key→string value for all string-typed attributes.
func spanStrAttrs(stub tracetest.SpanStub) map[string]string {
	attrs := make(map[string]string, len(stub.Attributes))
	for _, kv := range stub.Attributes {
		if kv.Value.Type().String() == "STRING" {
			attrs[string(kv.Key)] = kv.Value.AsString()
		}
	}
	return attrs
}

// spanBoolAttrs returns a map of key→bool value for all bool-typed attributes.
func spanBoolAttrs(stub tracetest.SpanStub) map[string]bool {
	attrs := make(map[string]bool)
	for _, kv := range stub.Attributes {
		if kv.Value.Type().String() == "BOOL" {
			attrs[string(kv.Key)] = kv.Value.AsBool()
		}
	}
	return attrs
}

// spanSliceAttrs returns a map of key→[]string for all string-slice-typed attributes.
func spanSliceAttrs(stub tracetest.SpanStub) map[string][]string {
	attrs := make(map[string][]string)
	for _, kv := range stub.Attributes {
		if kv.Value.Type().String() == "STRINGSLICE" {
			attrs[string(kv.Key)] = kv.Value.AsStringSlice()
		}
	}
	return attrs
}

// mockScmHandler is a minimal ScmHandler returning a fixed summary string.
type mockScmHandler struct {
	scm.ScmHandler
	summary string
}

func (m *mockScmHandler) Summary() string { return m.summary }

// pipelineWithSCM builds a Pipeline with a single SCM keyed by scmID.
func pipelineWithSCM(scmID, summary string) *Pipeline {
	return &Pipeline{
		SCMs: map[string]scm.Scm{
			scmID: {Handler: &mockScmHandler{summary: summary}},
		},
		Sources:    make(map[string]source.Source),
		Conditions: make(map[string]condition.Condition),
		Targets:    make(map[string]target.Target),
	}
}

// ---- logRepositoryHeader ----

func TestLogRepositoryHeader_Source(t *testing.T) {
	const scmID = "my-scm"
	p := pipelineWithSCM(scmID, "github.com/org/repo")
	p.Sources["src1"] = source.Source{
		Config: source.Config{ResourceConfig: resource.ResourceConfig{SCMID: scmID}},
	}

	var buf bytes.Buffer
	original := logrus.StandardLogger().Out
	logrus.SetOutput(&buf)
	t.Cleanup(func() { logrus.SetOutput(original) })

	p.logRepositoryHeader("source#src1", sourceCategory)

	assert.Contains(t, buf.String(), "github.com/org/repo")
}

func TestLogRepositoryHeader_Condition(t *testing.T) {
	const scmID = "my-scm"
	p := pipelineWithSCM(scmID, "github.com/org/repo")
	p.Conditions["cond1"] = condition.Condition{
		Config: condition.Config{ResourceConfig: resource.ResourceConfig{SCMID: scmID}},
	}

	var buf bytes.Buffer
	original := logrus.StandardLogger().Out
	logrus.SetOutput(&buf)
	t.Cleanup(func() { logrus.SetOutput(original) })

	p.logRepositoryHeader("condition#cond1", conditionCategory)

	assert.Contains(t, buf.String(), "github.com/org/repo")
}

func TestLogRepositoryHeader_Target(t *testing.T) {
	const scmID = "my-scm"
	p := pipelineWithSCM(scmID, "github.com/org/repo")
	p.Targets["tgt1"] = target.Target{
		Config: target.Config{ResourceConfig: resource.ResourceConfig{SCMID: scmID}},
	}

	var buf bytes.Buffer
	original := logrus.StandardLogger().Out
	logrus.SetOutput(&buf)
	t.Cleanup(func() { logrus.SetOutput(original) })

	p.logRepositoryHeader("target#tgt1", targetCategory)

	assert.Contains(t, buf.String(), "github.com/org/repo")
}

func TestLogRepositoryHeader_NoSCM_NoOutput(t *testing.T) {
	// Resource has no SCMID — nothing should be logged.
	p := &Pipeline{
		SCMs: make(map[string]scm.Scm),
		Sources: map[string]source.Source{
			"src1": {Config: source.Config{}},
		},
		Conditions: make(map[string]condition.Condition),
		Targets:    make(map[string]target.Target),
	}

	var buf bytes.Buffer
	original := logrus.StandardLogger().Out
	logrus.SetOutput(&buf)
	t.Cleanup(func() { logrus.SetOutput(original) })

	p.logRepositoryHeader("source#src1", sourceCategory)

	assert.Empty(t, buf.String())
}

// ---- recordSourceSpan ----

func TestRecordSourceSpan_SetsAttributes(t *testing.T) {
	provider, recorder := newRecordingProvider()
	p := &Pipeline{
		Sources: map[string]source.Source{
			"s1": {
				Config: source.Config{ResourceConfig: resource.ResourceConfig{
					Name: "My Source",
					Kind: "shell",
				}},
				Result: &result.Source{Description: "found version 1.2.3"},
			},
		},
	}

	span := startTestSpan(t, provider)
	p.recordSourceSpan(span, "s1", result.SUCCESS)
	span.End()

	require.Len(t, recorder.Ended(), 1)

	stub := endedStub(recorder)
	strAttrs := spanStrAttrs(stub)
	assert.Equal(t, "My Source", strAttrs["updatecli.resource.name"])
	assert.Equal(t, "shell", strAttrs["updatecli.resource.kind"])
	assert.Equal(t, result.SUCCESS, strAttrs["updatecli.resource.result"])
	assert.Equal(t, "found version 1.2.3", strAttrs["updatecli.resource.description"])
}

func TestRecordSourceSpan_EmptyDescription_NoAttribute(t *testing.T) {
	provider, recorder := newRecordingProvider()
	p := &Pipeline{
		Sources: map[string]source.Source{
			"s1": {
				Config: source.Config{ResourceConfig: resource.ResourceConfig{
					Name: "Minimal Source",
					Kind: "shell",
				}},
				Result: &result.Source{Description: ""},
			},
		},
	}

	span := startTestSpan(t, provider)
	p.recordSourceSpan(span, "s1", result.SUCCESS)
	span.End()

	strAttrs := spanStrAttrs(endedStub(recorder))
	_, hasDescription := strAttrs["updatecli.resource.description"]
	assert.False(t, hasDescription, "description attribute should be absent when empty")
}

func TestRecordSourceSpan_UnknownID_IsNoOp(t *testing.T) {
	provider, recorder := newRecordingProvider()
	p := &Pipeline{Sources: make(map[string]source.Source)}

	span := startTestSpan(t, provider)
	p.recordSourceSpan(span, "nonexistent", result.SUCCESS)
	span.End()

	assert.Empty(t, endedStub(recorder).Attributes)
}

// ---- recordConditionSpan ----

func TestRecordConditionSpan_SetsAttributes(t *testing.T) {
	provider, recorder := newRecordingProvider()
	p := &Pipeline{
		Conditions: map[string]condition.Condition{
			"c1": {
				Config: condition.Config{ResourceConfig: resource.ResourceConfig{
					Name: "My Condition",
					Kind: "shell",
				}},
				Result: &result.Condition{
					Pass:        true,
					SourceID:    "src1",
					Description: "condition met",
				},
			},
		},
	}

	span := startTestSpan(t, provider)
	p.recordConditionSpan(span, "c1", result.SUCCESS)
	span.End()

	stub := endedStub(recorder)
	strAttrs := spanStrAttrs(stub)
	boolAttrs := spanBoolAttrs(stub)

	assert.Equal(t, "My Condition", strAttrs["updatecli.resource.name"])
	assert.Equal(t, "shell", strAttrs["updatecli.resource.kind"])
	assert.Equal(t, result.SUCCESS, strAttrs["updatecli.resource.result"])
	assert.Equal(t, "src1", strAttrs["updatecli.condition.source_id"])
	assert.Equal(t, "condition met", strAttrs["updatecli.resource.description"])
	assert.True(t, boolAttrs["updatecli.condition.pass"])
}

func TestRecordConditionSpan_EmptyDescription_NoAttribute(t *testing.T) {
	provider, recorder := newRecordingProvider()
	p := &Pipeline{
		Conditions: map[string]condition.Condition{
			"c1": {
				Config: condition.Config{ResourceConfig: resource.ResourceConfig{
					Name: "Bare Condition",
					Kind: "shell",
				}},
				Result: &result.Condition{Description: ""},
			},
		},
	}

	span := startTestSpan(t, provider)
	p.recordConditionSpan(span, "c1", result.FAILURE)
	span.End()

	strAttrs := spanStrAttrs(endedStub(recorder))
	_, hasDescription := strAttrs["updatecli.resource.description"]
	assert.False(t, hasDescription)
}

func TestRecordConditionSpan_UnknownID_IsNoOp(t *testing.T) {
	provider, recorder := newRecordingProvider()
	p := &Pipeline{Conditions: make(map[string]condition.Condition)}

	span := startTestSpan(t, provider)
	p.recordConditionSpan(span, "nonexistent", result.SUCCESS)
	span.End()

	assert.Empty(t, endedStub(recorder).Attributes)
}

// ---- recordTargetSpan ----

func TestRecordTargetSpan_SetsAttributes(t *testing.T) {
	provider, recorder := newRecordingProvider()
	p := &Pipeline{
		Targets: map[string]target.Target{
			"t1": {
				Config: target.Config{ResourceConfig: resource.ResourceConfig{
					Name: "My Target",
					Kind: "shell",
				}},
				Result: &result.Target{
					DryRun:      true,
					SourceID:    "src1",
					Description: "updated dependency",
					Files:       []string{"go.mod", "go.sum"},
				},
			},
		},
	}

	span := startTestSpan(t, provider)
	p.recordTargetSpan(span, "t1", result.ATTENTION, true)
	span.End()

	stub := endedStub(recorder)
	strAttrs := spanStrAttrs(stub)
	boolAttrs := spanBoolAttrs(stub)
	sliceAttrs := spanSliceAttrs(stub)

	assert.Equal(t, "My Target", strAttrs["updatecli.resource.name"])
	assert.Equal(t, "shell", strAttrs["updatecli.resource.kind"])
	assert.Equal(t, result.ATTENTION, strAttrs["updatecli.resource.result"])
	assert.Equal(t, "updated dependency", strAttrs["updatecli.resource.description"])
	assert.Equal(t, "src1", strAttrs["updatecli.target.source_id"])
	assert.True(t, boolAttrs["updatecli.target.changed"])
	assert.True(t, boolAttrs["updatecli.target.dry_run"])
	assert.Equal(t, []string{"go.mod", "go.sum"}, sliceAttrs["updatecli.target.files"])

	require.Len(t, stub.Events, 1)
	assert.Equal(t, "target.changed", stub.Events[0].Name)
}

func TestRecordTargetSpan_NotChanged_NoEvent(t *testing.T) {
	provider, recorder := newRecordingProvider()
	p := &Pipeline{
		Targets: map[string]target.Target{
			"t1": {
				Config: target.Config{ResourceConfig: resource.ResourceConfig{
					Name: "No Change Target",
					Kind: "shell",
				}},
				Result: &result.Target{},
			},
		},
	}

	span := startTestSpan(t, provider)
	p.recordTargetSpan(span, "t1", result.SUCCESS, false)
	span.End()

	stub := endedStub(recorder)
	assert.Empty(t, stub.Events, "no event should be emitted when target did not change")

	boolAttrs := spanBoolAttrs(stub)
	assert.False(t, boolAttrs["updatecli.target.changed"])
}

func TestRecordTargetSpan_NoFiles_NoSliceAttribute(t *testing.T) {
	provider, recorder := newRecordingProvider()
	p := &Pipeline{
		Targets: map[string]target.Target{
			"t1": {
				Config: target.Config{ResourceConfig: resource.ResourceConfig{
					Name: "No Files Target",
					Kind: "shell",
				}},
				Result: &result.Target{Files: nil},
			},
		},
	}

	span := startTestSpan(t, provider)
	p.recordTargetSpan(span, "t1", result.SUCCESS, false)
	span.End()

	sliceAttrs := spanSliceAttrs(endedStub(recorder))
	_, hasFiles := sliceAttrs["updatecli.target.files"]
	assert.False(t, hasFiles, "files attribute should be absent when no files are present")
}

func TestRecordTargetSpan_UnknownID_IsNoOp(t *testing.T) {
	provider, recorder := newRecordingProvider()
	p := &Pipeline{Targets: make(map[string]target.Target)}

	span := startTestSpan(t, provider)
	p.recordTargetSpan(span, "nonexistent", result.SUCCESS, false)
	span.End()

	assert.Empty(t, endedStub(recorder).Attributes)
}
