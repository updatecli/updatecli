package npm

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
		packageName    string
		packageVersion string
		rootDir        string
		expectedResult bool
	}{
		{
			rules: MatchingRules{
				MatchingRule{
					HasVersionConstraint: boolPtr(false),
				},
			},
			filePath:       "package.json",
			packageName:    "@babel/core",
			packageVersion: "0.1.0",
			expectedResult: false,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					HasVersionConstraint: boolPtr(true),
				},
			},
			filePath:       "package.json",
			packageName:    "@babel/core",
			packageVersion: "0.1.x",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					HasVersionConstraint: boolPtr(true),
					Path:                 "package.json.2",
				},
			},
			filePath:       "package.json",
			packageName:    "@babel/core",
			packageVersion: "0.1.x",
			expectedResult: false,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "package.json",
				},
			},
			filePath:       "package.json",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "package.json",
				},
			},
			filePath:       "./website/package.json",
			expectedResult: false,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "package.json",
					Packages: map[string]string{
						"@babel/core": "0.1.0",
					},
				},
			},
			filePath:       "package.json",
			packageName:    "@babel/core",
			packageVersion: "0.1.0",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "package.json",
					Packages: map[string]string{
						"@babel/core": "0.1.0",
					},
				},
			},
			filePath:       "package.json",
			packageName:    "@babel/core",
			expectedResult: false,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "package.json",
					Packages: map[string]string{
						"@babel/core": "",
					},
				},
			},
			filePath:       "package.json",
			packageName:    "@babel/core",
			packageVersion: "0.1.0",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "package.json",
					Packages: map[string]string{
						"@babel/core": "",
					},
				},
			},
			filePath:       "package.json",
			packageName:    "@babel/core",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "package.json",
					Packages: map[string]string{
						"@babel/core": "0.1.0",
					},
				},
			},
			filePath:       "package.json",
			packageName:    "@babel/core",
			expectedResult: false,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "website/package.json",
					Packages: map[string]string{
						"@babel/core": "",
					},
				},
			},
			filePath:       "package.json",
			packageName:    "@babel/core",
			expectedResult: false,
		},
	}

	for _, d := range dataset {
		t.Run(d.name, func(t *testing.T) {
			gotResult := d.rules.isMatchingRules(
				d.rootDir,
				d.filePath,
				d.packageName,
				d.packageVersion)

			assert.Equal(t, d.expectedResult, gotResult)
		})
	}
}

func TestMatchingRulesValidate(t *testing.T) {
	boolTrue := true
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
				{Path: "package.json"},
			},
			expectError: false,
		},
		{
			name: "rule with packages should pass",
			rules: MatchingRules{
				{Packages: map[string]string{"express": ""}},
			},
			expectError: false,
		},
		{
			name: "rule with hasVersionConstraint should pass",
			rules: MatchingRules{
				{HasVersionConstraint: &boolTrue},
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
				{Path: "package.json"},
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
