package cargo

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestCargoUpgradeCheckerExecutor struct {
	errorMessage string
}

func (c TestCargoUpgradeCheckerExecutor) Run() error {
	if c.errorMessage != "" {
		return errors.New(c.errorMessage)
	}
	return nil
}

func TestDiscoverManifests(t *testing.T) {
	// Disable condition testing with running short test
	if testing.Short() {
		return
	}

	testdata := []struct {
		name                  string
		rootDir               string
		cargoUpgradeAvailable bool
		expectedPipelines     []string
	}{
		{
			name:                  "Scenario 1 -- Cargo Upgrade available",
			rootDir:               "testdata",
			cargoUpgradeAvailable: true,
			expectedPipelines: []string{`name: 'deps(cargo): bump dependencies "anyhow" for "test-crate" crate'
sources:
  anyhow:
    name: 'Get latest "anyhow" crate version'
    kind: 'cargopackage'
    spec:
      package: 'anyhow'
      versionfilter:
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
    name: 'Test if version of "anyhow" {{ source "anyhow-current-version" }} differs from {{ source "anyhow" }}'
    kind: 'shell'
    sourceid: 'anyhow'
    spec:
      command: 'test {{ source "anyhow-current-version" }} != '
targets:
  anyhow:
    name: 'deps(cargo): bump crate dependency "anyhow" to {{ source "anyhow" }}'
    kind: 'shell'
    spec:
      command: |
        ARGS=""
        if [ "$DRY_RUN" = "true" ]; then
          ARGS="--dry-run"
        fi
        cargo upgrade $ARGS --manifest-path Cargo.toml --package anyhow@{{ source "anyhow" }}
        cargo update $ARGS --manifest-path Cargo.toml anyhow@{{ source "anyhow" }} --precise {{ source "anyhow" }}
      changedif:
        kind: file/checksum
        spec:
          files:
            - "Cargo.toml"
            - "Cargo.lock"
    disablesourceinput: true
`, `name: 'deps(cargo): bump dependencies "rand" for "test-crate" crate'
sources:
  rand:
    name: 'Get latest "rand" crate version'
    kind: 'cargopackage'
    spec:
      package: 'rand'
      versionfilter:
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
    name: 'Test if version of "rand" {{ source "rand-current-version" }} differs from {{ source "rand" }}'
    kind: 'shell'
    sourceid: 'rand'
    spec:
      command: 'test {{ source "rand-current-version" }} != '
targets:
  rand:
    name: 'deps(cargo): bump crate dependency "rand" to {{ source "rand" }}'
    kind: 'shell'
    spec:
      command: |
        ARGS=""
        if [ "$DRY_RUN" = "true" ]; then
          ARGS="--dry-run"
        fi
        cargo upgrade $ARGS --manifest-path Cargo.toml --package rand@{{ source "rand" }}
        cargo update $ARGS --manifest-path Cargo.toml rand@{{ source "rand" }} --precise {{ source "rand" }}
      changedif:
        kind: file/checksum
        spec:
          files:
            - "Cargo.toml"
            - "Cargo.lock"
    disablesourceinput: true
`, `name: 'deps(cargo): bump dev-dependencies "futures" for "test-crate" crate'
sources:
  futures:
    name: 'Get latest "futures" crate version'
    kind: 'cargopackage'
    spec:
      package: 'futures'
      versionfilter:
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
    name: 'Test if version of "futures" {{ source "futures-current-version" }} differs from {{ source "futures" }}'
    kind: 'shell'
    sourceid: 'futures'
    spec:
      command: 'test {{ source "futures-current-version" }} != '
targets:
  futures:
    name: 'deps(cargo): bump crate dependency "futures" to {{ source "futures" }}'
    kind: 'shell'
    spec:
      command: |
        ARGS=""
        if [ "$DRY_RUN" = "true" ]; then
          ARGS="--dry-run"
        fi
        cargo upgrade $ARGS --manifest-path Cargo.toml --package futures@{{ source "futures" }}
        cargo update $ARGS --manifest-path Cargo.toml futures@{{ source "futures" }} --precise {{ source "futures" }}
      changedif:
        kind: file/checksum
        spec:
          files:
            - "Cargo.toml"
            - "Cargo.lock"
    disablesourceinput: true
`},
		}, {
			name:                  "Scenario 2 -- Cargo Upgrade unavailable",
			rootDir:               "testdata",
			cargoUpgradeAvailable: false,
			expectedPipelines: []string{`name: 'deps(cargo): bump dependencies "anyhow" for "test-crate" crate'
sources:
  anyhow:
    name: 'Get latest "anyhow" crate version'
    kind: 'cargopackage'
    spec:
      package: 'anyhow'
      versionfilter:
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
    name: 'Test if version of "anyhow" {{ source "anyhow-current-version" }} differs from {{ source "anyhow" }}'
    kind: 'shell'
    sourceid: 'anyhow'
    spec:
      command: 'test {{ source "anyhow-current-version" }} != '
targets:
  anyhow:
    name: 'deps(cargo): bump crate dependency "anyhow" to {{ source "anyhow" }}'
    kind: 'toml'
    spec:
      file: 'Cargo.toml'
      key: 'dependencies.anyhow'
    sourceid: 'anyhow'
`, `name: 'deps(cargo): bump dependencies "rand" for "test-crate" crate'
sources:
  rand:
    name: 'Get latest "rand" crate version'
    kind: 'cargopackage'
    spec:
      package: 'rand'
      versionfilter:
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
    name: 'Test if version of "rand" {{ source "rand-current-version" }} differs from {{ source "rand" }}'
    kind: 'shell'
    sourceid: 'rand'
    spec:
      command: 'test {{ source "rand-current-version" }} != '
targets:
  rand:
    name: 'deps(cargo): bump crate dependency "rand" to {{ source "rand" }}'
    kind: 'toml'
    spec:
      file: 'Cargo.toml'
      key: 'dependencies.rand.version'
    sourceid: 'rand'
`, `name: 'deps(cargo): bump dev-dependencies "futures" for "test-crate" crate'
sources:
  futures:
    name: 'Get latest "futures" crate version'
    kind: 'cargopackage'
    spec:
      package: 'futures'
      versionfilter:
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
    name: 'Test if version of "futures" {{ source "futures-current-version" }} differs from {{ source "futures" }}'
    kind: 'shell'
    sourceid: 'futures'
    spec:
      command: 'test {{ source "futures-current-version" }} != '
targets:
  futures:
    name: 'deps(cargo): bump crate dependency "futures" to {{ source "futures" }}'
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
			c, err := New(
				Spec{}, tt.rootDir, "", "")

			cargoUpgradeCheckError := ""
			if !tt.cargoUpgradeAvailable {
				cargoUpgradeCheckError = "not found in path"
			}
			e := TestCargoUpgradeCheckerExecutor{
				errorMessage: cargoUpgradeCheckError,
			}
			c.cargoUpgradeCheckerExecutor = e

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
