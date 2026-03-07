package maven

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsMatchingRule(t *testing.T) {

	dataset := []struct {
		name            string
		rules           MatchingRules
		filePath        string
		groupid         string
		artifactName    string
		artifactVersion string
		rootDir         string
		expectedResult  bool
	}{
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "testdata/pom.xml",
				},
			},
			filePath:       "testdata/pom.xml",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "testdata/pom.xml",
				},
			},
			filePath:       "./website/testdata/pom.xml",
			expectedResult: false,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "testdata/pom.xml",
					ArtifactIDs: map[string]string{
						"jenkins-war": "",
					},
				},
			},
			filePath:       "testdata/pom.xml",
			artifactName:   "jenkins-war",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "testdata/pom.xml",
					GroupIDs: []string{
						"junit",
					},
				},
			},
			filePath:       "testdata/pom.xml",
			groupid:        "junit",
			expectedResult: true,
		},
	}

	for _, d := range dataset {
		t.Run(d.name, func(t *testing.T) {
			gotResult := d.rules.isMatchingRules(
				d.rootDir,
				d.filePath,
				d.groupid,
				d.artifactName,
				d.artifactVersion)

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
				{Path: "pom.xml"},
			},
			expectError: false,
		},
		{
			name: "rule with groupids should pass",
			rules: MatchingRules{
				{GroupIDs: []string{"org.springframework"}},
			},
			expectError: false,
		},
		{
			name: "rule with artifactids should pass",
			rules: MatchingRules{
				{ArtifactIDs: map[string]string{"spring-core": ""}},
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
				{Path: "pom.xml"},
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
