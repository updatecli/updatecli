package nomad

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsMatchingRules(t *testing.T) {
	testdata := []struct {
		name           string
		rules          MatchingRules
		rootDir        string
		filePath       string
		job            string
		image          string
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
			name: "Scenario 1.1 - matching 1 rule",
			rules: MatchingRules{
				MatchingRule{
					Path: "docker-compose.yaml",
				},
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
					Jobs: []string{
						"mongodb",
					},
				},
			},
			filePath:       "docker-compose.yaml",
			job:            "mongodb",
			expectedResult: true,
		},
		{
			name: "Scenario 3 - not matching all rules",
			rules: MatchingRules{
				MatchingRule{
					Path: "docker-compose.2.yaml",
					Jobs: []string{
						"mongodb",
					},
				},
			},
			filePath:       "docker-compose.yaml",
			job:            "mongodb",
			expectedResult: false,
		},
		{
			name: "Scenario 4 - only matching image name",
			rules: MatchingRules{
				MatchingRule{
					Images: []string{
						"mongo",
					},
				},
			},
			filePath:       "docker-compose.yaml",
			job:            "mongodb",
			image:          "mongo:6",
			expectedResult: true,
		},
		{
			name: "Scenario 5 - matching image name and tag",
			rules: MatchingRules{
				MatchingRule{
					Images: []string{
						"mongo:6",
					},
				},
			},
			filePath:       "docker-compose.yaml",
			job:            "mongodb",
			image:          "mongo:6",
			expectedResult: true,
		},
		{
			name: "Scenario 6 - correct image but wrong tag",
			rules: MatchingRules{
				MatchingRule{
					Images: []string{
						"mongo:6",
					},
				},
			},
			filePath:       "docker-compose.yaml",
			job:            "mongodb",
			image:          "mongo:7",
			expectedResult: false,
		},
		{
			name: "Scenario 6 - correct image and arch",
			rules: MatchingRules{
				MatchingRule{
					Images: []string{
						"mongo:6",
					},
				},
			},
			filePath:       "docker-compose.yaml",
			job:            "mongodb",
			image:          "mongo:6",
			expectedResult: true,
		},
	}

	for _, tt := range testdata {

		t.Run(tt.name, func(t *testing.T) {
			gotResult := tt.rules.isMatchingRule(
				tt.rootDir,
				tt.filePath,
				tt.job,
				tt.image)

			assert.Equal(t, tt.expectedResult, gotResult)

		})
	}

}

func TestMatchingRulesValidate(t *testing.T) {
	tests := []struct {
		name        string
		rules       MatchingRules
		expectError bool
		errorMsg    string
	}{
		{
			name:        "empty rules should pass",
			rules:       MatchingRules{},
			expectError: false,
		},
		{
			name: "rule with path should pass",
			rules: MatchingRules{
				{Path: "*.hcl"},
			},
			expectError: false,
		},
		{
			name: "rule with jobs should pass",
			rules: MatchingRules{
				{Jobs: []string{"my-job"}},
			},
			expectError: false,
		},
		{
			name: "rule with images should pass",
			rules: MatchingRules{
				{Images: []string{"nginx"}},
			},
			expectError: false,
		},
		{
			name: "empty rule should fail",
			rules: MatchingRules{
				{},
			},
			expectError: true,
			errorMsg:    "rule 1 has no valid fields",
		},
		{
			name: "second empty rule should fail",
			rules: MatchingRules{
				{Path: "*.hcl"},
				{},
			},
			expectError: true,
			errorMsg:    "rule 2 has no valid fields",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rules.Validate()
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
