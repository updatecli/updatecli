package golang

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscoverManifests(t *testing.T) {

	testdata := []struct {
		name              string
		rootDir           string
		expectedPipelines []string
	}{
		{
			name:    "Golang Version",
			rootDir: "testdata/noModule",
			expectedPipelines: []string{`name: 'deps(go): bump module gopkg.in/yaml.v3'
sources:
  module:
    name: 'Get latest golang module gopkg.in/yaml.v3 version'
    kind: 'golang/module'
    spec:
      module: 'gopkg.in/yaml.v3'
      versionfilter:
        kind: 'semver'
        pattern: '>=3.0.1'
targets:
  module:
    name: 'deps(go): bump module gopkg.in/yaml.v3 to {{ source "module" }}'
    kind: 'golang/gomod'
    sourceid: 'module'
    spec:
      file: 'go.mod'
      module: 'gopkg.in/yaml.v3'
  tidy:
    name: 'clean: go mod tidy'
    disablesourceinput: true
    dependsonchange: true
    dependson:
      - 'module'
    kind: 'shell'
    spec:
      command: 'go mod tidy'
      environments:
        - name: HOME
        - name: PATH
      workdir: .
      changedif:
        kind: 'file/checksum'
        spec:
          files:
           - 'go.mod'
           - 'go.sum'
`, `name: 'deps(golang): bump Go version'
sources:
  go:
    name: 'Get latest Go version'
    kind: 'golang'
    spec:
      versionfilter:
        kind: 'semver'
        pattern: '>=1.20.0'
targets:
  go:
    name: 'deps(golang): bump Go version to {{ source "go" }}'
    kind: 'golang/gomod'
    sourceid: 'go'
    spec:
      file: 'go.mod'
`,
			},
		},
		{
			name:    "Golang Version",
			rootDir: "testdata/noSumFile",
			expectedPipelines: []string{`name: 'deps(go): bump module gopkg.in/yaml.v3'
sources:
  module:
    name: 'Get latest golang module gopkg.in/yaml.v3 version'
    kind: 'golang/module'
    spec:
      module: 'gopkg.in/yaml.v3'
      versionfilter:
        kind: 'semver'
        pattern: '>=3.0.1'
targets:
  module:
    name: 'deps(go): bump module gopkg.in/yaml.v3 to {{ source "module" }}'
    kind: 'golang/gomod'
    sourceid: 'module'
    spec:
      file: 'go.mod'
      module: 'gopkg.in/yaml.v3'
`, `name: 'deps(golang): bump Go version'
sources:
  go:
    name: 'Get latest Go version'
    kind: 'golang'
    spec:
      versionfilter:
        kind: 'semver'
        pattern: '>=1.20.0'
targets:
  go:
    name: 'deps(golang): bump Go version to {{ source "go" }}'
    kind: 'golang/gomod'
    sourceid: 'go'
    spec:
      file: 'go.mod'
`,
			},
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			resource, err := New(
				Spec{}, tt.rootDir, "", "")
			require.NoError(t, err)

			var pipelines []string
			bytesPipelines, err := resource.DiscoverManifests()
			require.NoError(t, err)

			assert.Equal(t, len(tt.expectedPipelines), len(bytesPipelines))

			stringPipelines := []string{}
			for i := range bytesPipelines {
				stringPipelines = append(stringPipelines, string(bytesPipelines[i]))
			}

			for i := range stringPipelines {
				pipelines = append(pipelines, stringPipelines...)
				assert.Equal(t, tt.expectedPipelines[i], pipelines[i])
			}
		})
	}
}
