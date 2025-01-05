package pipeline

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/condition"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
	"github.com/updatecli/updatecli/pkg/plugins/resources/shell"
	"github.com/updatecli/updatecli/pkg/plugins/resources/shell/success/exitcode"
)

func TestRunTarget(t *testing.T) {

	testdata := []struct {
		conf                   config.Config
		expectedTargetsResult  map[string]string
		expectedPipelineResult string
	}{
		{
			conf: config.Config{
				Spec: config.Spec{
					Name: "Test a simple successful target pipeline",
					Targets: map[string]target.Config{
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
							DisableSourceInput: true,
						},
					},
				},
			},
			expectedTargetsResult: map[string]string{
				"success": "✔",
			},
			expectedPipelineResult: "✔",
		},
		{
			conf: config.Config{
				Spec: config.Spec{
					Name: "Test a case with one target of each result type",
					Targets: map[string]target.Config{
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
							DisableSourceInput: true,
						},
						"changed": {
							ResourceConfig: resource.ResourceConfig{
								Kind: "shell",
								Name: "failure",
								Spec: shell.Spec{
									Command: "false",
									ChangedIf: shell.SpecChangedIf{
										Kind: "exitcode",
										Spec: exitcode.Spec{
											Warning: 1, Success: 0, Failure: 2,
										},
									},
								},
							},
							DisableSourceInput: true,
						},
					},
				},
			},
			expectedTargetsResult: map[string]string{
				"success": "✔",
				"changed": "⚠",
			},
			expectedPipelineResult: "⚠",
		},
		{
			conf: config.Config{
				Spec: config.Spec{
					Name: "Test a case with one successful dependsonchange",
					Targets: map[string]target.Config{
						"success": {
							ResourceConfig: resource.ResourceConfig{
								Kind:      "shell",
								Name:      "success",
								DependsOn: []string{"changed"},
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
							DisableSourceInput: true,
							DependsOnChange:    true,
						},
						"changed": {
							ResourceConfig: resource.ResourceConfig{
								Kind: "shell",
								Name: "failure",
								Spec: shell.Spec{
									Command: "false",
									ChangedIf: shell.SpecChangedIf{
										Kind: "exitcode",
										Spec: exitcode.Spec{
											Warning: 1, Success: 0, Failure: 2,
										},
									},
								},
							},
							DisableSourceInput: true,
						},
					},
				},
			},
			expectedTargetsResult: map[string]string{
				"success": "✔",
				"changed": "⚠",
			},
			expectedPipelineResult: "⚠",
		},
		{
			conf: config.Config{
				Spec: config.Spec{
					Name: "Test a case a skipped targeted due to unchanged dependsonchange",
					Targets: map[string]target.Config{
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
							DisableSourceInput: true,
						},
						"changed": {
							ResourceConfig: resource.ResourceConfig{
								Kind:      "shell",
								Name:      "failure",
								DependsOn: []string{"success"},
								Spec: shell.Spec{
									Command: "false",
									ChangedIf: shell.SpecChangedIf{
										Kind: "exitcode",
										Spec: exitcode.Spec{
											Warning: 1, Success: 0, Failure: 2,
										},
									},
								},
							},
							DisableSourceInput: true,
							DependsOnChange:    true,
						},
					},
				},
			},
			expectedTargetsResult: map[string]string{
				"success": "✔",
				"changed": "-",
			},
			expectedPipelineResult: "✔",
		},
		{
			conf: config.Config{
				Spec: config.Spec{
					Name: "Test a case a skipped targeted due to unchanged dependsonchange with and operator",
					Targets: map[string]target.Config{
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
							DisableSourceInput: true,
						},
						"changed": {
							ResourceConfig: resource.ResourceConfig{
								Kind:      "shell",
								Name:      "failure",
								DependsOn: []string{"success:and", "changed-2"},
								Spec: shell.Spec{
									Command: "false",
									ChangedIf: shell.SpecChangedIf{
										Kind: "exitcode",
										Spec: exitcode.Spec{
											Warning: 1, Success: 0, Failure: 2,
										},
									},
								},
							},
							DisableSourceInput: true,
							DependsOnChange:    true,
						},
						"changed-2": {
							ResourceConfig: resource.ResourceConfig{
								Kind: "shell",
								Name: "failure",
								Spec: shell.Spec{
									Command: "false",
									ChangedIf: shell.SpecChangedIf{
										Kind: "exitcode",
										Spec: exitcode.Spec{
											Warning: 1, Success: 0, Failure: 2,
										},
									},
								},
							},
							DisableSourceInput: true,
						},
					},
				},
			},
			expectedTargetsResult: map[string]string{
				"success":   "✔",
				"changed":   "-",
				"changed-2": "⚠",
			},
			expectedPipelineResult: "⚠",
		},
		{
			conf: config.Config{
				Spec: config.Spec{
					Name: "Test a case a skipped targeted due to unchanged dependsonchange with OR operator",
					Targets: map[string]target.Config{
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
							DisableSourceInput: true,
						},
						"changed": {
							ResourceConfig: resource.ResourceConfig{
								Kind:      "shell",
								Name:      "failure",
								DependsOn: []string{"success:or", "changed:2:or"},
								Spec: shell.Spec{
									Command: "false",
									ChangedIf: shell.SpecChangedIf{
										Kind: "exitcode",
										Spec: exitcode.Spec{
											Warning: 1, Success: 0, Failure: 2,
										},
									},
								},
							},
							DisableSourceInput: true,
							DependsOnChange:    true,
						},
						"changed:2": {
							ResourceConfig: resource.ResourceConfig{
								Kind: "shell",
								Name: "failure",
								Spec: shell.Spec{
									Command: "false",
									ChangedIf: shell.SpecChangedIf{
										Kind: "exitcode",
										Spec: exitcode.Spec{
											Warning: 1, Success: 0, Failure: 2,
										},
									},
								},
							},
							DisableSourceInput: true,
						},
					},
				},
			},
			expectedTargetsResult: map[string]string{
				"success":   "✔",
				"changed":   "⚠",
				"changed:2": "⚠",
			},
			expectedPipelineResult: "⚠",
		},
		{
			conf: config.Config{
				Spec: config.Spec{
					Name: "Test a case where all skipped targets results in a skipped pipeline",
					Conditions: map[string]condition.Config{
						"failing": {
							ResourceConfig: resource.ResourceConfig{
								Kind: "shell",
								Name: "success",
								Spec: shell.Spec{
									Command: "false",
									ChangedIf: shell.SpecChangedIf{
										Kind: "exitcode",
										Spec: exitcode.Spec{
											Warning: 1, Success: 0, Failure: 2,
										},
									},
								},
							},
							DisableSourceInput: true,
						},
					},
					Targets: map[string]target.Config{
						"skipped-1": {
							ResourceConfig: resource.ResourceConfig{
								Kind:      "shell",
								Name:      "failure",
								DependsOn: []string{"condition#failing"},
								Spec: shell.Spec{
									Command: "false",
									ChangedIf: shell.SpecChangedIf{
										Kind: "exitcode",
										Spec: exitcode.Spec{
											Warning: 1, Success: 0, Failure: 2,
										},
									},
								},
							},
							DisableSourceInput: true,
							DependsOnChange:    true,
						},
						"skipped-2": {
							ResourceConfig: resource.ResourceConfig{
								Kind:      "shell",
								Name:      "failure",
								DependsOn: []string{"condition#failing"},
								Spec: shell.Spec{
									Command: "false",
									ChangedIf: shell.SpecChangedIf{
										Kind: "exitcode",
										Spec: exitcode.Spec{
											Warning: 1, Success: 0, Failure: 2,
										},
									},
								},
							},
							DisableSourceInput: true,
							DependsOnChange:    true,
						},
					},
				},
			},
			expectedTargetsResult: map[string]string{
				"skipped-1": "-",
				"skipped-2": "-",
			},
			expectedPipelineResult: "-",
		},
	}

	for _, data := range testdata {
		t.Run(data.conf.Spec.Name, func(t *testing.T) {
			p := Pipeline{}
			err := p.Init(&data.conf, Options{})
			require.NoError(t, err)

			err = p.Run()
			require.NoError(t, err)

			require.Equal(t, len(data.expectedTargetsResult), len(p.Targets))
			for id, result := range p.Targets {
				require.Equal(t, data.expectedTargetsResult[id], result.Result.Result)
			}
			require.Equal(t, data.expectedPipelineResult, p.Report.Result)
		})
	}

}
