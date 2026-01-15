package terragrunt

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscoverManifests(t *testing.T) {
	expectedInlined := `name: 'Bump Terragrunt module terraform-aws-modules/rdss/aws version'
sources:
  latestVersion:
    name: 'Get latest version of the terraform-aws-modules/rdss/aws module'
    kind: 'terraform/registry'
    transformers:
      - addprefix: 'tfr:///terraform-aws-modules/rdss/aws?version='
    spec:
      versionfilter:
        kind: 'semver'
        pattern: '>=5.8.1'
      type: 'module'
      hostname: 'registry.terraform.io'
      namespace: 'terraform-aws-modules'
      name: 'rdss'
      targetsystem: 'aws'
targets:
  terragruntModuleFile:
    name: 'deps: bump terraform-aws-modules/rdss/aws to {{ source "latestVersion" }}'
    kind: 'hcl'
    sourceid: 'latestVersion'
    spec:
      file: 'inlined.hcl'
      path: 'terraform.source'
`
	expectedSimpleLocalized := `name: 'Bump Terragrunt module terraform-aws-modules/aurora/aws version'
sources:
  latestVersion:
    name: 'Get latest version of the terraform-aws-modules/aurora/aws module'
    kind: 'terraform/registry'
    transformers:
      - addprefix: 'tfr:///terraform-aws-modules/aurora/aws?version='
    spec:
      versionfilter:
        kind: 'semver'
        pattern: '>=5.8.1'
      type: 'module'
      hostname: 'registry.terraform.io'
      namespace: 'terraform-aws-modules'
      name: 'aurora'
      targetsystem: 'aws'
targets:
  terragruntModuleFile:
    name: 'deps: bump terraform-aws-modules/aurora/aws to {{ source "latestVersion" }}'
    kind: 'hcl'
    sourceid: 'latestVersion'
    spec:
      file: 'simple_localized.hcl'
      path: 'locals.base_source_url'
`

	expectedComplexLocalized := `name: 'Bump Terragrunt module terraform-aws-modules/vpc/aws version'
sources:
  latestVersion:
    name: 'Get latest version of the terraform-aws-modules/vpc/aws module'
    kind: 'terraform/registry'
    spec:
      versionfilter:
        kind: 'semver'
        pattern: '>=5.8.1'
      type: 'module'
      hostname: 'registry.terraform.io'
      namespace: 'terraform-aws-modules'
      name: 'vpc'
      targetsystem: 'aws'
targets:
  terragruntModuleFile:
    name: 'deps: bump terraform-aws-modules/vpc/aws to {{ source "latestVersion" }}'
    kind: 'hcl'
    sourceid: 'latestVersion'
    spec:
      file: 'complex_localized.hcl'
      path: 'locals.module_version'
`
	expectedSuperComplexLocalized := `name: 'Bump Terragrunt module terraform-aws-modules/auroravpc/aws version'
sources:
  latestVersion:
    name: 'Get latest version of the terraform-aws-modules/auroravpc/aws module'
    kind: 'terraform/registry'
    transformers:
      - addprefix: 'tfr:///${local.module}?version='
    spec:
      versionfilter:
        kind: 'semver'
        pattern: '>=1.2.3'
      type: 'module'
      hostname: 'registry.terraform.io'
      namespace: 'terraform-aws-modules'
      name: 'auroravpc'
      targetsystem: 'aws'
targets:
  terragruntModuleFile:
    name: 'deps: bump terraform-aws-modules/auroravpc/aws to {{ source "latestVersion" }}'
    kind: 'hcl'
    sourceid: 'latestVersion'
    spec:
      file: 'more_complex_localized.hcl'
      path: 'terraform.source'
`

	expectedNonTfr := `name: 'Bump Terragrunt module github.com/Azure/terraform-azurerm-avm-res-network-virtualnetwork.git version'
scms:
  module:
    kind: 'git'
    spec:
      url: 'https://github.com/Azure/terraform-azurerm-avm-res-network-virtualnetwork.git'
      username: 'oauth2'
sources:
  latestVersion:
    name: 'Get latest version of the github.com/Azure/terraform-azurerm-avm-res-network-virtualnetwork.git module'
    kind: 'gittag'
    transformers:
      - addprefix: 'git::https://github.com/Azure/terraform-azurerm-avm-res-network-virtualnetwork.git?ref='
    spec:
      versionfilter:
        kind: 'semver'
        pattern: '>=0.3.0'
    scmid: 'module'
targets:
  terragruntModuleFile:
    name: 'deps: bump github.com/Azure/terraform-azurerm-avm-res-network-virtualnetwork.git to {{ source "latestVersion" }}'
    kind: 'hcl'
    sourceid: 'latestVersion'
    spec:
      file: 'non_tfr.hcl'
      path: 'terraform.source'
`

	testdata := []struct {
		name              string
		rootDir           string
		expectedPipelines []string
	}{
		{
			name:    "Terraform Version",
			rootDir: "testdata",
			expectedPipelines: []string{
				expectedNonTfr,
				expectedSimpleLocalized,
				expectedSuperComplexLocalized,
				expectedInlined,
				expectedComplexLocalized,
			},
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			resource, err := New(
				Spec{}, tt.rootDir, "", "")
			require.NoError(t, err)

			bytesPipelines, err := resource.DiscoverManifests()
			require.NoError(t, err)

			if len(bytesPipelines) != len(tt.expectedPipelines) {
				t.Logf("%v pipeline detected but expecting %v", len(bytesPipelines), len(tt.expectedPipelines))
				t.Fail()
				return
			}

			pipelines := []string{}
			for i := range bytesPipelines {
				pipelines = append(pipelines, string(bytesPipelines[i]))
			}

			sort.Strings(pipelines)

			for i := range tt.expectedPipelines {
				assert.Equal(t, tt.expectedPipelines[i], pipelines[i])
			}
		})
	}
}
