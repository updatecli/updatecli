package pipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/condition"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/source"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// TestReportIDIgnoresRenderedValues ensures a manifest naming its target after a source value
// keeps the same report ID whatever version that source returns.
func TestReportIDIgnoresRenderedValues(t *testing.T) {
	reportID := func(version string) string {
		p := Pipeline{}
		require.NoError(t, p.Init(&config.Config{
			Spec: config.Spec{
				Name:       "Update Golang version",
				PipelineID: "golang/version",
				Sources: map[string]source.Config{
					"go": {ResourceConfig: resource.ResourceConfig{Name: "Get latest Golang version", Kind: "golang"}},
				},
				Targets: map[string]target.Config{
					"gomod": {
						ResourceConfig: resource.ResourceConfig{
							Name: `deps(go): update Go version to {{ source "go" }}`,
							Kind: "file",
							Spec: map[string]any{"file": "go.mod", "content": `go {{ source "go" }}`},
						},
						SourceID: "go",
					},
				},
			},
		}, Options{}))

		p.Sources["go"] = source.Source{
			Result: &result.Source{Result: result.SUCCESS},
			Output: version,
		}
		require.NoError(t, p.Update())

		// Same as updateTargetResult once the target ran
		reportConfig, err := resource.GetReportConfig(p.Config.Spec.Targets["gomod"].ResourceConfig)
		require.NoError(t, err)
		p.Report.Targets["gomod"].Config = reportConfig
		require.Equal(t, "deps(go): update Go version to "+version, p.Config.Spec.Targets["gomod"].Name)

		require.NoError(t, p.Report.UpdateID())
		return p.Report.ID
	}

	assert.Equal(t, reportID("1.25.2"), reportID("1.25.3"))
}

// TestReportIDFallsBackToConfigWithoutSpec ensures a resource whose report config can't be
// computed before its spec is rendered still counts in the report ID, without its spec.
func TestReportIDFallsBackToConfigWithoutSpec(t *testing.T) {
	initPipeline := func(image, kind string) Pipeline {
		p := Pipeline{}
		require.NoError(t, p.Init(&config.Config{
			Spec: config.Spec{
				Name: "Update Golang image",
				Sources: map[string]source.Config{
					"go": {ResourceConfig: resource.ResourceConfig{Name: "Get latest Golang version", Kind: "golang"}},
				},
				Conditions: map[string]condition.Config{
					"image": {
						ResourceConfig: resource.ResourceConfig{
							Kind: kind,
							Spec: map[string]any{
								"image":         image,
								"architectures": `{{ source "go" }}`,
							},
						},
					},
				},
			},
		}, Options{}))
		return p
	}

	golang, alpine := initPipeline("golang", "dockerimage"), initPipeline("alpine", "dockerimage")
	other := initPipeline("golang", "dockerdigest")

	_, err := resource.GetReportConfig(golang.Config.Spec.Conditions["image"].ResourceConfig)
	require.Error(t, err, "the unrendered spec must be rejected for this test to be meaningful")

	assert.Nil(t, golang.Report.Conditions["image"].Config, "the fallback config must not be reported")

	for _, p := range []*Pipeline{&golang, &alpine, &other} {
		require.NoError(t, p.Report.UpdateID())
	}

	// The spec may contain credentials, so it doesn't count in the report ID
	assert.Equal(t, golang.Report.ID, alpine.Report.ID)
	assert.NotEqual(t, golang.Report.ID, other.Report.ID)
}
