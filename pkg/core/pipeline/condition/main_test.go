package condition

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/resources/file"
)

func TestRun(t *testing.T) {

	tests := []struct {
		name           string
		condition      Condition
		expectedResult bool
	}{
		{
			name: "Passing case with successful matching pattern",
			condition: Condition{
				Config: Config{
					ResourceConfig: resource.ResourceConfig{
						Kind: "file",
						Spec: file.Spec{
							File:         "main.go",
							MatchPattern: "Run",
						},
					},
					DisableSourceInput: true,
				},
			},
			expectedResult: true,
		},
		{
			name: "None passing case with successful matching pattern",
			condition: Condition{
				Config: Config{
					FailWhen: true,
					ResourceConfig: resource.ResourceConfig{
						Kind: "file",
						Spec: file.Spec{
							File:         "main.go",
							MatchPattern: "Run",
						},
					},
					DisableSourceInput: true,
				},
			},
			expectedResult: false,
		},
		{
			name: "None Passing case with none matching pattern",
			condition: Condition{
				Config: Config{
					ResourceConfig: resource.ResourceConfig{
						Kind: "file",
						Spec: file.Spec{
							File:         "main.go",
							MatchPattern: "TestDoNotExist",
						},
					},
					DisableSourceInput: true,
				},
			},
			expectedResult: false,
		},
		{
			name: "Passing case with none matching pattern",
			condition: Condition{
				Config: Config{
					FailWhen: true,
					ResourceConfig: resource.ResourceConfig{
						Kind: "file",
						Spec: file.Spec{
							File:         "main.go",
							MatchPattern: "TestDoNotExist",
						},
					},
					DisableSourceInput: true,
				},
			},
			expectedResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Normally the result is initialized by the pipeline, we do it here for testing purpose
			tt.condition.Result = &result.Condition{}

			gotErr := tt.condition.Run("")
			require.NoError(t, gotErr)

			assert.Equal(t, tt.expectedResult, tt.condition.Result.Pass)
		})
	}
}
