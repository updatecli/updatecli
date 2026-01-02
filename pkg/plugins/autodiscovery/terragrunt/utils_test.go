package terragrunt

import (
	"os"
	"strings"
	"testing"

	terraformRegistryAddress "github.com/hashicorp/terraform-registry-address"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	terraformUtils "github.com/updatecli/updatecli/pkg/plugins/resources/terraform"
)

func TestSearchTerragruntFiles(t *testing.T) {
	dataset := []struct {
		name               string
		rootDir            string
		expectedFoundFiles []string
	}{
		{
			name:    "Default working scenario",
			rootDir: "testdata",
			expectedFoundFiles: []string{
				"testdata/complex_localized.hcl",
				"testdata/inlined.hcl",
				"testdata/more_complex_localized.hcl",
				"testdata/non_tfr.hcl",
				"testdata/simple_localized.hcl",
				"testdata/terraform_without_source.hcl",
				"testdata/terragrunt.hcl",
			},
		},
	}

	for _, d := range dataset {
		t.Run(d.name, func(t *testing.T) {
			foundFiles, err := searchTerragruntFiles(d.rootDir)
			require.NoError(t, err)

			assert.Equal(t, foundFiles, d.expectedFoundFiles)
		})
	}
}

func TestGetTerragruntModules(t *testing.T) {
	dataset := []struct {
		name           string
		file           string
		expectedModule *terragruntModule
	}{
		{
			name: "Default working scenario",
			file: "testdata/inlined.hcl",
			expectedModule: &terragruntModule{
				registryModule: &terraformRegistryAddress.Module{
					Package: terraformRegistryAddress.ModulePackage{
						Host:         "registry.terraform.io",
						Namespace:    "terraform-aws-modules",
						Name:         "rdss",
						TargetSystem: "aws",
					},
				},
				source: terragruntModuleSource{
					baseUrl:         "terraform-aws-modules/rdss/aws",
					rawSource:       "\"tfr:///terraform-aws-modules/rdss/aws?version=5.8.1\"",
					evaluatedSource: "tfr:///terraform-aws-modules/rdss/aws?version=5.8.1",
					version:         "5.8.1",
					sourceType:      SourceTypeRegistry,
				},
			},
		},
		{
			name: "Simple localized scenario",
			file: "testdata/simple_localized.hcl",
			expectedModule: &terragruntModule{
				registryModule: &terraformRegistryAddress.Module{
					Package: terraformRegistryAddress.ModulePackage{
						Host:         "registry.terraform.io",
						Namespace:    "terraform-aws-modules",
						Name:         "aurora",
						TargetSystem: "aws",
					},
				},
				source: terragruntModuleSource{
					baseUrl:         "terraform-aws-modules/aurora/aws",
					rawSource:       "local.base_source_url",
					evaluatedSource: "tfr:///terraform-aws-modules/aurora/aws?version=5.8.1",
					version:         "5.8.1",
					sourceType:      SourceTypeRegistry,
				},
				hclContext: &map[string]string{
					"base_source_url": "tfr:///terraform-aws-modules/aurora/aws?version=5.8.1",
				},
			},
		},
		{
			name: "Complex localized scenario",
			file: "testdata/complex_localized.hcl",
			expectedModule: &terragruntModule{
				registryModule: &terraformRegistryAddress.Module{
					Package: terraformRegistryAddress.ModulePackage{
						Host:         "registry.terraform.io",
						Namespace:    "terraform-aws-modules",
						Name:         "vpc",
						TargetSystem: "aws",
					},
				},
				source: terragruntModuleSource{
					baseUrl:         "terraform-aws-modules/vpc/aws",
					rawSource:       "\"tfr:///${local.module}?version=${local.module_version}\"",
					evaluatedSource: "tfr:///terraform-aws-modules/vpc/aws?version=5.8.1",
					version:         "5.8.1",
					sourceType:      SourceTypeRegistry,
				},
				hclContext: &map[string]string{
					"module":         "terraform-aws-modules/vpc/aws",
					"module_version": "5.8.1",
				},
			},
		},
		{
			name: "Super Complex localized scenario",
			file: "testdata/more_complex_localized.hcl",
			expectedModule: &terragruntModule{
				registryModule: &terraformRegistryAddress.Module{
					Package: terraformRegistryAddress.ModulePackage{
						Host:         "registry.terraform.io",
						Namespace:    "terraform-aws-modules",
						Name:         "auroravpc",
						TargetSystem: "aws",
					},
				},
				source: terragruntModuleSource{
					baseUrl:         "terraform-aws-modules/auroravpc/aws",
					rawSource:       "\"tfr:///${local.module}?version=1.2.3\"",
					evaluatedSource: "tfr:///terraform-aws-modules/auroravpc/aws?version=1.2.3",
					version:         "1.2.3",
					sourceType:      SourceTypeRegistry,
				},
				hclContext: &map[string]string{
					"module": "terraform-aws-modules/auroravpc/aws",
				},
			},
		},
		{
			name: "Non tfr scenario",
			file: "testdata/non_tfr.hcl",
			expectedModule: &terragruntModule{
				registryModule: nil,
				source: terragruntModuleSource{
					protocol:        "git::https",
					baseUrl:         "github.com/Azure/terraform-azurerm-avm-res-network-virtualnetwork.git",
					rawSource:       "\"git::https://github.com/Azure/terraform-azurerm-avm-res-network-virtualnetwork.git?ref=v0.3.0\"",
					evaluatedSource: "git::https://github.com/Azure/terraform-azurerm-avm-res-network-virtualnetwork.git?ref=v0.3.0",
					version:         "0.3.0",
					sourceType:      SourceTypeGit,
				},
			},
		},
		{
			name:           "Terragrunt terraform block without source",
			file:           "testdata/terraform_without_source.hcl",
			expectedModule: nil,
		},
	}

	for _, d := range dataset {
		t.Run(d.name, func(t *testing.T) {
			module, err := getTerragruntModule(d.file, false)
			require.NoError(t, err)
			assert.Equal(t, module, d.expectedModule)
		})
	}
}

