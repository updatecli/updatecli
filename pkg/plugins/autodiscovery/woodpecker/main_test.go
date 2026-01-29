package woodpecker

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscoverManifests(t *testing.T) {

	testdata := []struct {
		name              string
		rootDir           string
		digest            bool
		expectedPipelines []string
	}{
		{
			name:    "Scenario 1 - simple steps format with digest",
			rootDir: "testdata/simple",
			digest:  true,
			expectedPipelines: []string{`name: 'deps(woodpecker): bump "golang" digest'
sources:
  build:
    name: 'get latest image tag for "golang"'
    kind: 'dockerimage'
    spec:
      image: 'golang'
      tagfilter: '^\d*(\.\d*){1}$'
      versionfilter:
        kind: 'semver'
        pattern: '>=1.21'
  build-digest:
    name: 'get latest image "golang" digest'
    kind: 'dockerdigest'
    spec:
      image: 'golang'
      tag: '{{ source "build" }}'
    dependson:
      - 'build'
targets:
  build:
    name: 'deps: update Woodpecker image "golang" to "{{ source "build" }}"'
    kind: 'yaml'
    spec:
      file: '.woodpecker.yml'
      key: '$.steps[0].image'
    sourceid: 'build-digest'
    transformers:
      - addprefix: 'golang:'
`, `name: 'deps(woodpecker): bump "golang" digest'
sources:
  test:
    name: 'get latest image tag for "golang"'
    kind: 'dockerimage'
    spec:
      image: 'golang'
      tagfilter: '^\d*(\.\d*){1}$'
      versionfilter:
        kind: 'semver'
        pattern: '>=1.21'
  test-digest:
    name: 'get latest image "golang" digest'
    kind: 'dockerdigest'
    spec:
      image: 'golang'
      tag: '{{ source "test" }}'
    dependson:
      - 'test'
targets:
  test:
    name: 'deps: update Woodpecker image "golang" to "{{ source "test" }}"'
    kind: 'yaml'
    spec:
      file: '.woodpecker.yml'
      key: '$.steps[1].image'
    sourceid: 'test-digest'
    transformers:
      - addprefix: 'golang:'
`,
			},
		},
		{
			name:    "Scenario 2 - simple steps format without digest",
			rootDir: "testdata/simple",
			digest:  false,
			expectedPipelines: []string{`name: 'deps(woodpecker): bump "golang" tag'
sources:
  build:
    name: 'get latest image tag for "golang"'
    kind: 'dockerimage'
    spec:
      image: 'golang'
      tagfilter: '^\d*(\.\d*){1}$'
      versionfilter:
        kind: 'semver'
        pattern: '>=1.21'
targets:
  build:
    name: 'deps: update Woodpecker image "golang" to "{{ source "build" }}"'
    kind: 'yaml'
    spec:
      file: '.woodpecker.yml'
      key: '$.steps[0].image'
    sourceid: 'build'
    transformers:
      - addprefix: 'golang:'
`, `name: 'deps(woodpecker): bump "golang" tag'
sources:
  test:
    name: 'get latest image tag for "golang"'
    kind: 'dockerimage'
    spec:
      image: 'golang'
      tagfilter: '^\d*(\.\d*){1}$'
      versionfilter:
        kind: 'semver'
        pattern: '>=1.21'
targets:
  test:
    name: 'deps: update Woodpecker image "golang" to "{{ source "test" }}"'
    kind: 'yaml'
    spec:
      file: '.woodpecker.yml'
      key: '$.steps[1].image'
    sourceid: 'test'
    transformers:
      - addprefix: 'golang:'
`,
			},
		},
		{
			name:    "Scenario 3 - legacy pipeline format",
			rootDir: "testdata/legacy",
			digest:  false,
			expectedPipelines: []string{`name: 'deps(woodpecker): bump "golang" tag'
sources:
  build:
    name: 'get latest image tag for "golang"'
    kind: 'dockerimage'
    spec:
      image: 'golang'
      tagfilter: '^\d*(\.\d*){1}$'
      versionfilter:
        kind: 'semver'
        pattern: '>=1.20'
targets:
  build:
    name: 'deps: update Woodpecker image "golang" to "{{ source "build" }}"'
    kind: 'yaml'
    spec:
      file: '.woodpecker.yml'
      key: '$.pipeline.build.image'
    sourceid: 'build'
    transformers:
      - addprefix: 'golang:'
`, `name: 'deps(woodpecker): bump "golang" tag'
sources:
  test:
    name: 'get latest image tag for "golang"'
    kind: 'dockerimage'
    spec:
      image: 'golang'
      tagfilter: '^\d*(\.\d*){1}$'
      versionfilter:
        kind: 'semver'
        pattern: '>=1.20'
targets:
  test:
    name: 'deps: update Woodpecker image "golang" to "{{ source "test" }}"'
    kind: 'yaml'
    spec:
      file: '.woodpecker.yml'
      key: '$.pipeline.test.image'
    sourceid: 'test'
    transformers:
      - addprefix: 'golang:'
`,
			},
		},
		{
			name:    "Scenario 4 - services",
			rootDir: "testdata/services",
			digest:  false,
			expectedPipelines: []string{`name: 'deps(woodpecker): bump "python" tag'
sources:
  test:
    name: 'get latest image tag for "python"'
    kind: 'dockerimage'
    spec:
      image: 'python'
      tagfilter: '^\d*(\.\d*){1}$'
      versionfilter:
        kind: 'semver'
        pattern: '>=3.11'
targets:
  test:
    name: 'deps: update Woodpecker image "python" to "{{ source "test" }}"'
    kind: 'yaml'
    spec:
      file: '.woodpecker.yml'
      key: '$.steps[0].image'
    sourceid: 'test'
    transformers:
      - addprefix: 'python:'
`, `name: 'deps(woodpecker): bump "postgres" tag'
sources:
  service-database:
    name: 'get latest image tag for "postgres"'
    kind: 'dockerimage'
    spec:
      image: 'postgres'
      tagfilter: '^\d*$'
      versionfilter:
        kind: 'semver'
        pattern: '>=15'
targets:
  service-database:
    name: 'deps: update Woodpecker image "postgres" to "{{ source "service-database" }}"'
    kind: 'yaml'
    spec:
      file: '.woodpecker.yml'
      key: '$.services[0].image'
    sourceid: 'service-database'
    transformers:
      - addprefix: 'postgres:'
`, `name: 'deps(woodpecker): bump "redis" tag'
sources:
  service-cache:
    name: 'get latest image tag for "redis"'
    kind: 'dockerimage'
    spec:
      image: 'redis'
      tagfilter: '^\d*-alpine$'
      versionfilter:
        kind: 'semver'
        pattern: '>=7-alpine'
targets:
  service-cache:
    name: 'deps: update Woodpecker image "redis" to "{{ source "service-cache" }}"'
    kind: 'yaml'
    spec:
      file: '.woodpecker.yml'
      key: '$.services[1].image'
    sourceid: 'service-cache'
    transformers:
      - addprefix: 'redis:'
`,
			},
		},
		{
			name:    "Scenario 5 - directory format",
			rootDir: "testdata/directory",
			digest:  false,
			expectedPipelines: []string{`name: 'deps(woodpecker): bump "node" tag'
sources:
  build:
    name: 'get latest image tag for "node"'
    kind: 'dockerimage'
    spec:
      image: 'node'
      tagfilter: '^\d*-alpine$'
      versionfilter:
        kind: 'semver'
        pattern: '>=18-alpine'
targets:
  build:
    name: 'deps: update Woodpecker image "node" to "{{ source "build" }}"'
    kind: 'yaml'
    spec:
      file: '.woodpecker/build.yml'
      key: '$.steps[0].image'
    sourceid: 'build'
    transformers:
      - addprefix: 'node:'
`, `name: 'deps(woodpecker): bump "node" tag'
sources:
  lint:
    name: 'get latest image tag for "node"'
    kind: 'dockerimage'
    spec:
      image: 'node'
      tagfilter: '^\d*-alpine$'
      versionfilter:
        kind: 'semver'
        pattern: '>=18-alpine'
targets:
  lint:
    name: 'deps: update Woodpecker image "node" to "{{ source "lint" }}"'
    kind: 'yaml'
    spec:
      file: '.woodpecker/build.yml'
      key: '$.steps[1].image'
    sourceid: 'lint'
    transformers:
      - addprefix: 'node:'
`,
			},
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			digest := tt.digest
			woodpecker, err := New(
				Spec{
					Digest: &digest,
				}, tt.rootDir, "", "")

			require.NoError(t, err)

			rawPipelines, err := woodpecker.DiscoverManifests()
			require.NoError(t, err)

			if len(rawPipelines) == 0 {
				t.Errorf("No pipelines found for %s", tt.name)
			}

			var pipelines []string
			assert.Equal(t, len(tt.expectedPipelines), len(rawPipelines))

			for i := range rawPipelines {
				pipelines = append(pipelines, string(rawPipelines[i]))
				assert.Equal(t, tt.expectedPipelines[i], pipelines[i])
			}
		})
	}
}

func TestNew(t *testing.T) {
	testdata := []struct {
		name          string
		spec          Spec
		rootDir       string
		expectedError bool
	}{
		{
			name:    "Default spec",
			spec:    Spec{},
			rootDir: "testdata",
		},
		{
			name: "Custom FileMatch",
			spec: Spec{
				FileMatch: []string{"*.yml"},
			},
			rootDir: "testdata",
		},
		{
			name:          "Empty rootDir",
			spec:          Spec{},
			rootDir:       "",
			expectedError: false, // New() returns empty Woodpecker but no error
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			w, err := New(tt.spec, tt.rootDir, "", "")
			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.rootDir != "" {
					assert.NotEmpty(t, w.rootDir)
				}
			}
		})
	}
}
