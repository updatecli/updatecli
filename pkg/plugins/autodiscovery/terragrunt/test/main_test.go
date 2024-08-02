package test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/pipeline/source"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
	"github.com/updatecli/updatecli/pkg/core/transformer"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/terragrunt"
	"github.com/updatecli/updatecli/pkg/plugins/resources/gittag"
	"github.com/updatecli/updatecli/pkg/plugins/resources/hcl"
	terraformRegistry "github.com/updatecli/updatecli/pkg/plugins/resources/terraform/registry"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git"
	"github.com/updatecli/updatecli/pkg/plugins/utils/test"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func TestDiscoverManifests(t *testing.T) {
	expectedInlined := config.Spec{
		Name: "Bump Terragrunt module terraform-aws-modules/rdss/aws version",
		Sources: map[string]source.Config{
			"latestVersion": {
				ResourceConfig: resource.ResourceConfig{
					Name: "Get latest version of the terraform-aws-modules/rdss/aws module",
					Kind: "terraform/registry",
					Spec: terraformRegistry.Spec{
						Hostname:     "registry.terraform.io",
						Type:         "module",
						Namespace:    "terraform-aws-modules",
						Name:         "rdss",
						TargetSystem: "aws",
						VersionFilter: version.Filter{
							Kind:    "semver",
							Pattern: ">=5.8.1",
						},
					},
					Transformers: []transformer.Transformer{{
						AddPrefix: "tfr://terraform-aws-modules/rdss/aws?version=",
					}},
				},
			},
		},
		Targets: map[string]target.Config{
			"terragruntModuleFile": {
				SourceID: "latestVersion",
				ResourceConfig: resource.ResourceConfig{
					Name: `deps: bump terraform-aws-modules/rdss/aws to {{ source "latestVersion" }}`,
					Kind: "hcl",
					Spec: hcl.Spec{
						File: "inlined.hcl",
						Path: "terraform.source",
					},
				},
			},
		},
	}
	expectedSimpleLocalized := config.Spec{
		Name: "Bump Terragrunt module terraform-aws-modules/aurora/aws version",
		Sources: map[string]source.Config{
			"latestVersion": {
				ResourceConfig: resource.ResourceConfig{
					Name: "Get latest version of the terraform-aws-modules/aurora/aws module",
					Kind: "terraform/registry",
					Spec: terraformRegistry.Spec{
						Hostname:     "registry.terraform.io",
						Type:         "module",
						Namespace:    "terraform-aws-modules",
						Name:         "aurora",
						TargetSystem: "aws",
						VersionFilter: version.Filter{
							Kind:    "semver",
							Pattern: ">=5.8.1",
						},
					},
					Transformers: []transformer.Transformer{{
						AddPrefix: "tfr://terraform-aws-modules/aurora/aws?version=",
					}},
				},
			},
		},
		Targets: map[string]target.Config{
			"terragruntModuleFile": {
				SourceID: "latestVersion",
				ResourceConfig: resource.ResourceConfig{
					Name: `deps: bump terraform-aws-modules/aurora/aws to {{ source "latestVersion" }}`,
					Kind: "hcl",
					Spec: hcl.Spec{
						File: "simple_localized.hcl",
						Path: "locals.base_source_url",
					},
				},
			},
		},
	}
	expectedComplexLocalized := config.Spec{
		Name: "Bump Terragrunt module terraform-aws-modules/vpc/aws version",
		Sources: map[string]source.Config{
			"latestVersion": {
				ResourceConfig: resource.ResourceConfig{
					Name: "Get latest version of the terraform-aws-modules/vpc/aws module",
					Kind: "terraform/registry",
					Spec: terraformRegistry.Spec{
						Hostname:     "registry.terraform.io",
						Type:         "module",
						Namespace:    "terraform-aws-modules",
						Name:         "vpc",
						TargetSystem: "aws",
						VersionFilter: version.Filter{
							Kind:    "semver",
							Pattern: ">=5.8.1",
						},
					},
				},
			},
		},
		Targets: map[string]target.Config{
			"terragruntModuleFile": {
				SourceID: "latestVersion",
				ResourceConfig: resource.ResourceConfig{
					Name: `deps: bump terraform-aws-modules/vpc/aws to {{ source "latestVersion" }}`,
					Kind: "hcl",
					Spec: hcl.Spec{
						File: "complex_localized.hcl",
						Path: "locals.module_version",
					},
				},
			},
		},
	}
	expectedSuperComplexLocalized := config.Spec{
		Name: "Bump Terragrunt module terraform-aws-modules/auroravpc/aws version",
		Sources: map[string]source.Config{
			"latestVersion": {
				ResourceConfig: resource.ResourceConfig{
					Name: "Get latest version of the terraform-aws-modules/auroravpc/aws module",
					Kind: "terraform/registry",
					Spec: terraformRegistry.Spec{
						Hostname:     "registry.terraform.io",
						Type:         "module",
						Namespace:    "terraform-aws-modules",
						Name:         "auroravpc",
						TargetSystem: "aws",
						VersionFilter: version.Filter{
							Kind:    "semver",
							Pattern: ">=1.2.3",
						},
					},
					Transformers: []transformer.Transformer{{
						AddPrefix: "tfr://${local.module}?version=",
					}},
				},
			},
		},
		Targets: map[string]target.Config{
			"terragruntModuleFile": {
				SourceID: "latestVersion",
				ResourceConfig: resource.ResourceConfig{
					Name: `deps: bump terraform-aws-modules/auroravpc/aws to {{ source "latestVersion" }}`,
					Kind: "hcl",
					Spec: hcl.Spec{
						File: "more_complex_localized.hcl",
						Path: "terraform.source",
					},
				},
			},
		},
	}
	expectedNonTfr := config.Spec{
		Name: "Bump Terragrunt module github.com/Azure/terraform-azurerm-avm-res-network-virtualnetwork.git version",
		SCMs: map[string]scm.Config{
			"module": {
				Kind: "git",
				Spec: git.Spec{
					URL: "https://github.com/Azure/terraform-azurerm-avm-res-network-virtualnetwork.git",
				},
			},
		},
		Sources: map[string]source.Config{
			"latestVersion": {
				ResourceConfig: resource.ResourceConfig{
					Name:  "Get latest version of the github.com/Azure/terraform-azurerm-avm-res-network-virtualnetwork.git module",
					Kind:  "gittag",
					SCMID: "module",
					Spec: gittag.Spec{

						VersionFilter: version.Filter{
							Kind:    "semver",
							Pattern: ">=0.3.0",
						},
					},
					Transformers: []transformer.Transformer{{
						AddPrefix: "git::https://github.com/Azure/terraform-azurerm-avm-res-network-virtualnetwork.git?ref=",
					}},
				},
			},
		},
		Targets: map[string]target.Config{
			"terragruntModuleFile": {
				SourceID: "latestVersion",
				ResourceConfig: resource.ResourceConfig{
					Name: `deps: bump github.com/Azure/terraform-azurerm-avm-res-network-virtualnetwork.git to {{ source "latestVersion" }}`,
					Kind: "hcl",
					Spec: hcl.Spec{
						File: "non_tfr.hcl",
						Path: "terraform.source",
					},
				},
			},
		},
	}
	testdata := []struct {
		name              string
		rootDir           string
		expectedPipelines []config.Spec
	}{
		{
			name:    "Terraform Version",
			rootDir: "testdata",
			expectedPipelines: []config.Spec{
				expectedComplexLocalized,
				expectedInlined,
				expectedSimpleLocalized,
				expectedSuperComplexLocalized,
				expectedNonTfr,
			},
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			resource, err := terragrunt.New(
				terragrunt.Spec{}, tt.rootDir, "")
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
