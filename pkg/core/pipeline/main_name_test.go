package pipeline

import (
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
