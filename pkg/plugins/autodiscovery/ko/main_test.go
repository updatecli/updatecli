package ko

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscoverManifests(t *testing.T) {
	// Disable condition testing with running short test
	if testing.Short() {
		return
	}

	testdata := []struct {
		name              string
		rootDir           string
		expectedPipelines []string
		digest            bool
		registryMappings  map[string]string
	}{
		{
			name:    "Scenario 1 no digest",
			rootDir: "testdata/success",
			expectedPipelines: []string{`name: 'deps: bump container image "golang"'
sources:
  golang:
    name: 'get latest container image tag for "golang"'
    kind: 'dockerimage'
    spec:
      image: 'golang'
      tagfilter: '^\d*(\.\d*){2}$'
      versionfilter:
        kind: 'semver'
        pattern: '>=1.19.0'
targets:
  golang:
    name: 'deps: bump container image "golang" to {{ source "golang" }}'
    kind: 'yaml'
    spec:
      file: '.ko.yaml'
      key: "$.baseImageOverrides.'github.com/google/ko'"
    sourceid: 'golang'
    transformers:
      - addprefix: 'golang:'
`},
		},
		{
			name:    "Scenario 2 digest",
			rootDir: "testdata/success",
			digest:  true,
			expectedPipelines: []string{`name: 'deps: bump container image "golang"'
sources:
  golang:
    name: 'get latest container image tag for "golang"'
    kind: 'dockerimage'
    spec:
      image: 'golang'
      tagfilter: '^\d*(\.\d*){2}$'
      versionfilter:
        kind: 'semver'
        pattern: '>=1.19.0'
  golang-digest:
    name: 'get latest container image digest for "golang:1.19.0"'
    kind: 'dockerdigest'
    spec:
      image: 'golang'
      tag: '{{ source "golang" }}'
    dependson:
      - 'golang'
targets:
  golang:
    name: 'deps: bump container image digest for "golang:1.19.0"'
    kind: 'yaml'
    spec:
      file: '.ko.yaml'
      key: "$.baseImageOverrides.'github.com/google/ko'"
    sourceid: 'golang-digest'
    transformers:
      - addprefix: 'golang:'
`},
		},
		{
			name:    "Scenario 3 updating digest",
			rootDir: "testdata/digest",
			digest:  true,
			expectedPipelines: []string{`name: 'deps: bump container image "golang"'
sources:
  golang:
    name: 'get latest container image tag for "golang"'
    kind: 'dockerimage'
    spec:
      image: 'golang'
      tagfilter: '^\d*(\.\d*){2}$'
      versionfilter:
        kind: 'semver'
        pattern: '>=1.22.0'
  golang-digest:
    name: 'get latest container image digest for "golang:1.22.0"'
    kind: 'dockerdigest'
    spec:
      image: 'golang'
      tag: '{{ source "golang" }}'
    dependson:
      - 'golang'
targets:
  golang:
    name: 'deps: bump container image digest for "golang:1.22.0"'
    kind: 'yaml'
    spec:
      file: '.ko.yaml'
      key: "$.baseImageOverrides.'github.com/google/ko'"
    sourceid: 'golang-digest'
    transformers:
      - addprefix: 'golang:'
`},
		},
		{
			name:             "Scenario 4 registry mappings",
			rootDir:          "testdata/registryMappings",
			digest:           false,
			registryMappings: map[string]string{"my-registry.com/ghcr-io/": "ghcr.io/"},
			expectedPipelines: []string{`name: 'deps: bump container image "my-registry.com/ghcr-io/updatecli/updatecli"'
sources:
  my-registry.com/ghcr-io/updatecli/updatecli:
    name: 'get latest container image tag for "ghcr.io/updatecli/updatecli"'
    kind: 'dockerimage'
    spec:
      image: 'ghcr.io/updatecli/updatecli'
      tagfilter: '^v\d*(\.\d*){2}$'
      versionfilter:
        kind: 'semver'
        pattern: '>=v0.67.0'
conditions:
  my-registry.com/ghcr-io/updatecli/updatecli-mirror:
    name: 'check container image "my-registry.com/ghcr-io/updatecli/updatecli:{{ source "my-registry.com/ghcr-io/updatecli/updatecli" }}" is available on the mirror'
    kind: 'dockerimage'
    sourceid: 'my-registry.com/ghcr-io/updatecli/updatecli'
    spec:
      image: 'my-registry.com/ghcr-io/updatecli/updatecli'
targets:
  my-registry.com/ghcr-io/updatecli/updatecli:
    name: 'deps: bump container image "my-registry.com/ghcr-io/updatecli/updatecli" to {{ source "my-registry.com/ghcr-io/updatecli/updatecli" }}'
    kind: 'yaml'
    spec:
      file: '.ko.yaml'
      key: "$.defaultBaseImage"
    sourceid: 'my-registry.com/ghcr-io/updatecli/updatecli'
    transformers:
      - addprefix: 'my-registry.com/ghcr-io/updatecli/updatecli:'
`},
		},
		{
			name:             "Scenario 5 registry mappings with digest",
			rootDir:          "testdata/registryMappings",
			digest:           true,
			registryMappings: map[string]string{"my-registry.com/ghcr-io/": "ghcr.io/"},
			expectedPipelines: []string{`name: 'deps: bump container image "my-registry.com/ghcr-io/updatecli/updatecli"'
sources:
  my-registry.com/ghcr-io/updatecli/updatecli:
    name: 'get latest container image tag for "ghcr.io/updatecli/updatecli"'
    kind: 'dockerimage'
    spec:
      image: 'ghcr.io/updatecli/updatecli'
      tagfilter: '^v\d*(\.\d*){2}$'
      versionfilter:
        kind: 'semver'
        pattern: '>=v0.67.0'
  my-registry.com/ghcr-io/updatecli/updatecli-digest:
    name: 'get latest container image digest for "ghcr.io/updatecli/updatecli:v0.67.0"'
    kind: 'dockerdigest'
    spec:
      image: 'ghcr.io/updatecli/updatecli'
      tag: '{{ source "my-registry.com/ghcr-io/updatecli/updatecli" }}'
    dependson:
      - 'my-registry.com/ghcr-io/updatecli/updatecli'
conditions:
  my-registry.com/ghcr-io/updatecli/updatecli-mirror:
    name: 'check container image "my-registry.com/ghcr-io/updatecli/updatecli:{{ source "my-registry.com/ghcr-io/updatecli/updatecli-digest" }}" is available on the mirror'
    kind: 'dockerimage'
    sourceid: 'my-registry.com/ghcr-io/updatecli/updatecli-digest'
    spec:
      image: 'my-registry.com/ghcr-io/updatecli/updatecli'
targets:
  my-registry.com/ghcr-io/updatecli/updatecli:
    name: 'deps: bump container image digest for "my-registry.com/ghcr-io/updatecli/updatecli:v0.67.0"'
    kind: 'yaml'
    spec:
      file: '.ko.yaml'
      key: "$.defaultBaseImage"
    sourceid: 'my-registry.com/ghcr-io/updatecli/updatecli-digest'
    transformers:
      - addprefix: 'my-registry.com/ghcr-io/updatecli/updatecli:'
`},
		},
	}

	for _, tt := range testdata {

		t.Run(tt.name, func(t *testing.T) {
			digest := tt.digest

			pod, err := New(
				Spec{
					Digest:           &digest,
					RegistryMappings: tt.registryMappings,
				}, tt.rootDir, "", "")

			require.NoError(t, err)

			rawPipelines, err := pod.DiscoverManifests()
			require.NoError(t, err)

			assert.Equal(t, len(tt.expectedPipelines), len(rawPipelines))

			for i := range tt.expectedPipelines {
				assert.Equal(t, tt.expectedPipelines[i], string(rawPipelines[i]))
			}
		})
	}
}
