package pipeline

import (
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/source"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func TestUpdateRefreshesName(t *testing.T) {
	testdata := []struct {
		name         string
		sourceResult string
		expectedName string
	}{
		{
			name:         "source succeeded",
			sourceResult: result.SUCCESS,
			expectedName: "deps(go): bump module golang.org/x/net to v0.38.0",
		},
		{
			name:         "source skipped keeps the placeholder",
			sourceResult: result.SKIPPED,
			expectedName: `deps(go): bump module golang.org/x/net to {{ source "fixed" }}`,
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			rawName := `deps(go): bump module golang.org/x/net to {{ source "fixed" }}`

			p := Pipeline{
				ID: "pipeline-id",
				Config: &config.Config{
					Spec: config.Spec{
						Name:       rawName,
						PipelineID: "pipeline-id",
					},
				},
				Sources: map[string]source.Source{
					"fixed": {
						Result: &result.Source{Result: tt.sourceResult},
						Output: "v0.38.0",
					},
				},
			}
			p.Report.Name = rawName

			require.NoError(t, p.Update())

			assert.Equal(t, tt.expectedName, p.Name)
			assert.Equal(t, tt.expectedName, p.Report.Name)
			assert.Equal(t, "pipeline-id", p.ID)
			assert.Equal(t, "pipeline-id", p.Config.Spec.PipelineID)
		})
	}
}

// TestActionReportIDIgnoresRenderedName ensures the action report ID doesn't change when the pipeline name
// is rendered, so pullrequest reports created before and after a version change are merged together.
func TestActionReportIDIgnoresRenderedName(t *testing.T) {
	rawName := `deps(go): bump module golang.org/x/net to {{ source "fixed" }}`

	p := Pipeline{}
	require.NoError(t, p.Init(&config.Config{
		Spec: config.Spec{
			Name:       rawName,
			PipelineID: "pipeline-id",
		},
	}, Options{}))

	p.Sources["fixed"] = source.Source{
		Result: &result.Source{Result: result.SUCCESS},
		Output: "v0.38.0",
	}
	require.NoError(t, p.Update())

	require.Equal(t, "deps(go): bump module golang.org/x/net to v0.38.0", p.Name)
	assert.Equal(t, fmt.Sprintf("%x", sha256.Sum256([]byte(rawName))), p.actionReportID())
}

func TestActionReportIDWithoutInit(t *testing.T) {
	p := Pipeline{Name: "pipeline"}

	assert.Equal(t, fmt.Sprintf("%x", sha256.Sum256([]byte("pipeline"))), p.actionReportID())
}

// TestInitNameFallsBackToTitle ensures a specification defining only a title exposes that title
// as both the pipeline name and the report name, including before the configuration is rendered.
func TestInitNameFallsBackToTitle(t *testing.T) {
	p := Pipeline{}
	require.NoError(t, p.Init(&config.Config{
		Spec: config.Spec{
			Title:      "Bump golang.org/x/net",
			PipelineID: "pipeline-id",
		},
	}, Options{}))

	assert.Equal(t, "Bump golang.org/x/net", p.Name)
	assert.Equal(t, "Bump golang.org/x/net", p.Report.Name)

	require.NoError(t, p.Update())

	assert.Equal(t, "Bump golang.org/x/net", p.Name)
	assert.Equal(t, "Bump golang.org/x/net", p.Report.Name)
}
