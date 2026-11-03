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
			name: "rule with dependencygroups should pass",
			rules: MatchingRules{
				{DependencyGroups: []string{"dev"}},
			},
			expectError: false,
		},
		{
			name: "invalid dependencygroups pattern should fail",
			rules: MatchingRules{
				{DependencyGroups: []string{"[dev"}},
			},
			expectError: true,
			errorMsg:    "invalid dependencygroups pattern",
		},
		{
			// "[a_]" is valid but normalizes to "[a-]", an unterminated range.
			name: "dependencygroups pattern invalid once normalized should fail",
			rules: MatchingRules{
				{DependencyGroups: []string{"[a_]"}},
			},
			expectError: true,
			errorMsg:    "invalid dependencygroups pattern",
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
		name            string
		rules           MatchingRules
		rootDir         string
		filePath        string
		packageName     string
		packageVersion  string
		dependencyGroup string
		expectedResult  bool
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
		{
			name: "dependency group matches",
			rules: MatchingRules{
				{DependencyGroups: []string{"dev"}},
			},
			filePath:        "pyproject.toml",
			packageName:     "pytest",
			dependencyGroup: "dev",
			expectedResult:  true,
		},
		{
			name: "dependency group names are compared normalized",
			rules: MatchingRules{
				{DependencyGroups: []string{"dev-tools"}},
			},
			filePath:        "pyproject.toml",
			packageName:     "pytest",
			dependencyGroup: "Dev_Tools",
			expectedResult:  true,
		},
		{
			name: "dependency group patterns are normalized too",
			rules: MatchingRules{
				{DependencyGroups: []string{"Dev.*"}},
			},
			filePath:        "pyproject.toml",
			packageName:     "pytest",
			dependencyGroup: "dev-tools",
			expectedResult:  true,
		},
		{
			name: "dependency group wildcard matches any group",
			rules: MatchingRules{
				{DependencyGroups: []string{"*"}},
			},
			filePath:        "pyproject.toml",
			packageName:     "pytest",
			dependencyGroup: "test",
			expectedResult:  true,
		},
		{
			name: "dependency group does not match another group",
			rules: MatchingRules{
				{DependencyGroups: []string{"dev"}},
			},
			filePath:        "pyproject.toml",
			packageName:     "pytest",
			dependencyGroup: "test",
			expectedResult:  false,
		},
		{
			name: "dependency group never matches a main dependency",
			rules: MatchingRules{
				{DependencyGroups: []string{"*"}},
			},
			filePath:       "pyproject.toml",
			packageName:    "requests",
			expectedResult: false,
		},
		{
			name: "dependency group and package are AND-ed",
			rules: MatchingRules{
				{DependencyGroups: []string{"dev"}, Packages: map[string]string{"ruff": ""}},
			},
			filePath:        "pyproject.toml",
			packageName:     "pytest",
			dependencyGroup: "dev",
			expectedResult:  false,
		},
	}

	for _, d := range dataset {
		t.Run(d.name, func(t *testing.T) {
			var groups []string
			if d.dependencyGroup != "" {
				groups = []string{d.dependencyGroup}
			}
			got := d.rules.isMatchingRules(d.rootDir, d.filePath, d.packageName, d.packageVersion, groups)
			assert.Equal(t, d.expectedResult, got)
		})
	}
}
