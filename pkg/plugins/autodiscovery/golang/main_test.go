package golang

import (
	"sort"
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
			name:    "Golang Replace module",
			rootDir: "testdata/replace",
			expectedPipelines: []string{`name: 'deps(go): bump module github.com/crewjam/saml'
sources:
  module:
    name: 'Get latest golang module github.com/crewjam/saml version'
    kind: 'golang/module'
    spec:
      module: 'github.com/crewjam/saml'
      versionfilter:
        kind: 'semver'
        pattern: '>=0.6.0'
targets:
  module:
    name: 'deps(go): bump module github.com/crewjam/saml to {{ source "module" }}'
    kind: 'golang/gomod'
    sourceid: 'module'
    spec:
      file: 'go.mod'
      module: 'github.com/crewjam/saml'
`,
				`name: 'deps(go): bump module github.com/rancher/saml'
sources:
  module:
    name: 'Get latest golang module github.com/rancher/saml version'
    kind: 'golang/module'
    spec:
      module: 'github.com/rancher/saml'
      versionfilter:
        kind: 'semver'
        pattern: '>=0.3.0'
targets:
  module:
    name: 'deps(go): bump module github.com/rancher/saml to {{ source "module" }}'
    kind: 'golang/gomod'
    sourceid: 'module'
    spec:
      file: 'go.mod'
      module: 'github.com/rancher/saml'
`,
				`name: 'deps(go): bump module github.com/stretchr/testify'
sources:
  module:
    name: 'Get latest golang module github.com/stretchr/testify version'
    kind: 'golang/module'
    spec:
      module: 'github.com/stretchr/testify'
      versionfilter:
        kind: 'semver'
        pattern: '>=1.8.4'
targets:
  module:
    name: 'deps(go): bump module github.com/stretchr/testify to {{ source "module" }}'
    kind: 'golang/gomod'
    sourceid: 'module'
    spec:
      file: 'go.mod'
      module: 'github.com/stretchr/testify'
`,
				`name: 'deps(go): bump replaced module github.com/crewjam/saml'
sources:
  module:
    name: 'Get latest golang module github.com/crewjam/saml version'
    kind: 'golang/module'
    spec:
      module: 'github.com/crewjam/saml'
      versionfilter:
        kind: 'semver'
        pattern: '>=0.5.0'
targets:
  module:
    name: 'deps(go): bump module github.com/crewjam/saml to {{ source "module" }}'
    kind: 'golang/gomod'
    sourceid: 'module'
    spec:
      file: 'go.mod'
      module: 'github.com/crewjam/saml'
      replace: true
      replaceVersion: 'v0.6.0'

`,
				`name: 'deps(go): bump replaced module github.com/rancher/saml'
sources:
  module:
    name: 'Get latest golang module github.com/rancher/saml version'
    kind: 'golang/module'
    spec:
      module: 'github.com/rancher/saml'
      versionfilter:
        kind: 'semver'
        pattern: '>=0.2.0'
targets:
  module:
    name: 'deps(go): bump module github.com/rancher/saml to {{ source "module" }}'
    kind: 'golang/gomod'
    sourceid: 'module'
    spec:
      file: 'go.mod'
      module: 'github.com/rancher/saml'
      replace: true
`,
				`name: 'deps(golang): bump Go version'
sources:
  go:
    name: 'Get latest Go version'
    kind: 'golang'
    spec:
      versionfilter:
        kind: 'semver'
        pattern: '>=1.25.0'
targets:
  go:
    name: 'deps(golang): bump Go version to {{ source "go" }}'
    kind: 'golang/gomod'
    sourceid: 'go'
    spec:
      file: 'go.mod'
`},
		},
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

			// sort stringsPipelines to ensure that the test result
			// is always the same
			sort.Slice(stringPipelines, func(i, j int) bool {
				return stringPipelines[i] < stringPipelines[j]
			})

			for i := range stringPipelines {
				pipelines = append(pipelines, stringPipelines...)
				assert.Equal(t, tt.expectedPipelines[i], pipelines[i])
			}
		})
	}
}
