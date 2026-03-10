package githubaction

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
				{Path: ".github/workflows/*.yaml"},
			},
			expectError: false,
		},
		{
			name: "rule with actions should pass",
			rules: MatchingRules{
				{Actions: map[string]string{"actions/checkout": ""}},
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
				{Path: ".github/workflows/*.yaml"},
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
