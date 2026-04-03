package pyproject

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
				{Path: "pyproject.toml"},
			},
			expectError: false,
		},
		{
			name: "rule with packages should pass",
			rules: MatchingRules{
				{Packages: map[string]string{"requests": ""}},
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
				{Path: "pyproject.toml"},
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

func TestIsMatchingRules(t *testing.T) {
	dataset := []struct {
		name           string
		rules          MatchingRules
		rootDir        string
		filePath       string
		packageName    string
		packageVersion string
		expectedResult bool
	}{
		{
			name: "path matches exactly",
			rules: MatchingRules{
				{Path: "pyproject.toml"},
			},
			filePath:       "pyproject.toml",
			expectedResult: true,
		},
		{
			name: "path does not match",
			rules: MatchingRules{
				{Path: "pyproject.toml"},
			},
			filePath:       "subdir/pyproject.toml",
			expectedResult: false,
		},
		{
			name: "package name matches",
			rules: MatchingRules{
				{Packages: map[string]string{"requests": ""}},
			},
			filePath:       "pyproject.toml",
			packageName:    "requests",
			expectedResult: true,
		},
		{
			name: "package name does not match",
			rules: MatchingRules{
				{Packages: map[string]string{"requests": ""}},
			},
			filePath:       "pyproject.toml",
			packageName:    "flask",
			expectedResult: false,
		},
		{
			name: "package version constraint matches",
			rules: MatchingRules{
				{Packages: map[string]string{"requests": ">=2.0"}},
			},
			filePath:       "pyproject.toml",
			packageName:    "requests",
			packageVersion: "2.28",
			expectedResult: true,
		},
		{
			name: "package version constraint does not match",
			rules: MatchingRules{
				{Packages: map[string]string{"requests": ">=3.0"}},
			},
			filePath:       "pyproject.toml",
			packageName:    "requests",
			packageVersion: "2.28",
			expectedResult: false,
		},
		{
			name: "path and package AND logic — both must match",
			rules: MatchingRules{
				{
					Path:     "pyproject.toml",
					Packages: map[string]string{"requests": ""},
				},
			},
			filePath:       "pyproject.toml",
			packageName:    "requests",
			expectedResult: true,
		},
		{
			name: "path and package AND logic — path matches but package does not",
			rules: MatchingRules{
				{
					Path:     "pyproject.toml",
					Packages: map[string]string{"requests": ""},
				},
			},
			filePath:       "pyproject.toml",
			packageName:    "flask",
			expectedResult: false,
		},
		{
			name: "multiple rules use OR logic — second rule matches",
			rules: MatchingRules{
				{Packages: map[string]string{"requests": ""}},
				{Packages: map[string]string{"flask": ""}},
			},
			filePath:       "pyproject.toml",
			packageName:    "flask",
			expectedResult: true,
		},
		{
			name:           "empty rules always return false",
			rules:          MatchingRules{},
			filePath:       "pyproject.toml",
			packageName:    "requests",
			expectedResult: false,
		},
		{
			name: "PEP 440 pre-release version matches specifier",
			rules: MatchingRules{
				{Packages: map[string]string{"uvicorn": ">=0.50"}},
			},
			filePath:       "pyproject.toml",
			packageName:    "uvicorn",
			packageVersion: "0.51b0",
			expectedResult: true,
		},
		{
			name: "PEP 440 pre-release version does not match specifier",
			rules: MatchingRules{
				{Packages: map[string]string{"uvicorn": ">=1.0"}},
			},
			filePath:       "pyproject.toml",
			packageName:    "uvicorn",
			packageVersion: "0.51b0",
			expectedResult: false,
		},
	}

	for _, d := range dataset {
		t.Run(d.name, func(t *testing.T) {
			got := d.rules.isMatchingRules(d.rootDir, d.filePath, d.packageName, d.packageVersion)
			assert.Equal(t, d.expectedResult, got)
		})
	}
}
