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
		action         string
		reference      string
		rootDir        string
		expectedResult bool
	}{
		{
			name: "Test matching with only path",
			rules: MatchingRules{
				MatchingRule{
					Path: "testdata/helmrelease.yaml",
				},
			},
			filePath:       "testdata/helmrelease.yaml",
			expectedResult: true,
		},
		{
			name: "Test matching with subpath",
			rules: MatchingRules{
				MatchingRule{
					Path: "testdata/helmrelease.yaml",
				},
			},
			filePath:       "./website/testdata/helmrelease.yaml",
			expectedResult: false,
		},
		{
			name: "test matching with path and action",
			rules: MatchingRules{
				MatchingRule{
					Path: "testdata/helmrelease.yaml",
					Actions: map[string]string{
						"udash": "",
					},
				},
			},
			filePath:       "testdata/helmrelease.yaml",
			expectedResult: false,
		},
		{
			name: "test failing path but matching action",
			rules: MatchingRules{
				MatchingRule{
					Path: "testdata/helmrelease.yaml",
					Actions: map[string]string{
						"updatecli/updatecli": "",
					},
				},
			},
			filePath:       "testdata/updatecli/.github/workflows/updatecli.yaml",
			action:         "updatecli/updatecli",
			expectedResult: false,
		},
	}

	for _, d := range dataset {
		t.Run(d.name, func(t *testing.T) {
			gotResult := d.rules.isMatchingRules(
				d.rootDir,
				d.filePath,
				d.action,
				d.reference)

			assert.Equal(t, d.expectedResult, gotResult)
		})
	}
}
