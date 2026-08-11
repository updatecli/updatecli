package dockerfile

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
		service        string
		image          string
		arch           string
		expectedResult bool
	}{
		{
			name: "Scenario 1 - matching 1 rule",
			rules: MatchingRules{
				MatchingRule{
					Path: "Dockerfile",
				},
			},
			filePath:       "Dockerfile",
			expectedResult: true,
		},
		{
			name: "Scenario 1 - matching 1 rule",
			rules: MatchingRules{
				MatchingRule{
					Path: "Dockerfile",
				},
				MatchingRule{
					Path: "alpine/Dockerfile",
				},
			},
			filePath:       "Dockerfile",
			expectedResult: true,
		},
		{
			name: "Scenario 2 - matching all rule",
			rules: MatchingRules{
				MatchingRule{
					Path: "Dockerfile.alpine",
				},
			},
			filePath:       "Dockerfile",
			expectedResult: false,
		},
		{
			name: "Scenario 4 - only matching image name",
			rules: MatchingRules{
				MatchingRule{
					Images: []string{
						"updatecli/updatecli:latest",
					},
				},
			},
			filePath:       "Dockerfile",
			expectedResult: false,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "Dockerfile",
				},
				MatchingRule{
					Images: []string{
						"updatecli/updatecli:latest",
					},
				},
			},
			filePath:       "alpine/Dockerfile",
			image:          "updatecli/updatecli:latest",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "Dockerfile",
					Images: []string{
						"updatecli/updatecli:latest",
					},
				},
			},
			filePath:       "alpine/Dockerfile",
			image:          "updatecli/updatecli:latest",
			expectedResult: false,
		},
	}

	for _, tt := range testdata {

		t.Run(tt.name, func(t *testing.T) {
			gotResult := tt.rules.isMatchingRule(
				tt.rootDir,
				tt.filePath,
				tt.image,
				tt.arch)

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
				{Path: "Dockerfile"},
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
			name: "rule with archs should pass",
			rules: MatchingRules{
				{Archs: []string{"amd64"}},
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
				{Path: "Dockerfile"},
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