func TestToSourceUrl(t *testing.T) {
	dataset := []struct {
		name           string
		source         string
		expectedModule terragruntModuleSource
	}{
		{
			name:   "standard registry",
			source: "tfr:///terraform-aws-modules/vpc/aws?version=3.3.0",
			expectedModule: terragruntModuleSource{
				baseUrl:         "terraform-aws-modules/vpc/aws",
				rawSource:       "tfr:///terraform-aws-modules/vpc/aws?version=3.3.0",
				evaluatedSource: "tfr:///terraform-aws-modules/vpc/aws?version=3.3.0",
				version:         "3.3.0",
				sourceType:      SourceTypeRegistry,
			}}, {
			name:   "non standard registry",
			source: "tfr://registry.opentofu.org/terraform-aws-modules/vpc/aws?version=3.3.0",
			expectedModule: terragruntModuleSource{
				baseUrl:         "registry.opentofu.org/terraform-aws-modules/vpc/aws",
				rawSource:       "tfr://registry.opentofu.org/terraform-aws-modules/vpc/aws?version=3.3.0",
				evaluatedSource: "tfr://registry.opentofu.org/terraform-aws-modules/vpc/aws?version=3.3.0",
				version:         "3.3.0",
				sourceType:      SourceTypeRegistry,
			},
		},
		{
			name:   "git http repo",
			source: "git::https://github.com/terraform-aws-modules/terraform-aws-lambda.git?ref=v6.0.0",
			expectedModule: terragruntModuleSource{
				protocol:        "git::https",
				baseUrl:         "github.com/terraform-aws-modules/terraform-aws-lambda.git",
				rawSource:       "git::https://github.com/terraform-aws-modules/terraform-aws-lambda.git?ref=v6.0.0",
				evaluatedSource: "git::https://github.com/terraform-aws-modules/terraform-aws-lambda.git?ref=v6.0.0",
				version:         "6.0.0",
				sourceType:      SourceTypeGit,
			},
		},
		{
			name:   "git http repo with double slash path separator",
			source: "git::https://github.com/terraform-aws-modules/terraform-aws-lambda.git//?ref=v6.0.0",
			expectedModule: terragruntModuleSource{
				protocol:        "git::https",
				baseUrl:         "github.com/terraform-aws-modules/terraform-aws-lambda.git",
				rawSource:       "git::https://github.com/terraform-aws-modules/terraform-aws-lambda.git//?ref=v6.0.0",
				evaluatedSource: "git::https://github.com/terraform-aws-modules/terraform-aws-lambda.git//?ref=v6.0.0",
				version:         "6.0.0",
				sourceType:      SourceTypeGit,
			},
		},
		{
			name:   "git http repo with double slash and submodule",
			source: "git::https://github.com/terraform-aws-modules/terraform-aws-eks.git//modules/karpenter?ref=v20.0.0",
			expectedModule: terragruntModuleSource{
				protocol:        "git::https",
				baseUrl:         "github.com/terraform-aws-modules/terraform-aws-eks.git",
				rawSource:       "git::https://github.com/terraform-aws-modules/terraform-aws-eks.git//modules/karpenter?ref=v20.0.0",
				evaluatedSource: "git::https://github.com/terraform-aws-modules/terraform-aws-eks.git//modules/karpenter?ref=v20.0.0",
				version:         "20.0.0",
				sourceType:      SourceTypeGit,
			},
		},
		{
			name:   "github repo",
			source: "github.com/gruntwork-io/terraform-google-network.git//modules/vpc-network?ref=v0.2.9",
			expectedModule: terragruntModuleSource{
				baseUrl:         "github.com/gruntwork-io/terraform-google-network.git//modules/vpc-network",
				rawSource:       "github.com/gruntwork-io/terraform-google-network.git//modules/vpc-network?ref=v0.2.9",
				evaluatedSource: "github.com/gruntwork-io/terraform-google-network.git//modules/vpc-network?ref=v0.2.9",
				version:         "0.2.9",
				sourceType:      SourceTypeGithub,
			},
		},
		{
			name:   "local path",
			source: "./random_module",
			expectedModule: terragruntModuleSource{
				sourceType:      SourceTypeLocal,
				rawSource:       "./random_module",
				evaluatedSource: "./random_module",
			},
		},
		{
			name:   "http module",
			source: "https://example.com/vpc-module?archive=zip",
			expectedModule: terragruntModuleSource{
				sourceType:      SourceTypeHttp,
				rawSource:       "https://example.com/vpc-module?archive=zip",
				evaluatedSource: "https://example.com/vpc-module?archive=zip",
			},
		},
		{
			name:   "mercurial module",
			source: "hg::http://example.com/vpc.hg",
			expectedModule: terragruntModuleSource{
				sourceType:      SourceTypeMercurial,
				rawSource:       "hg::http://example.com/vpc.hg",
				evaluatedSource: "hg::http://example.com/vpc.hg",
			},
		},
		{
			name:   "s3 module",
			source: "s3::https://s3-eu-west-1.amazonaws.com/examplecorp-terraform-modules/vpc.zip",
			expectedModule: terragruntModuleSource{
				sourceType:      SourceTypeS3,
				rawSource:       "s3::https://s3-eu-west-1.amazonaws.com/examplecorp-terraform-modules/vpc.zip",
				evaluatedSource: "s3::https://s3-eu-west-1.amazonaws.com/examplecorp-terraform-modules/vpc.zip",
			},
		},
		{
			name:   "gcs module",
			source: "gcs::https://www.googleapis.com/storage/v1/modules/foomodule.zip",
			expectedModule: terragruntModuleSource{
				sourceType:      SourceTypeGCS,
				rawSource:       "gcs::https://www.googleapis.com/storage/v1/modules/foomodule.zip",
				evaluatedSource: "gcs::https://www.googleapis.com/storage/v1/modules/foomodule.zip",
			},
		},
	}
	for _, d := range dataset {
		t.Run(d.name, func(t *testing.T) {
			module, err := parseSourceUrl(d.source, d.source, false)
			require.NoError(t, err)
			assert.Equal(t, module, d.expectedModule)
		})
	}
}

