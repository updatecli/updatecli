package dockercompose

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsMatchingRules(t *testing.T) {
	testdata := []struct {
		name           string
		rules          MatchingRules
		rootDir        string
		filePath       string
		service        string
		expectedResult bool
	}{
		{
			name: "Scenario 1 - matching 1 rule",
			rules: MatchingRules{
				MatchingRule{
					Path: "docker-compose.yaml",
				},
			},
			filePath:       "docker-compose.yaml",
			expectedResult: true,
		},
		{
			name: "Scenario 2 - matching all rule",
			rules: MatchingRules{
				MatchingRule{
					Path: "docker-compose.yaml",
					Services: []string{
						"mongodb",
					},
				},
			},
			filePath:       "docker-compose.yaml",
			service:        "mongodb",
			expectedResult: true,
		},
		{
			name: "Scenario 3 - not matching all rules",
			rules: MatchingRules{
				MatchingRule{
					Path: "docker-compose.2.yaml",
					Services: []string{
						"mongodb",
					},
				},
			},
			filePath:       "docker-compose.yaml",
			service:        "mongodb",
			expectedResult: false,
		},
	}

	for _, tt := range testdata {

		t.Run(tt.name, func(t *testing.T) {
			gotResult := tt.rules.isMatchingRule(tt.rootDir, tt.filePath, tt.service)

			assert.Equal(t, tt.expectedResult, gotResult)

		})
	}

}
