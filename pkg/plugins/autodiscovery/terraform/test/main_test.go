package test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/source"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/terraform"
	terraformLock "github.com/updatecli/updatecli/pkg/plugins/resources/terraform/lock"
	terraformRegistry "github.com/updatecli/updatecli/pkg/plugins/resources/terraform/registry"
	"github.com/updatecli/updatecli/pkg/plugins/utils/test"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func TestDiscoverManifests(t *testing.T) {
	expectedAws := config.Spec{
		Name: "Bump Terraform provider hashicorp/aws version",
		Sources: map[string]source.Config{
			"latestVersion": {
				ResourceConfig: resource.ResourceConfig{
					Name: "Get latest version of the hashicorp/aws provider",
					Kind: "terraform/registry",
					Spec: terraformRegistry.Spec{
						Type:      "provider",
						Namespace: "hashicorp",
						Name:      "aws",
						VersionFilter: version.Filter{
							Kind:    "semver",
							Pattern: ">=5.9.0",
						},
					},
				},
			},
		},
		Targets: map[string]target.Config{
			"terraformLockVersion": {
				SourceID: "latestVersion",
				ResourceConfig: resource.ResourceConfig{
					Name: `Bump hashicorp/aws to {{ source "latestVersion" }}`,
					Kind: "terraform/lock",
					Spec: terraformLock.Spec{
						File:     ".terraform.lock.hcl",
						Provider: "hashicorp/aws",
						Platforms: []string{
							"linux_amd64",
							"linux_arm64",
							"darwin_amd64",
							"darwin_arm64",
						},
					},
				},
			},
		},
	}

	expectedCloudInit := config.Spec{
		Name: "Bump Terraform provider hashicorp/cloudinit version",
		Sources: map[string]source.Config{
			"latestVersion": {
				ResourceConfig: resource.ResourceConfig{
					Name: "Get latest version of the hashicorp/cloudinit provider",
					Kind: "terraform/registry",
					Spec: terraformRegistry.Spec{
						Type:      "provider",
						Namespace: "hashicorp",
						Name:      "cloudinit",
						VersionFilter: version.Filter{
							Kind:    "semver",
							Pattern: ">=2.3.2",
						},
					},
				},
			},
		},
		Targets: map[string]target.Config{
			"terraformLockVersion": {
				SourceID: "latestVersion",
				ResourceConfig: resource.ResourceConfig{
					Name: `Bump hashicorp/cloudinit to {{ source "latestVersion" }}`,
					Kind: "terraform/lock",
					Spec: terraformLock.Spec{
						File:     ".terraform.lock.hcl",
						Provider: "hashicorp/cloudinit",
						Platforms: []string{
							"linux_amd64",
							"linux_arm64",
							"darwin_amd64",
							"darwin_arm64",
						},
					},
				},
			},
		},
	}

	testdata := []struct {
		name              string
		rootDir           string
		platforms         []string
		only              terraform.MatchingRules
		ignore            terraform.MatchingRules
		expectedPipelines []config.Spec
	}{
		{
			name:    "Terraform Version",
			rootDir: "testdata",
			platforms: []string{
				"linux_amd64",
				"linux_arm64",
				"darwin_amd64",
				"darwin_arm64",
			},
			expectedPipelines: []config.Spec{
				expectedAws,
				expectedCloudInit,
			},
		},
		{
			name:    "Terraform Version - Only",
			rootDir: "testdata",
			platforms: []string{
				"linux_amd64",
				"linux_arm64",
				"darwin_amd64",
				"darwin_arm64",
			},
			only: terraform.MatchingRules{
				terraform.MatchingRule{
					Providers: map[string]string{
						"registry.terraform.io/hashicorp/aws": "",
					},
				},
			},
			expectedPipelines: []config.Spec{
				expectedAws,
			},
		},
		{
			name:    "Terraform Version - Only",
			rootDir: "testdata",
			platforms: []string{
				"linux_amd64",
				"linux_arm64",
				"darwin_amd64",
				"darwin_arm64",
			},
			ignore: terraform.MatchingRules{
				terraform.MatchingRule{
					Providers: map[string]string{
						"registry.terraform.io/hashicorp/aws": "",
					},
				},
			},
			expectedPipelines: []config.Spec{
				expectedCloudInit,
			},
		},
		{
			name:    "Terraform Version - Only",
			rootDir: "testdata",
			platforms: []string{
				"linux_amd64",
				"linux_arm64",
				"darwin_amd64",
				"darwin_arm64",
			},
			only: terraform.MatchingRules{
				terraform.MatchingRule{
					Providers: map[string]string{
						"registry.terraform.io/hashicorp/aws":       ">6",
						"registry.terraform.io/hashicorp/cloudinit": ">=2",
					},
				},
			},
			expectedPipelines: []config.Spec{
				expectedCloudInit,
			},
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			resource, err := terraform.New(
				terraform.Spec{
					RootDir:   tt.rootDir,
					Platforms: tt.platforms,
					Only:      tt.only,
					Ignore:    tt.ignore,
				}, "", "")
			require.NoError(t, err)

			pipelines, err := resource.DiscoverManifests()
			require.NoError(t, err)

			if len(pipelines) != len(tt.expectedPipelines) {
				t.Logf("%v pipeline detected but expecting %v", len(pipelines), len(tt.expectedPipelines))
				t.Fail()
				return
			}

			// We sort both the pipelines and the expectedPipelines using the same algorithm
			// to ensure the order is the same as map in Golang are unordered
			test.SortConfigSpecArray(t, tt.expectedPipelines, pipelines)

			for i := range pipelines {
				test.AssertConfigSpecEqualByteArray(t, &tt.expectedPipelines[i], string(pipelines[i]))
			}
		})
	}
}
