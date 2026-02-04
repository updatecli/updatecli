package woodpecker

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
