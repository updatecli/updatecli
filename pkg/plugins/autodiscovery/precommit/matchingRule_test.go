package precommit

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
					Path: ".pre-commit-config.yaml",
				},
			},
			filePath:       ".pre-commit-config.yaml",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: ".pre-commit-config.yaml",
				},
			},
			filePath:       "./website/.pre-commit-config.yaml",
			expectedResult: false,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: ".pre-commit-config.yaml",
					Repos: map[string]string{
						"https://github.com/pre-commit/pre-commit-hooks": "v4.6.0",
					},
				},
			},
			filePath:       ".pre-commit-config.yaml",
			packageName:    "https://github.com/pre-commit/pre-commit-hooks",
			packageVersion: "v4.6.0",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: ".pre-commit-config.yaml",
					Repos: map[string]string{
						"https://github.com/pre-commit/pre-commit-hooks": "0.1.0",
					},
				},
			},
			filePath:       ".pre-commit-config.yaml",
			packageName:    "https://github.com/pre-commit/pre-commit-hooks",
			expectedResult: false,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: ".pre-commit-config.yaml",
					Repos: map[string]string{
						"https://github.com/pre-commit/pre-commit-hooks": "",
					},
				},
			},
			filePath:       ".pre-commit-config.yaml",
			packageName:    "https://github.com/pre-commit/pre-commit-hooks",
			packageVersion: "0.1.0",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: ".pre-commit-config.yaml",
					Repos: map[string]string{
						"https://github.com/pre-commit/pre-commit-hooks": "",
					},
				},
			},
			filePath:       ".pre-commit-config.yaml",
			packageName:    "https://github.com/pre-commit/pre-commit-hooks",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: ".pre-commit-config.yaml",
					Repos: map[string]string{
						"https://github.com/pre-commit/pre-commit-hooks": "0.1.0",
					},
				},
			},
			filePath:       ".pre-commit-config.yaml",
			packageName:    "https://github.com/pre-commit/pre-commit-hooks",
			expectedResult: false,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "website/.pre-commit-config.yaml",
					Repos: map[string]string{
						"https://github.com/pre-commit/pre-commit-hooks": "",
					},
				},
			},
			filePath:       ".pre-commit-config.yaml",
			packageName:    "https://github.com/pre-commit/pre-commit-hooks",
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
