package woodpecker

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsMatchingRule(t *testing.T) {
	testdata := []struct {
		name     string
		rules    MatchingRules
		rootDir  string
		filePath string
		image    string
		expected bool
	}{
		{
			name: "Match by path",
			rules: MatchingRules{
				{Path: ".woodpecker.yml"},
			},
			rootDir:  "/project",
			filePath: ".woodpecker.yml",
			image:    "golang:1.21",
			expected: true,
		},
		{
			name: "No match by path",
			rules: MatchingRules{
				{Path: ".woodpecker.yaml"},
			},
			rootDir:  "/project",
			filePath: ".woodpecker.yml",
			image:    "golang:1.21",
			expected: false,
		},
		{
			name: "Match by image prefix",
			rules: MatchingRules{
				{Images: []string{"golang"}},
			},
			rootDir:  "/project",
			filePath: ".woodpecker.yml",
			image:    "golang:1.21",
			expected: true,
		},
		{
			name: "No match by image",
			rules: MatchingRules{
				{Images: []string{"node"}},
			},
			rootDir:  "/project",
			filePath: ".woodpecker.yml",
			image:    "golang:1.21",
			expected: false,
		},
		{
			name: "Match by path and image",
			rules: MatchingRules{
				{
					Path:   ".woodpecker.yml",
					Images: []string{"golang"},
				},
			},
			rootDir:  "/project",
			filePath: ".woodpecker.yml",
			image:    "golang:1.21",
			expected: true,
		},
		{
			name: "Path matches but image does not",
			rules: MatchingRules{
				{
					Path:   ".woodpecker.yml",
					Images: []string{"node"},
				},
			},
			rootDir:  "/project",
			filePath: ".woodpecker.yml",
			image:    "golang:1.21",
			expected: false,
		},
		{
			name: "Multiple rules - one matches",
			rules: MatchingRules{
				{Images: []string{"node"}},
				{Images: []string{"golang"}},
			},
			rootDir:  "/project",
			filePath: ".woodpecker.yml",
			image:    "golang:1.21",
			expected: true,
		},
		{
			name:     "Empty rules",
			rules:    MatchingRules{},
			rootDir:  "/project",
			filePath: ".woodpecker.yml",
			image:    "golang:1.21",
			expected: false,
		},
		{
			name: "Match with wildcard path pattern",
			rules: MatchingRules{
				{Path: ".woodpecker/*.yml"},
			},
			rootDir:  "/project",
			filePath: ".woodpecker/build.yml",
			image:    "golang:1.21",
			expected: true,
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.rules.isMatchingRule(tt.rootDir, tt.filePath, tt.image)
			assert.Equal(t, tt.expected, result)
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
				{Path: ".woodpecker/*.yaml"},
			},
			expectError: false,
		},
		{
			name: "rule with images should pass",
			rules: MatchingRules{
				{Images: []string{"nginx"}},
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
				{Path: ".woodpecker/*.yaml"},
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
