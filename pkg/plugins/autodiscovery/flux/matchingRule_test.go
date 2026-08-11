package flux

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
					Artifacts: map[string]string{
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
				{Path: "flux/*.yaml"},
			},
			expectError: false,
		},
		{
			name: "rule with repositories should pass",
			rules: MatchingRules{
				{Repositories: []string{"https://helm.example.com"}},
			},
			expectError: false,
		},
		{
			name: "rule with artifacts should pass",
			rules: MatchingRules{
				{Artifacts: map[string]string{"nginx": ""}},
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
				{Path: "flux/*.yaml"},
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
