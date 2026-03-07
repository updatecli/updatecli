package helm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsMatchingRule(t *testing.T) {

	dataset := []struct {
		rules             MatchingRules
		name              string
		filePath          string
		containerVersion  string
		containerName     string
		dependencyName    string
		dependencyVersion string
		rootDir           string
		expectedResult    bool
	}{
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "test/testdata-1",
				},
			},
			filePath:       "test/testdata-1",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "testdata-1",
				},
			},
			filePath:       "test/testdata-1",
			expectedResult: false,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "test/testdata-1",
					Dependencies: map[string]string{
						"jenkins": "2.234",
					},
				},
			},
			filePath:          "test/testdata-1",
			dependencyName:    "jenkins",
			dependencyVersion: "2.234",
			expectedResult:    true,
		},
	}

	for _, d := range dataset {
		t.Run(d.name, func(t *testing.T) {
			gotResult := d.rules.isMatchingRules(
				d.rootDir,
				d.filePath,
				d.dependencyName,
				d.dependencyVersion,
				d.containerName,
				d.containerVersion,
			)

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
				{Path: "Chart.yaml"},
			},
			expectError: false,
		},
		{
			name: "rule with dependencies should pass",
			rules: MatchingRules{
				{Dependencies: map[string]string{"nginx": ""}},
			},
			expectError: false,
		},
		{
			name: "rule with containers should pass",
			rules: MatchingRules{
				{Containers: map[string]string{"nginx": ""}},
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
				{Path: "Chart.yaml"},
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
