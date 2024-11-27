package pipeline

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
	"github.com/updatecli/updatecli/pkg/core/result"
)

/*
 */
func TestRun(t *testing.T) {

	testdata := []struct {
		confPath                    string
		expectedSourcesResult       map[string]string
		expectedSourcesInformations map[string][]result.SourceInformation
		expectedConditionsResult    map[string]string
		expectedTargetsResult       map[string]string
		expectedPipelineResult      string
		targetDryRun                bool
	}{
		{
			confPath: "../../../e2e/updatecli.d/success.d/command.yaml",
			expectedSourcesResult: map[string]string{
				"1": "✔",
			},
			expectedConditionsResult: map[string]string{
				"1": "✔",
			},
			expectedTargetsResult: map[string]string{
				"1": "✔",
				"5": "⚠",
				"6": "-",
				"7": "✔",
			},
			expectedPipelineResult: "⚠",
			targetDryRun:           true,
		}, {
			confPath: "../../../e2e/updatecli.d/success.d/dependsonchangeor.yaml",
			expectedSourcesResult: map[string]string{
				"1": "✔",
			},
			expectedConditionsResult: map[string]string{
				"1": "✔",
			},
			expectedTargetsResult: map[string]string{
				"1": "✔",
				"2": "✔",
				"3": "-",
			},
			expectedPipelineResult: "✔",
			targetDryRun:           true,
		}, {
			confPath: "../../../e2e/updatecli.d/success.d/loops.yaml",
			expectedSourcesResult: map[string]string{
				"create_dummy_json":    result.SUCCESS,
				"dummy_json_filepath":  result.SUCCESS,
				"read_dummy_json_file": result.SUCCESS,
			},
			expectedSourcesInformations: map[string][]result.SourceInformation{
				"create_dummy_json": {{
					Value: "{\"o\": [\"0\", \"1\", \"2\"]}",
				}},
				"read_dummy_json_file": {{
					Key:   "0",
					Value: "0",
				}, {
					Key:   "1",
					Value: "1",
				}, {
					Key:   "2",
					Value: "2",
				}},
			},
			expectedConditionsResult: map[string]string{},
			expectedTargetsResult: map[string]string{
				"create_dummy_json_file": result.ATTENTION,
				"delete_dummy_json_file": result.SUCCESS,
				"echo_json_value":        result.ATTENTION,
			},
			expectedPipelineResult: result.ATTENTION,
		},
	}

	for _, data := range testdata {
		c, err := config.New(config.Option{
			ManifestFile: data.confPath,
		})
		require.NoError(t, err)
		p := Pipeline{}
		_ = p.Init(&c[0], Options{
			Target: target.Options{
				DryRun: data.targetDryRun,
			},
		})
		t.Run(p.Config.Spec.Name, func(t *testing.T) {
			err := p.Run()
			if err != nil {
				logrus.Errorf("Got error running test: %s", err)
			}
			require.NoError(t, err)

			require.Equal(t, len(data.expectedSourcesResult), len(p.Sources))
			for id, result := range p.Sources {
				require.Equal(t, data.expectedSourcesResult[id], result.Result.Result)
			}
			for id, expectedSourceInformations := range data.expectedSourcesInformations {
				require.Equal(t, expectedSourceInformations, p.Sources[id].Result.Information)
			}
			require.Equal(t, len(data.expectedConditionsResult), len(p.Conditions))
			for id, result := range p.Conditions {
				require.Equal(t, data.expectedConditionsResult[id], result.Result.Result)
			}
			require.Equal(t, len(data.expectedTargetsResult), len(p.Targets))
			for id, result := range p.Targets {
				require.Equal(t, data.expectedTargetsResult[id], result.Result.Result)
			}
			require.Equal(t, data.expectedPipelineResult, p.Report.Result)
		})
	}

}
