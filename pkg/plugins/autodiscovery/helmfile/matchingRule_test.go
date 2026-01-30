package helmfile

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
					Path: "testdata/helmfile.d/cik8s.yaml",
				},
			},
			filePath:       "testdata/helmfile.d/cik8s.yaml",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "testdata/helmfile.d/cik8s.yaml",
				},
			},
			filePath:       "./website/testdata/helmfile.d/cik8s.yaml",
			expectedResult: false,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					//Path: "testdata/helmfile.d/cik8s.yaml",
					Charts: map[string]string{
						"datadog": "",
					},
				},
			},
			filePath:       "testdata/helmfile.d/cik8s.yaml",
			chartName:      "datadog",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "testdata/helmfile.d/cik8s.yaml",
					Repositories: []string{
						"https://helm.example.com",
					},
				},
			},
			filePath:       "testdata/helmfile.d/cik8s.yaml",
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
				{Path: "testdata/helmfile.d/*.yaml"},
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
			name: "rule with charts should pass",
			rules: MatchingRules{
				{Charts: map[string]string{"datadog": ""}},
			},
			expectError: false,
		},
		{
			name: "rule with multiple fields should pass",
			rules: MatchingRules{
				{
					Path:         "testdata/*.yaml",
					Repositories: []string{"https://helm.example.com"},
				},
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
				{Path: "testdata/*.yaml"},
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
