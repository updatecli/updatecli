package fleet

import (
	"testing"

	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/cargo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/fleet"
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
	}{
		{
			name:    "Scenario 1",
			rootDir: "testdata",
			expectedPipelines: []string{`name: 'Bump dependencies "anyhow" for "test-crate" crate'
sources:
  anyhow:
    name: 'Get latest "anyhow" crate version'
    kind: 'cargopackage'
    spec:
      package: 'anyhow'
      versionFilter:
        kind: 'semver'
        pattern: '*'
  anyhow-current-version:
    name: 'Get current "anyhow" crate version'
    kind: 'toml'
    spec:
      file: 'Cargo.toml'
      Key: 'dependencies.anyhow'
conditions:
  anyhow:
    name: 'Ensure Cargo chart named "anyhow" is specified'
    kind: 'toml'
    spec:
      file: 'Cargo.toml'
      query: 'dependencies.(?:-=anyhow)'
    sourceid: 'anyhow-current-version'
targets:
  anyhow:
    name: 'Bump crate dependency "anyhow" for crate "test-crate"'
    kind: 'toml'
    spec:
      file: 'Cargo.toml'
      key: 'dependencies.anyhow'
    sourceid: 'anyhow'
`, `name: 'Bump dependencies "rand" for "test-crate" crate'
sources:
  rand:
    name: 'Get latest "rand" crate version'
    kind: 'cargopackage'
    spec:
      package: 'rand'
      versionFilter:
        kind: 'semver'
        pattern: '*'
  rand-current-version:
    name: 'Get current "rand" crate version'
    kind: 'toml'
    spec:
      file: 'Cargo.toml'
      Key: 'dependencies.rand.version'
conditions:
  rand:
    name: 'Ensure Cargo chart named "rand" is specified'
    kind: 'toml'
    spec:
      file: 'Cargo.toml'
      query: 'dependencies.(?:-=rand).version'
    sourceid: 'rand-current-version'
targets:
  rand:
    name: 'Bump crate dependency "rand" for crate "test-crate"'
    kind: 'toml'
    spec:
      file: 'Cargo.toml'
      key: 'dependencies.rand.version'
    sourceid: 'rand'
`, `name: 'Bump dev-dependencies "futures" for "test-crate" crate'
sources:
  futures:
    name: 'Get latest "futures" crate version'
    kind: 'cargopackage'
    spec:
      package: 'futures'
      versionFilter:
        kind: 'semver'
        pattern: '*'
  futures-current-version:
    name: 'Get current "futures" crate version'
    kind: 'toml'
    spec:
      file: 'Cargo.toml'
      Key: 'dev-dependencies.futures.version'
conditions:
  futures:
    name: 'Ensure Cargo chart named "futures" is specified'
    kind: 'toml'
    spec:
      file: 'Cargo.toml'
      query: 'dev-dependencies.(?:-=futures).version'
    sourceid: 'futures-current-version'
targets:
  futures:
    name: 'Bump crate dependency "futures" for crate "test-crate"'
    kind: 'toml'
    spec:
      file: 'Cargo.toml'
      key: 'dev-dependencies.futures.version'
    sourceid: 'futures'
`},
		},
	}

	for _, tt := range testdata {

		t.Run(tt.name, func(t *testing.T) {
			c, err := cargo.New(
				fleet.Spec{
					RootDir: tt.rootDir,
				}, "", "")

			require.NoError(t, err)

			var pipelines []string
			rawPipelines, err := c.DiscoverManifests()
			require.NoError(t, err)

			assert.Equal(t, len(rawPipelines), len(tt.expectedPipelines))

			for i := range rawPipelines {
				// We expect manifest generated by the autodiscovery to use the yaml syntax
				pipelines = append(pipelines, string(rawPipelines[i]))
				assert.Equal(t, tt.expectedPipelines[i], pipelines[i])
			}
		})
	}

}
