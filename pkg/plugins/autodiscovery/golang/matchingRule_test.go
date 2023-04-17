package golang

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
				d.moduleVersion)

			assert.Equal(t, d.expectedResult, gotResult)
		})
	}
}