func TestIsLocalSourceUrl(t *testing.T) {
	dataset := []struct {
		source string
		local  bool
	}{
		{
			source: "tfr://registry.terraform.io/terraform-aws-modules/vpc/aws?version=3.3.0",
			local:  false,
		},
		{
			source: "tfr:///terraform-aws-modules/vpc/aws?version=3.3.0",
			local:  false,
		},
		{
			source: "git::https://github.com/terraform-aws-modules/terraform-aws-lambda.git?ref=v6.0.0",
			local:  false,
		},
		{
			source: "git@github.com:gruntwork-io/terraform-google-network.git//iam-group?ref=v1.2.0",
			local:  false,
		},
		{
			source: "github.com/gruntwork-io/terraform-google-network.git//modules/vpc-network?ref=v0.2.9",
			local:  false,
		},
		{
			source: "https://example.com/vpc-module?archive=zip",
			local:  false,
		},
		{
			source: "hg::http://example.com/vpc.hg",
			local:  false,
		},
		{
			source: "s3::https://s3-eu-west-1.amazonaws.com/examplecorp-terraform-modules/vpc.zip",
			local:  false,
		},
		{
			source: "gcs::https://www.googleapis.com/storage/v1/modules/foomodule.zip",
			local:  false,
		},
		{
			source: "./random_module",
			local:  true,
		},
		{
			source: "../../random_module",
			local:  true,
		},
		{
			source: "/random_module",
			local:  true,
		},
		{
			source: "~/random_module",
			local:  true,
		},
		{
			source: ".\\random_module",
			local:  true,
		},
		{
			source: "..\\..\\random_module",
			local:  true,
		},
	}
	for _, d := range dataset {
		t.Run(d.source, func(t *testing.T) {
			assert.Equal(t, isModuleSourceLocal(d.source), d.local)
		})
	}
}

