package helmfile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsMatchingRule(t *testing.T) {

	dataset := []struct {
		rules          MatchingRules
		name           string
		filePath       string
		repositoryURL  string
		chartName      string
		chartVersion   string
		rootDir        string
		expectedResult bool
	}{
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "testdata/helmfile.d/cik8s.yaml",
				},
			},
			filePath:       "testdata/helmfile.d/cik8s.yaml",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "testdata/helmfile.d/cik8s.yaml",
				},
			},
			filePath:       "./website/testdata/helmfile.d/cik8s.yaml",
			expectedResult: false,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					//Path: "testdata/helmfile.d/cik8s.yaml",
					Charts: map[string]string{
						"datadog": "",
					},
				},
			},
			filePath:       "testdata/helmfile.d/cik8s.yaml",
			chartName:      "datadog",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "testdata/helmfile.d/cik8s.yaml",
					Repositories: []string{
						"https://helm.example.com",
					},
				},
			},
			filePath:       "testdata/helmfile.d/cik8s.yaml",
			repositoryURL:  "https://helm.example.com",
			expectedResult: true,
		},
	}

	for _, d := range dataset {
		t.Run(d.name, func(t *testing.T) {
			gotResult := d.rules.isMatchingRules(
				d.rootDir,
				d.filePath,
				d.repositoryURL,
				d.chartName,
				d.chartVersion)

			assert.Equal(t, d.expectedResult, gotResult)
		})
	}
}
