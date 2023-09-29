package terraform

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsMatchingRule(t *testing.T) {
	dataset := []struct {
		rules           MatchingRules
		name            string
		filePath        string
		providerName    string
		providerVersion string
		rootDir         string
		expectedResult  bool
	}{
		{
			rules: MatchingRules{
				MatchingRule{
					Path: ".terraform.lock.hcl",
				},
			},
			filePath:       ".terraform.lock.hcl",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: ".terraform.lock.hcl",
				},
			},
			filePath:       ".terraform.lock.hcl",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: ".terraform.lock.hcl",
				},
			},
			filePath:       "./module/.terraform.lock.hcl",
			expectedResult: false,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: ".terraform.lock.hcl",
				},
			},
			filePath:       ".terraform.lock.hcl",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: ".terraform.lock.hcl",
				},
				MatchingRule{
					Providers: map[string]string{
						"registry.terraform.io/hashicorp/aws": "",
					},
				},
			},
			filePath:       ".terraform.lock.hcl",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: ".terraform.lock.hcl",
					Providers: map[string]string{
						"registry.terraform.io/hashicorp/aws": "",
					},
				},
			},
			filePath:       ".terraform.lock.hcl",
			providerName:   "registry.terraform.io/hashicorp/aws",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Providers: map[string]string{
						"registry.terraform.io/hashicorp/aws": "",
					},
				},
			},
			providerName:   "registry.terraform.io/hashicorp/aws",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Providers: map[string]string{
						"registry.terraform.io/hashicorp/aws": ">=5",
					},
				},
			},
			providerName:    "registry.terraform.io/hashicorp/aws",
			providerVersion: "5.9.0",
			expectedResult:  true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Providers: map[string]string{
						"registry.terraform.io/hashicorp/aws": ">6",
					},
				},
			},
			providerName:    "registry.terraform.io/hashicorp/aws",
			providerVersion: "5.9.0",
			expectedResult:  false,
		},
	}

	for _, d := range dataset {
		t.Run(d.name, func(t *testing.T) {
			gotResult := d.rules.isMatchingRules(
				d.rootDir,
				d.filePath,
				d.providerName,
				d.providerVersion)

			assert.Equal(t, d.expectedResult, gotResult)
		})
	}
}
