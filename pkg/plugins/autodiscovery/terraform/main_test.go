package terraform

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscoverManifests(t *testing.T) {
	expectedAWS := `name: 'Bump Terraform provider hashicorp/aws version'
sources:
  latestVersion:
    name: 'Get latest version of the hashicorp/aws provider'
    kind: terraform/registry
    spec:
      type: provider
      namespace: hashicorp
      name: aws
      versionfilter:
        kind: 'semver'
        pattern: '>=5.9.0'
targets:
  terraformLockVersion:
    name: Bump hashicorp/aws to {{ source "latestVersion" }}
    kind: terraform/lock
    sourceid: latestVersion
    spec:
      file: '.terraform.lock.hcl'
      provider: 'hashicorp/aws'
      platforms:
        - linux_amd64
        - linux_arm64
        - darwin_amd64
        - darwin_arm64
`

	expectedCloudInit := `name: 'Bump Terraform provider hashicorp/cloudinit version'
sources:
  latestVersion:
    name: 'Get latest version of the hashicorp/cloudinit provider'
    kind: terraform/registry
    spec:
      type: provider
      namespace: hashicorp
      name: cloudinit
      versionfilter:
        kind: 'semver'
        pattern: '>=2.3.2'
targets:
  terraformLockVersion:
    name: Bump hashicorp/cloudinit to {{ source "latestVersion" }}
    kind: terraform/lock
    sourceid: latestVersion
    spec:
      file: '.terraform.lock.hcl'
      provider: 'hashicorp/cloudinit'
      platforms:
        - linux_amd64
        - linux_arm64
        - darwin_amd64
        - darwin_arm64
`
	testdata := []struct {
		name              string
		rootDir           string
		platforms         []string
		only              MatchingRules
		ignore            MatchingRules
		expectedPipelines []string
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
			expectedPipelines: []string{
				expectedAWS,
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
			only: MatchingRules{
				MatchingRule{
					Providers: map[string]string{
						"registry.terraform.io/hashicorp/aws": "",
					},
				},
			},
			expectedPipelines: []string{
				expectedAWS,
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
			ignore: MatchingRules{
				MatchingRule{
					Providers: map[string]string{
						"registry.terraform.io/hashicorp/aws": "",
					},
				},
			},
			expectedPipelines: []string{
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
			only: MatchingRules{
				MatchingRule{
					Providers: map[string]string{
						"registry.terraform.io/hashicorp/aws":       ">6",
						"registry.terraform.io/hashicorp/cloudinit": ">=2",
					},
				},
			},
			expectedPipelines: []string{
				expectedCloudInit,
			},
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			resource, err := New(
				Spec{
					Platforms: tt.platforms,
					Only:      tt.only,
					Ignore:    tt.ignore,
				}, tt.rootDir, "", "")
			require.NoError(t, err)

			bytePipelines, err := resource.DiscoverManifests()
			require.NoError(t, err)

			if len(bytePipelines) != len(tt.expectedPipelines) {
				t.Logf("%v pipeline detected but expecting %v", len(bytePipelines), len(tt.expectedPipelines))
				t.Fail()
				return
			}

			pipelines := []string{}
			for i := range bytePipelines {
				pipelines = append(pipelines, string(bytePipelines[i]))
			}

			sort.Strings(pipelines)
			sort.Strings(tt.expectedPipelines)

			assert.Equal(t, len(tt.expectedPipelines), len(pipelines))

			for i := range tt.expectedPipelines {
				assert.Equal(t, tt.expectedPipelines[i], pipelines[i])
			}
		})
	}
}
