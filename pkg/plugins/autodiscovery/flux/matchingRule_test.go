package flux

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
					Path: "testdata/helmrelease.yaml",
				},
			},
			filePath:       "testdata/helmrelease.yaml",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "testdata/helmrelease.yaml",
				},
			},
			filePath:       "./website/testdata/helmrelease.yaml",
			expectedResult: false,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "testdata/helmrelease.yaml",
					Charts: map[string]string{
						"udash": "",
					},
				},
			},
			filePath:       "testdata/helmrelease.yaml",
			chartName:      "udash",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "testdata/helmrelease.yaml",
					Repositories: []string{
						"https://helm.example.com",
					},
				},
			},
			filePath:       "testdata/helmrelease.yaml",
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