func TestGetSourceType(t *testing.T) {
	dataset := []struct {
		source      string
		source_type string
	}{
		{
			source:      "tfr://registry.terraform.io/terraform-aws-modules/vpc/aws?version=3.3.0",
			source_type: SourceTypeRegistry,
		},
		{
			source:      "tfr:///terraform-aws-modules/vpc/aws?version=3.3.0",
			source_type: SourceTypeRegistry,
		},
		{
			source:      "git::https://github.com/terraform-aws-modules/terraform-aws-lambda.git?ref=v6.0.0",
			source_type: SourceTypeGit,
		},
		{
			source:      "git@github.com:gruntwork-io/terraform-google-network.git//iam-group?ref=v1.2.0",
			source_type: SourceTypeGit,
		},
		{
			source:      "github.com/gruntwork-io/terraform-google-network.git//modules/vpc-network?ref=v0.2.9",
			source_type: SourceTypeGithub,
		},
		{
			source:      "https://example.com/vpc-module?archive=zip",
			source_type: SourceTypeHttp,
		},
		{
			source:      "hg::http://example.com/vpc.hg",
			source_type: SourceTypeMercurial,
		},
		{
			source:      "s3::https://s3-eu-west-1.amazonaws.com/examplecorp-terraform-modules/vpc.zip",
			source_type: SourceTypeS3,
		},
		{
			source:      "gcs::https://www.googleapis.com/storage/v1/modules/foomodule.zip",
			source_type: SourceTypeGCS,
		},
		{
			source:      "./random_module",
			source_type: SourceTypeLocal,
		},
		{
			source:      "unknown",
			source_type: "",
		},
	}
	for _, d := range dataset {
		t.Run(d.source, func(t *testing.T) {
			assert.Equal(t, getSourceType(d.source), d.source_type)
		})
	}
}

func TestHclExpr(t *testing.T) {
	dataset := []struct {
		name   string
		file   string
		result string
	}{
		{
			name:   "No Expression",
			file:   "testdata/inlined.hcl",
			result: "tfr:///terraform-aws-modules/rdss/aws?version=5.8.1",
		},
		{
			name:   "Simple localized scenario",
			file:   "testdata/simple_localized.hcl",
			result: "tfr:///terraform-aws-modules/aurora/aws?version=5.8.1",
		},
		{
			name:   "Complex localized scenario",
			file:   "testdata/complex_localized.hcl",
			result: "tfr:///terraform-aws-modules/vpc/aws?version=5.8.1",
		},
	}

	for _, d := range dataset {
		t.Run(d.name, func(t *testing.T) {
			data, err := os.ReadFile(d.file)
			assert.NoError(t, err)
			hclfile, err := terraformUtils.ParseHcl(string(data), d.file)
			assert.NoError(t, err)
			var source string

			for _, block := range hclfile.Body().Blocks() {
				if block.Type() == "terraform" {
					source = strings.TrimSpace(string(block.Body().GetAttribute("source").Expr().BuildTokens(nil).Bytes()))
					break
				}
			}
			assert.NotNil(t, source)
			result, _, err := evaluateHcl(source, data, d.file)
			require.NoError(t, err)
			assert.Equal(t, result, d.result)
		})
	}
}
