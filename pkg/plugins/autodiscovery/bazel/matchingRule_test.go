package bazel

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
				{Path: "testdata/project1/MODULE.bazel"},
			},
			expectError: false,
		},
		{
			name: "rule with modules should pass",
			rules: MatchingRules{
				{Modules: map[string]string{"rules_go": ""}},
			},
			expectError: false,
		},
		{
			name: "rule with multiple fields should pass",
			rules: MatchingRules{
				{
					Path:    "testdata/*/MODULE.bazel",
					Modules: map[string]string{"rules_go": ""},
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
				{Path: "testdata/*/MODULE.bazel"},
				{},
			},
			expectError: true,
			errorMsg:    "rule 2 has no valid fields",
		},
		{
			name: "rule with empty path and empty modules map should fail",
			rules: MatchingRules{
				{
					Path:    "",
					Modules: map[string]string{},
				},
			},
			expectError: true,
			errorMsg:    "rule 1 has no valid fields",
		},
		{
			name: "rule with empty module name in modules map should fail",
			rules: MatchingRules{
				{
					Modules: map[string]string{"": ""},
				},
			},
			expectError: true,
			errorMsg:    "rule 1 contains empty module name",
		},
		{
			name: "rule with empty module name and valid module name should fail",
			rules: MatchingRules{
				{
					Modules: map[string]string{
						"rules_go": "",
						"":         "",
					},
				},
			},
			expectError: true,
			errorMsg:    "rule 1 contains empty module name",
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
