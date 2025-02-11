package updatecli

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDiscoverManifests(t *testing.T) {
	testdata := []struct {
		name              string
		rootDir           string
		expectedPipelines []string
	}{
		{
			name:    "Scenario 1",
			rootDir: "testdata",
			expectedPipelines: []string{`name: 'deps(updatecli/policy): bump "ghcr.io/updatecli/policies/policies/hugo/netlify" Updatecli policy version'
sources:
  version:
    name: 'Get latest "ghcr.io/updatecli/policies/policies/hugo/netlify" Updatecli policy version'
    kind: 'dockerimage'
    spec:
      image: 'ghcr.io/updatecli/policies/policies/hugo/netlify'
      versionfilter:
        kind: 'semver'
        pattern: '*'
  digest:
    name: 'Get latest "ghcr.io/updatecli/policies/policies/hugo/netlify" Updatecli policy digest'
    kind: 'dockerdigest'
    dependson:
     - version
    spec:
      image: 'ghcr.io/updatecli/policies/policies/hugo/netlify'
      tag: '{{ source "version" }}'
targets:
  compose:
    name: 'deps(updatecli): bump "ghcr.io/updatecli/policies/policies/hugo/netlify" policy to {{ source "version"}}'
    kind: 'yaml'
    spec:
      file: 'testdata/website/updatecli-compose.yaml'
      key: '$.policies[3].policy'
    transformers:
      - addprefix: 'ghcr.io/updatecli/policies/policies/hugo/netlify:'
    sourceid: 'digest'
`, `name: 'deps(updatecli/policy): bump "ghcr.io/updatecli/policies/policies/nodejs/githubaction" Updatecli policy version'
sources:
  version:
    name: 'Get latest "ghcr.io/updatecli/policies/policies/nodejs/githubaction" Updatecli policy version'
    kind: 'dockerimage'
    spec:
      image: 'ghcr.io/updatecli/policies/policies/nodejs/githubaction'
      versionfilter:
        kind: 'semver'
        pattern: '*'
  digest:
    name: 'Get latest "ghcr.io/updatecli/policies/policies/nodejs/githubaction" Updatecli policy digest'
    kind: 'dockerdigest'
    dependson:
     - version
    spec:
      image: 'ghcr.io/updatecli/policies/policies/nodejs/githubaction'
      tag: '{{ source "version" }}'
targets:
  compose:
    name: 'deps(updatecli): bump "ghcr.io/updatecli/policies/policies/nodejs/githubaction" policy to {{ source "version"}}'
    kind: 'yaml'
    spec:
      file: 'testdata/website/updatecli-compose.yaml'
      key: '$.policies[1].policy'
    transformers:
      - addprefix: 'ghcr.io/updatecli/policies/policies/nodejs/githubaction:'
    sourceid: 'digest'
`, `name: 'deps(updatecli/policy): bump "ghcr.io/updatecli/policies/policies/nodejs/netlify" Updatecli policy version'
sources:
  version:
    name: 'Get latest "ghcr.io/updatecli/policies/policies/nodejs/netlify" Updatecli policy version'
    kind: 'dockerimage'
    spec:
      image: 'ghcr.io/updatecli/policies/policies/nodejs/netlify'
      versionfilter:
        kind: 'semver'
        pattern: '*'
  digest:
    name: 'Get latest "ghcr.io/updatecli/policies/policies/nodejs/netlify" Updatecli policy digest'
    kind: 'dockerdigest'
    dependson:
     - version
    spec:
      image: 'ghcr.io/updatecli/policies/policies/nodejs/netlify'
      tag: '{{ source "version" }}'
targets:
  compose:
    name: 'deps(updatecli): bump "ghcr.io/updatecli/policies/policies/nodejs/netlify" policy to {{ source "version"}}'
    kind: 'yaml'
    spec:
      file: 'testdata/website/updatecli-compose.yaml'
      key: '$.policies[2].policy'
    transformers:
      - addprefix: 'ghcr.io/updatecli/policies/policies/nodejs/netlify:'
    sourceid: 'digest'
`},
		},
	}

	for _, tt := range testdata {

		t.Run(tt.name, func(t *testing.T) {
			updatecli, err := New(
				Spec{
					Files: []string{"updatecli-compose.yaml"},
				}, tt.rootDir, "", "")
			require.NoError(t, err)

			bytesPipelines, err := updatecli.DiscoverManifests()
			require.NoError(t, err)

			//require.Equal(t, len(tt.expectedPipelines), len(bytesPipelines))

			pipelines := []string{}
			for i := range bytesPipelines {
				pipelines = append(pipelines, string(bytesPipelines[i]))
			}

			sort.Strings(pipelines)

			for i := range tt.expectedPipelines {
				require.Equal(t, tt.expectedPipelines[i], pipelines[i])
			}
		})
	}
}
