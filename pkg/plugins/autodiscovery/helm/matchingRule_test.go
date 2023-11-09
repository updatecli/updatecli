package helm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsMatchingRule(t *testing.T) {

	dataset := []struct {
		rules             MatchingRules
		name              string
		filePath          string
		containerVersion  string
		containerName     string
		dependencyName    string
		dependencyVersion string
		rootDir           string
		expectedResult    bool
	}{
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "test/testdata-1",
				},
			},
			filePath:       "test/testdata-1",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "testdata-1",
				},
			},
			filePath:       "test/testdata-1",
			expectedResult: false,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "test/testdata-1",
					Dependencies: map[string]string{
						"jenkins": "2.234",
					},
				},
			},
			filePath:          "test/testdata-1",
			dependencyName:    "jenkins",
			dependencyVersion: "2.234",
			expectedResult:    true,
		},
	}

	for _, d := range dataset {
		t.Run(d.name, func(t *testing.T) {
			gotResult := d.rules.isMatchingRules(
				d.rootDir,
				d.filePath,
				d.dependencyName,
				d.dependencyVersion,
				d.containerName,
				d.containerVersion,
			)

			assert.Equal(t, d.expectedResult, gotResult)
		})
	}
}
