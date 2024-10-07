package pipeline

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/condition"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/source"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
	"github.com/updatecli/updatecli/pkg/plugins/resources/shell"
)

/*
 */
func TestRun(t *testing.T) {

	testdata := []struct {
		conf                     config.Config
		expectedSourcesResult    map[string]string
		expectedConditionsResult map[string]string
		expectedTargetsResult    map[string]string
		expectedPipelineResult   string
	}{
		{
			conf: config.Config{
				Spec: config.Spec{
					Name:       "Test Various target scenario",
					PipelineID: "e2e/command",
					Sources: map[string]source.Config{
						"1": {
							resource.ResourceConfig{
								Kind: "shell",
								Name: "Should be succeeding",
								Spec: shell.Spec{
									Command: "echo 1.2.3",
								},
							},
						},
					},
					Conditions: map[string]condition.Config{
						"1": {
							ResourceConfig: resource.ResourceConfig{
								Kind: "shell",
								Name: "Should be succeeding",
								Spec: shell.Spec{
									Command: "true",
								},
							},
							DisableSourceInput: true,
						},
					},
					Targets: map[string]target.Config{
						"1": {
							ResourceConfig: resource.ResourceConfig{
								Kind: "shell",
								Name: "Should be succeeding",
								Spec: shell.Spec{
									Command: "true",
								},
							},
							DisableSourceInput: true,
						},
						"5": {
							ResourceConfig: resource.ResourceConfig{
								Kind: "shell",
								Name: "Should be succeeding and report change",
								Spec: shell.Spec{
									Command: "echo done",
								},
								DependsOn: []string{
									"1",
								},
							},
							DisableSourceInput: true,
						},
						"6": {
							ResourceConfig: resource.ResourceConfig{
								Kind: "shell",
								Name: "Should be skipped",
								Spec: shell.Spec{
									Command: "true",
								},
								DependsOn: []string{
									"1",
								},
							},
							DependsOnChange:    true,
							DisableSourceInput: true,
						},
						"7": {
							ResourceConfig: resource.ResourceConfig{
								Kind: "shell",
								Name: "Should be run",
								Spec: shell.Spec{
									Command: "true",
								},
								DependsOn: []string{
									"5",
								},
							},
							DependsOnChange:    true,
							DisableSourceInput: true,
						},
					},
				},
			},
			expectedSourcesResult: map[string]string{
				"1": "✔",
			},
			expectedConditionsResult: map[string]string{
				"1": "✔",
			},
			expectedTargetsResult: map[string]string{
				"1": "✔",
				"5": "-",
				"6": "-",
				"7": "-",
			},
			expectedPipelineResult: "✔",
		},
	}

	for _, data := range testdata {
		t.Run(data.conf.Spec.Name, func(t *testing.T) {
			p := Pipeline{}
			err := p.Init(&data.conf, Options{})
			require.NoError(t, err)

			err = p.Run()
			require.NoError(t, err)

			require.Equal(t, len(data.expectedSourcesResult), len(p.Sources))
			for id, result := range p.Sources {
				require.Equal(t, data.expectedSourcesResult[id], result.Result.Result)
			}
			require.Equal(t, len(data.expectedConditionsResult), len(p.Conditions))
			for id, result := range p.Conditions {
				require.Equal(t, data.expectedConditionsResult[id], result.Result.Result)
			}
			require.Equal(t, len(data.expectedTargetsResult), len(p.Targets))
			for id, result := range p.Targets {
				logrus.Errorf("Target %s", id)
				require.Equal(t, data.expectedTargetsResult[id], result.Result.Result)
			}
			require.Equal(t, data.expectedPipelineResult, p.Report.Result)
		})
	}

}
