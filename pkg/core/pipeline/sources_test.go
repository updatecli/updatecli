package pipeline

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/condition"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/source"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/resources/shell"
	"github.com/updatecli/updatecli/pkg/plugins/resources/shell/success/exitcode"
)

func TestRunSources(t *testing.T) {

	testdata := []struct {
		conf                   config.Config
		expectedSourcesResult  map[string]string
		expectedPipelineResult string
		expectedError          bool
	}{
		{
			conf: config.Config{
				Spec: config.Spec{
					Name: "Test a simple successful source pipeline",
					Sources: map[string]source.Config{
						"success": {
							ResourceConfig: resource.ResourceConfig{
								Kind: "shell",
								Name: "success",
								Spec: shell.Spec{
									Command: "true",
									ChangedIf: shell.SpecChangedIf{
										Kind: "exitcode",
										Spec: exitcode.Spec{
											Warning: 1, Success: 0, Failure: 2,
										},
									},
								},
							},
						},
					},
				},
			},
			expectedSourcesResult: map[string]string{
				"success": result.SUCCESS,
			},
			expectedPipelineResult: result.SUCCESS,
		},
		{
			conf: config.Config{
				Spec: config.Spec{
					Name: "Test a case with one source of each result type",
					Sources: map[string]source.Config{
						"success": {
							ResourceConfig: resource.ResourceConfig{
								Kind: "shell",
								Name: "success",
								Spec: shell.Spec{
									Command: "true",
								},
							},
						},
						"failed": {
							ResourceConfig: resource.ResourceConfig{
								Kind: "shell",
								Name: "failure",
								Spec: shell.Spec{
									Command: "false",
								},
							},
						},
					},
				},
			},
			expectedError: true,
			expectedSourcesResult: map[string]string{
				"success": result.SUCCESS,
				"failed":  result.FAILURE,
			},
			expectedPipelineResult: result.FAILURE,
		},
		{
			conf: config.Config{
				Spec: config.Spec{
					Name: "Test a case with a skipped source",
					Sources: map[string]source.Config{
						"success": {
							ResourceConfig: resource.ResourceConfig{
								Kind: "shell",
								Name: "success",
								Spec: shell.Spec{
									Command: "true",
								},
								DependsOn: []string{"condition#skip"},
							},
						},
					},
					Conditions: map[string]condition.Config{
						"skip": {
							ResourceConfig: resource.ResourceConfig{
								Kind: "shell",
								Name: "skip",
								Spec: shell.Spec{
									Command: "false",
								},
							},
						},
					},
				},
			},
			expectedError: false,
			expectedSourcesResult: map[string]string{
				"success": result.SKIPPED,
			},
			// As expected, the source is skipped because the condition is not met
			// so the pipeline result is considered as success
			expectedPipelineResult: result.SUCCESS,
		},
		{
			conf: config.Config{
				Spec: config.Spec{
					Name: "Test a case with a skipped source and warning second source",
					Sources: map[string]source.Config{
						"success": {
							ResourceConfig: resource.ResourceConfig{
								Kind: "shell",
								Name: "success",
								Spec: shell.Spec{
									Command: "true",
								},
								DependsOn: []string{"condition#skip"},
							},
						},
						"failed": {
							ResourceConfig: resource.ResourceConfig{
								Kind: "shell",
								Name: "failure",
								Spec: shell.Spec{
									Command: "false",
								},
							},
						},
					},
					Conditions: map[string]condition.Config{
						"skip": {
							ResourceConfig: resource.ResourceConfig{
								Kind: "shell",
								Name: "skip",
								Spec: shell.Spec{
									Command: "false",
								},
							},
							DisableSourceInput: true,
						},
					},
				},
			},
			expectedError: true,
			expectedSourcesResult: map[string]string{
				"success": result.SKIPPED,
				"failed":  result.FAILURE,
			},
			expectedPipelineResult: result.FAILURE,
		},
		{
			conf: config.Config{
				Spec: config.Spec{
					Name: "Test a case with a skipped source and a succes second source",
					Sources: map[string]source.Config{
						"success": {
							ResourceConfig: resource.ResourceConfig{
								Kind: "shell",
								Name: "success",
								Spec: shell.Spec{
									Command: "true",
								},
								DependsOn: []string{"condition#skip"},
							},
						},
						"successbis": {
							ResourceConfig: resource.ResourceConfig{
								Kind: "shell",
								Name: "successbis",
								Spec: shell.Spec{
									Command: "true",
								},
							},
						},
					},
					Conditions: map[string]condition.Config{
						"skip": {
							ResourceConfig: resource.ResourceConfig{
								Kind: "shell",
								Name: "skip",
								Spec: shell.Spec{
									Command: "false",
								},
							},
							DisableSourceInput: true,
						},
					},
				},
			},
			expectedError: true,
			expectedSourcesResult: map[string]string{
				"success":    result.SKIPPED,
				"successbis": result.SUCCESS,
			},
			expectedPipelineResult: result.SUCCESS,
		},
	}

	for _, data := range testdata {
		t.Run(data.conf.Spec.Name, func(t *testing.T) {
			p := Pipeline{}
			err := p.Init(&data.conf, Options{})
			require.NoError(t, err)

			err = p.Run()
			if !data.expectedError {
				require.NoError(t, err)
			}

			require.Equal(t, len(data.expectedSourcesResult), len(p.Sources))
			for id, result := range p.Sources {
				require.Equal(t, data.expectedSourcesResult[id], result.Result.Result)
			}
			require.Equal(t, data.expectedPipelineResult, p.Report.Result)
		})
	}

}
