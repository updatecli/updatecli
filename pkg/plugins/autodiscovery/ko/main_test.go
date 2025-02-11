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
	}

	for _, tt := range testdata {

		t.Run(tt.name, func(t *testing.T) {
			digest := tt.digest

			pod, err := New(
				Spec{
					Digest: &digest,
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
