package cargo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsMatchingRule(t *testing.T) {

	dataset := []struct {
		name           string
		rules          MatchingRules
		filePath       string
		repository     string
		crateName      string
		crateVersion   string
		rootDir        string
		expectedResult bool
	}{
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "testdata/cargo.toml",
				},
			},
			filePath:       "testdata/cargo.toml",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "testdata/cargo.toml",
				},
			},
			filePath:       "./website/testdata/cargo.toml",
			expectedResult: false,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "testdata/cargo.toml",
					Crates: map[string]string{
						"clap": "",
					},
				},
			},
			filePath:       "testdata/cargo.toml",
			crateName:      "clap",
			expectedResult: true,
		},
	}

	for _, d := range dataset {
		t.Run(d.name, func(t *testing.T) {
			gotResult := d.rules.isMatchingRules(
				d.rootDir,
				d.filePath,
				d.repository,
				d.crateName,
				d.crateVersion)

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
				{Path: "Cargo.toml"},
			},
			expectError: false,
		},
		{
			name: "rule with crates should pass",
			rules: MatchingRules{
				{Crates: map[string]string{"serde": ""}},
			},
			expectError: false,
		},
		{
			name: "rule with registries should pass",
			rules: MatchingRules{
				{Registries: []string{"https://crates.io"}},
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
				{Path: "Cargo.toml"},
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
