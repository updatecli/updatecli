package npm

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
					VersionConstraint: true,
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
					VersionConstraint: true,
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
