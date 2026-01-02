package terragrunt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsMatchingRule(t *testing.T) {
	dataset := []struct {
		rules          MatchingRules
		name           string
		filePath       string
		moduleUrl      string
		rootDir        string
		expectedResult bool
	}{
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "terragrunt.hcl",
				},
			},
			filePath:       "terragrunt.hcl",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "terragrunt.hcl",
				},
			},
			filePath:       "terragrunt.hcl",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "terragrunt.hcl",
				},
			},
			filePath:       "./module/terragrunt.hcl",
			expectedResult: false,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "terragrunt.hcl",
				},
			},
			filePath:       "terragrunt.hcl",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "terragrunt.hcl",
				},
				MatchingRule{
					Modules: map[string]string{
						"registry.terraform.io/hashicorp/aws": "",
					},
				},
			},
			filePath:       "terragrunt.hcl",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "terragrunt.hcl",
					Modules: map[string]string{
						"tfr://registry.terraform.io": "",
					},
				},
			},
			filePath:       "terragrunt.hcl",
			moduleUrl:      "tfr:///terraform-aws-modules/rds/aws?version=1.0.0",
			expectedResult: true,
		},

		{
			rules: MatchingRules{
				MatchingRule{
					Path: "terragrunt.hcl",
					Modules: map[string]string{
						"tfr:///": "",
					},
				},
			},
			filePath:       "terragrunt.hcl",
			moduleUrl:      "tfr:///terraform-aws-modules/rds/aws?version=1.0.0",
			expectedResult: true,
		},

		{
			rules: MatchingRules{
				MatchingRule{
					Path: "terragrunt.hcl",
					Modules: map[string]string{
						"tfr:///terraform-aws-modules": "",
					},
				},
			},
			filePath:       "terragrunt.hcl",
			moduleUrl:      "tfr:///terraform-aws-modules/rds/aws?version=1.0.0",
			expectedResult: true,
		},

		{
			rules: MatchingRules{
				MatchingRule{
					Path: "terragrunt.hcl",
					Modules: map[string]string{
						"tfr:///terraform-aws-modules/rds": "",
					},
				},
			},
			filePath:       "terragrunt.hcl",
			moduleUrl:      "tfr:///terraform-aws-modules/rds/aws?version=1.0.0",
			expectedResult: true,
		},

		{
			rules: MatchingRules{
				MatchingRule{
					Path: "terragrunt.hcl",
					Modules: map[string]string{
						"tfr:///terraform-aws-modules/rds/aws": "",
					},
				},
			},
			filePath:       "terragrunt.hcl",
			moduleUrl:      "tfr:///terraform-aws-modules/rds/aws?version=1.0.0",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "terragrunt.hcl",
					Modules: map[string]string{
						"tfr://registry.opentofu.org/terraform-aws-modules/rds/aws": "",
					},
				},
			},
			filePath:       "terragrunt.hcl",
			moduleUrl:      "tfr:///terraform-aws-modules/rds/aws?version=1.0.0",
			expectedResult: false,
		},

		{
			rules: MatchingRules{
				MatchingRule{
					Modules: map[string]string{
						"git@github.com:hashicorp": "",
					},
				},
			},
			moduleUrl:      "git@github.com:hashicorp/exampleLongNameForSorting.git?ref=v2.5.1",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Modules: map[string]string{
						"tfr:///terraform-aws-modules/rds/aws": ">=5",
					},
				},
			},
			moduleUrl:      "tfr:///terraform-aws-modules/rds/aws?version=5.9.0",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Modules: map[string]string{
						"tfr:///terraform-aws-modules/rds/aws": ">=6",
					},
				},
			},
			moduleUrl:      "tfr:///terraform-aws-modules/rds/aws?version=5.9.0",
			expectedResult: false,
		},
	}

	for _, d := range dataset {
		t.Run(d.name, func(t *testing.T) {
			m, _ := getModuleFromUrl(d.moduleUrl, d.moduleUrl, false)
			gotResult, err := d.rules.isMatchingRules(
				d.rootDir,
				d.filePath,
				m)

			assert.NoError(t, err)
			assert.Equal(t, d.expectedResult, gotResult)
		})
	}
}
