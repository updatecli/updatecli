package pipeline

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
)

/*
 */
func TestRun(t *testing.T) {

	testdata := []struct {
		confPath                 string
		expectedSourcesResult    map[string]string
		expectedConditionsResult map[string]string
		expectedTargetsResult    map[string]string
		expectedPipelineResult   string
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
		},
	}

	for _, data := range testdata {
		c, _ := config.New(config.Option{
			ManifestFile: data.confPath,
		}, []string{})
		p := Pipeline{}
		_ = p.Init(&c[0], Options{
			Target: target.Options{
				DryRun: true,
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
