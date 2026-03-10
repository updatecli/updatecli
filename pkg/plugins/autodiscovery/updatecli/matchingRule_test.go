package updatecli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsMatchingRule(t *testing.T) {

	dataset := []struct {
		rules          MatchingRules
		name           string
		filePath       string
		policyName     string
		policyVersion  string
		rootDir        string
		expectedResult bool
	}{
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "updatecli-compose.yaml",
				},
			},
			filePath:       "updatecli-compose.yaml",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "updatecli-compose.yaml",
				},
			},
			filePath:       "./website/updatecli-compose.yaml",
			expectedResult: false,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "updatecli-compose.yaml",
					Policies: map[string]string{
						"ghcr.io/updatecli/policies/policies/nodejs/githubaction": "0.1.0",
					},
				},
			},
			filePath:       "updatecli-compose.yaml",
			policyName:     "ghcr.io/updatecli/policies/policies/nodejs/githubaction",
			policyVersion:  "0.1.0",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "updatecli-compose.yaml",
					Policies: map[string]string{
						"ghcr.io/updatecli/policies/policies/nodejs/githubaction": "0.1.0",
					},
				},
			},
			filePath:       "updatecli-compose.yaml",
			policyName:     "ghcr.io/updatecli/policies/policies/nodejs/githubaction",
			expectedResult: false,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "updatecli-compose.yaml",
					Policies: map[string]string{
						"ghcr.io/updatecli/policies/policies/nodejs/githubaction": "",
					},
				},
			},
			filePath:       "updatecli-compose.yaml",
			policyName:     "ghcr.io/updatecli/policies/policies/nodejs/githubaction",
			policyVersion:  "0.1.0",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "updatecli-compose.yaml",
					Policies: map[string]string{
						"ghcr.io/updatecli/policies/policies/nodejs/githubaction": "",
					},
				},
			},
			filePath:       "updatecli-compose.yaml",
			policyName:     "ghcr.io/updatecli/policies/policies/nodejs/githubaction",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "updatecli-compose.yaml",
					Policies: map[string]string{
						"ghcr.io/updatecli/policies/policies/nodejs/githubaction": "0.1.0",
					},
				},
			},
			filePath:       "updatecli-compose.yaml",
			policyName:     "ghcr.io/updatecli/policies/policies/nodejs/netlify",
			expectedResult: false,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "website/updatecli-compose.yaml",
					Policies: map[string]string{
						"ghcr.io/updatecli/policies/policies/nodejs/githubaction": "",
					},
				},
			},
			filePath:       "updatecli-compose.yaml",
			policyName:     "ghcr.io/updatecli/policies/policies/nodejs/githubaction",
			expectedResult: false,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "updatecli-compose.yaml",
				},
				MatchingRule{
					Policies: map[string]string{
						"ghcr.io/updatecli/policies/policies/nodejs/githubaction": "0.1.0",
					},
				},
			},
			filePath:       "updatecli-compose.yaml",
			policyName:     "ghcr.io/updatecli/policies/policies/nodejs/githubaction",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "updatecli-compose.yaml",
				},
				MatchingRule{
					Policies: map[string]string{
						"ghcr.io/updatecli/policies/policies/nodejs/githubaction": ">=0.1.0",
					},
				},
			},
			filePath:       "updatecli-compose.yaml",
			policyName:     "ghcr.io/updatecli/policies/policies/nodejs/githubaction",
			policyVersion:  "0.1.0",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "updatecli-compose.yaml",
					Policies: map[string]string{
						"ghcr.io/updatecli/policies/policies/nodejs/githubaction": ">=1.0.0",
					},
				},
			},
			filePath:       "updatecli-compose.yaml",
			policyName:     "ghcr.io/updatecli/policies/policies/nodejs/githubaction",
			policyVersion:  "0.1.0",
			expectedResult: false,
		},
	}

	for _, d := range dataset {
		t.Run(d.name, func(t *testing.T) {
			gotResult := d.rules.isMatchingRules(
				d.rootDir,
				d.filePath,
				d.policyName,
				d.policyVersion)

			assert.Equal(t, d.expectedResult, gotResult)
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
				{Path: "updatecli-compose.yaml"},
			},
			expectError: false,
		},
		{
			name: "rule with policies should pass",
			rules: MatchingRules{
				{Policies: map[string]string{"ghcr.io/updatecli/policies/policy": ""}},
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
				{Path: "updatecli-compose.yaml"},
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
