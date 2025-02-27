package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func TestChangelog(t *testing.T) {
	tests := []struct {
		name           string
		from           string
		to             string
		version        TerraformRegistry
		expectedResult *result.Changelogs
	}{
		{
			name: "Test getting changelog from github",
			from: "v5.1.2",
			to:   "v5.1.2",
			version: TerraformRegistry{
				scm: "https://github.com/terraform-aws-modules/terraform-aws-vpc",
			},
			expectedResult: &result.Changelogs{
				{
					Title:       "v5.1.2",
					Body:        "### [5.1.2](https://github.com/terraform-aws-modules/terraform-aws-vpc/compare/v5.1.1...v5.1.2) (2023-09-07)\n\n\n### Bug Fixes\n\n* The number of intra subnets should not influence the number of NAT gateways provisioned ([#968](https://github.com/terraform-aws-modules/terraform-aws-vpc/issues/968)) ([1e36f9f](https://github.com/terraform-aws-modules/terraform-aws-vpc/commit/1e36f9f8a01eb26be83d8e1ce2227a6890390b0e))\n\n",
					PublishedAt: "2023-09-07 20:04:01 +0000 UTC",
					URL:         "https://github.com/terraform-aws-modules/terraform-aws-vpc/releases/tag/v5.1.2",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedResult, tt.version.Changelog(tt.from, tt.to))
		})
	}
}
