package test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/pipeline/source"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/terragrunt"
	"github.com/updatecli/updatecli/pkg/plugins/resources/hcl"
	terraformRegistry "github.com/updatecli/updatecli/pkg/plugins/resources/terraform/registry"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git"
	"github.com/updatecli/updatecli/pkg/plugins/utils/test"
)

func TestDiscoverManifests(t *testing.T) {
	expectedInlined := config.Spec{
		Name: "Bump Terraform module terraform-aws-modules/rdss/aws version",
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
					},
				},
			},
		},
		Targets: map[string]target.Config{
			"terragruntModuleFile": {
				SourceID: "latestVersion",
				ResourceConfig: resource.ResourceConfig{
					Name: `Bump terraform-aws-modules/rdss/aws to {{ source "latestVersion" }}`,
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
		Name: "Bump Terraform module terraform-aws-modules/aurora/aws version",
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
					},
				},
			},
		},
		Targets: map[string]target.Config{
			"terragruntModuleFile": {
				SourceID: "latestVersion",
				ResourceConfig: resource.ResourceConfig{
					Name: `Bump terraform-aws-modules/aurora/aws to {{ source "latestVersion" }}`,
					Kind: "hcl",
					Spec: hcl.Spec{
						File: "simple_localized.hcl",
						Path: "local.base_source_url",
					},
				},
			},
		},
	}
	expectedComplexLocalized := config.Spec{
		Name: "Bump Terraform module terraform-aws-modules/vpc/aws version",
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
					},
				},
			},
		},
		Targets: map[string]target.Config{
			"terragruntModuleFile": {
				SourceID: "latestVersion",
				ResourceConfig: resource.ResourceConfig{
					Name: `Bump terraform-aws-modules/vpc/aws to {{ source "latestVersion" }}`,
					Kind: "hcl",
					Spec: hcl.Spec{
						File: "complex_localized.hcl",
						Path: "local.module_version",
					},
				},
			},
		},
	}
	expectedNonTfr := config.Spec{
		Name: "Bump Terraform module github.com:hashicorp/exampleLongNameForSorting.git version",
		SCMs: map[string]scm.Config{
			"module": {
				Kind: "git",
				Spec: git.Spec{
					URL: "git::ssh://github.com:hashicorp/exampleLongNameForSorting.git",
				},
			},
		},
		Sources: map[string]source.Config{
			"latestVersion": {
				ResourceConfig: resource.ResourceConfig{
					Name:  "Get latest version of the github.com:hashicorp/exampleLongNameForSorting.git module",
					Kind:  "gitag",
					SCMID: "module",
				},
			},
		},
		Targets: map[string]target.Config{
			"terragruntModuleFile": {
				SourceID: "latestVersion",
				ResourceConfig: resource.ResourceConfig{
					Name: `Bump github.com:hashicorp/exampleLongNameForSorting.git to {{ source "latestVersion" }}`,
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
				expectedNonTfr,
			},
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			resource, err := terragrunt.New(
				terragrunt.Spec{
					RootDir: tt.rootDir,
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
