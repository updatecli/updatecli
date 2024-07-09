package updatecli

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
