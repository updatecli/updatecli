package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsMatchingRule(t *testing.T) {

	dataset := []struct {
		rules          MatchingRules
		name           string
		filePath       string
		image          string
		rootDir        string
		expectedResult bool
	}{
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "test/testdata/success/pod.yaml",
				},
			},
			filePath:       "test/testdata/success/pod.yaml",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "test/testdata/success/pod.yaml",
				},
			},
			filePath:       "wrong/test/testdata/success/pod.yaml",
			expectedResult: false,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "test/testdata/success/pod.yaml",
					Images: []string{
						"updatecli/updatecli",
					},
				},
			},
			filePath:       "test/testdata/success/pod.yaml",
			image:          "updatecli/updatecli",
			expectedResult: true,
		},
	}

	for _, d := range dataset {
		t.Run(d.name, func(t *testing.T) {
			gotResult := d.rules.isMatchingRules(
				d.rootDir,
				d.filePath,
				d.image)

			assert.Equal(t, d.expectedResult, gotResult)
		})
	}
}
