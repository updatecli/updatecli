package githubaction

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsMatchingRule(t *testing.T) {

	dataset := []struct {
		rules          MatchingRules
		name           string
		filePath       string
		repository     string
		reference      string
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
					Actions: map[string]string{
						"udash": "",
					},
				},
			},
			filePath:       "testdata/helmrelease.yaml",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "testdata/helmrelease.yaml",
					Actions: map[string]string{
						"updatecli/updatecli": "",
					},
				},
			},
			filePath:       "testdata/updatecli/.github/workflows/updatecli.yaml",
			repository:     "updatecli/updatecli",
			expectedResult: true,
		},
	}

	for _, d := range dataset {
		t.Run(d.name, func(t *testing.T) {
			gotResult := d.rules.isMatchingRules(
				d.rootDir,
				d.filePath,
				d.repository,
				d.reference)

			assert.Equal(t, d.expectedResult, gotResult)
		})
	}
}
