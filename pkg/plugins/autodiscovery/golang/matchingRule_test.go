package golang

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
		goVersion      string
		moduleName     string
		moduleVersion  string
		rootDir        string
		expectedResult bool
	}{
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "go.mod",
				},
			},
			filePath:       "go.mod",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "go.mod",
				},
			},
			filePath:       "go.mod",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "go.mod",
				},
			},
			filePath:       "./pkg/go.mod",
			expectedResult: false,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path:      "go.mod",
					GoVersion: ">=1",
				},
			},
			filePath:       "go.mod",
			goVersion:      "1.20",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "go.mod",
				},
				MatchingRule{
					Modules: map[string]string{
						"github.com/updatecli/updatecli": "",
					},
				},
			},
			filePath:       "go.mod",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "go.mod",
					Modules: map[string]string{
						"github.com/updatecli/updatecli": "",
					},
				},
			},
			filePath:       "go.mod",
			moduleName:     "github.com/updatecli/updatecli",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Modules: map[string]string{
						"github.com/updatecli/updatecli": "",
					},
				},
			},
			moduleName:     "github.com/updatecli/updatecli",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Modules: map[string]string{
						"github.com/updatecli/*": "",
					},
				},
			},
			moduleName:     "github.com/updatecli/updatecli",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Modules: map[string]string{
						"github.com/*": "",
					},
				},
			},
			moduleName:     "github.com/updatecli/updatecli",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Modules: map[string]string{
						"github.com/updatecli/updatecli": ">=0",
					},
				},
			},
			moduleName:     "github.com/updatecli/updatecli",
			moduleVersion:  "0.42.0",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Modules: map[string]string{
						"github.com/updatecli/updatecli": ">0",
					},
				},
			},
			moduleName:     "github.com/updatecli/updatecli",
			moduleVersion:  "0.42.0",
			expectedResult: false,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Modules: map[string]string{
						"github.com/updatecli/updatecli": "updatecli/1.0",
					},
				},
			},
			moduleName:     "github.com/updatecli/updatecli",
			moduleVersion:  "updatecli/1.0",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Modules: map[string]string{
						"github.com/updatecli/updatecli": "updatecli/1.0",
					},
				},
			},
			moduleName:     "github.com/updatecli/updatecli",
			moduleVersion:  "updatecli/2.0",
			expectedResult: false,
		},
	}

	for _, d := range dataset {
		t.Run(d.name, func(t *testing.T) {
			gotResult := d.rules.isMatchingRules(
				d.rootDir,
				d.filePath,
				d.goVersion,
				d.moduleName,
				d.moduleVersion,
				false,
			)

			assert.Equal(t, d.expectedResult, gotResult)
		})
	}
}

func TestIsGoOnly(t *testing.T) {

	dataset := []struct {
		name           string
		rules          MatchingRules
		expectedResult bool
	}{
		{
			name: "Only path specified",
			rules: MatchingRules{
				MatchingRule{
					Path: "go.mod",
				},
			},
			expectedResult: false,
		},
		{
			name: "Only go version specified",
			rules: MatchingRules{
				MatchingRule{
					GoVersion: "*",
				},
			},
			expectedResult: true,
		},
		{
			name: "Multiple go version specified",
			rules: MatchingRules{
				MatchingRule{
					GoVersion: "1.19.*",
				},
				MatchingRule{
					GoVersion: ">=1.20.0",
				},
			},
			expectedResult: true,
		},
		{
			name: "Go version specified with second go module rule",
			rules: MatchingRules{
				MatchingRule{
					GoVersion: "1.19.*",
				},
				MatchingRule{
					Modules: map[string]string{
						"github.com/updatecli/updatecli": "1.0.0",
					},
				},
			},
			expectedResult: false,
		},
		{
			name: "Go version specified with go module within the same rule",
			rules: MatchingRules{
				MatchingRule{
					GoVersion: "1.19.*",
					Modules: map[string]string{
						"github.com/updatecli/updatecli": "1.0.0",
					},
				},
			},
			expectedResult: false,
		},
	}

	for _, d := range dataset {
		gotReset := d.rules.isGoVersionOnly()
		assert.Equal(t, d.expectedResult, gotReset)
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
				{Path: "go.mod"},
			},
			expectError: false,
		},
		{
			name: "rule with modules should pass",
			rules: MatchingRules{
				{Modules: map[string]string{"github.com/example/pkg": ""}},
			},
			expectError: false,
		},
		{
			name: "rule with goversion should pass",
			rules: MatchingRules{
				{GoVersion: ">=1.20"},
			},
			expectError: false,
		},
		{
			name: "rule with replace should pass",
			rules: MatchingRules{
				{Replace: &boolTrue},
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
				{Path: "go.mod"},
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
