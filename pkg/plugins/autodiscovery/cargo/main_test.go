package cargo

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
		name                  string
		rootDir               string
		expectedScmID         string
		expectedActionID      string
		cargoAvailable        bool
		cargoUpgradeAvailable bool
		expectedPipelines     []string
	}{
		{
			name:                  "Scenario 1 -- simple_crate -- Cargo and Cargo Upgrade unavailable",
			rootDir:               "testdata/simple_crate",
			cargoAvailable:        false,
			cargoUpgradeAvailable: false,
			expectedPipelines: []string{`name: 'deps(cargo): bump dependency "anyhow" for "test-crate" crate'
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
`, `name: 'deps(cargo): bump dependency "rand" for "test-crate" crate'
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
`, `name: 'deps(cargo): bump dev dependency "futures" for "test-crate" crate'
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
    name: 'deps(cargo): bump crate dev dependency "futures" to {{ source "futures" }}'
    kind: 'toml'
    spec:
      file: 'Cargo.toml'
      key: 'dev-dependencies.futures.version'
    sourceid: 'futures'
`},
		},
		{
			name:                  "Scenario 2 -- simple_crate -- Cargo available and Cargo Upgrade unavailable",
			rootDir:               "testdata/simple_crate",
			cargoAvailable:        true,
			cargoUpgradeAvailable: false,
			expectedPipelines: []string{`name: 'deps(cargo): bump dependency "anyhow" for "test-crate" crate'
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
`, `name: 'deps(cargo): bump dependency "rand" for "test-crate" crate'
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
`, `name: 'deps(cargo): bump dev dependency "futures" for "test-crate" crate'
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
    name: 'deps(cargo): bump crate dev dependency "futures" to {{ source "futures" }}'
    kind: 'toml'
    spec:
      file: 'Cargo.toml'
      key: 'dev-dependencies.futures.version'
    sourceid: 'futures'
`},
		},
		{
			name:                  "Scenario 3 -- simple_crate -- Cargo and Cargo Upgrade available",
			rootDir:               "testdata/simple_crate",
			cargoAvailable:        true,
			cargoUpgradeAvailable: true,
			expectedPipelines: []string{`name: 'deps(cargo): bump dependency "anyhow" for "test-crate" crate'
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
      changedif:
        kind: file/checksum
        spec:
          files:
            - "Cargo.toml"
    disablesourceinput: true
`, `name: 'deps(cargo): bump dependency "rand" for "test-crate" crate'
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
      changedif:
        kind: file/checksum
        spec:
          files:
            - "Cargo.toml"
    disablesourceinput: true
`, `name: 'deps(cargo): bump dev dependency "futures" for "test-crate" crate'
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
    name: 'deps(cargo): bump crate dev dependency "futures" to {{ source "futures" }}'
    kind: 'shell'
    spec:
      command: |
        ARGS=""
        if [ "$DRY_RUN" = "true" ]; then
          ARGS="--dry-run"
        fi
        cargo upgrade $ARGS --manifest-path Cargo.toml --package futures@{{ source "futures" }}
      changedif:
        kind: file/checksum
        spec:
          files:
            - "Cargo.toml"
    disablesourceinput: true
`},
		},
		{
			name:                  "Scenario 4 -- simple_crate_lock -- Cargo and Cargo Upgrade unavailable and lockfile",
			rootDir:               "testdata/simple_crate_lock",
			cargoAvailable:        false,
			cargoUpgradeAvailable: false,
			expectedPipelines:     []string{},
		},
		{
			name:                  "Scenario 5 -- simple_crate_lock -- Cargo available and Cargo Upgrade unavailable and lockfile",
			rootDir:               "testdata/simple_crate_lock",
			cargoAvailable:        true,
			cargoUpgradeAvailable: false,
			expectedPipelines: []string{`name: 'deps(cargo): bump dependency "anyhow" for "test-crate" crate'
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
  lockfile:
    name: 'deps(cargo): update Cargo.lock following bump crate dependency "anyhow" to {{ source "anyhow" }}'
    kind: 'shell'
    dependson:
      - target#anyhow
    spec:
      command: |
        ARGS=""
        if [ "$DRY_RUN" = "true" ]; then
          ARGS="--dry-run"
        fi
        cargo update $ARGS --manifest-path Cargo.toml anyhow@{{ source "anyhow" }} --precise {{ source "anyhow" }}
      changedif:
        kind: file/checksum
        spec:
          files:
            - "Cargo.lock"
    disablesourceinput: true
`, `name: 'deps(cargo): bump dependency "rand" for "test-crate" crate'
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
  lockfile:
    name: 'deps(cargo): update Cargo.lock following bump crate dependency "rand" to {{ source "rand" }}'
    kind: 'shell'
    dependson:
      - target#rand
    spec:
      command: |
        ARGS=""
        if [ "$DRY_RUN" = "true" ]; then
          ARGS="--dry-run"
        fi
        cargo update $ARGS --manifest-path Cargo.toml rand@{{ source "rand" }} --precise {{ source "rand" }}
      changedif:
        kind: file/checksum
        spec:
          files:
            - "Cargo.lock"
    disablesourceinput: true
`, `name: 'deps(cargo): bump dev dependency "futures" for "test-crate" crate'
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
    name: 'deps(cargo): bump crate dev dependency "futures" to {{ source "futures" }}'
    kind: 'toml'
    spec:
      file: 'Cargo.toml'
      key: 'dev-dependencies.futures.version'
    sourceid: 'futures'
  lockfile:
    name: 'deps(cargo): update Cargo.lock following bump crate dev dependency "futures" to {{ source "futures" }}'
    kind: 'shell'
    dependson:
      - target#futures
    spec:
      command: |
        ARGS=""
        if [ "$DRY_RUN" = "true" ]; then
          ARGS="--dry-run"
        fi
        cargo update $ARGS --manifest-path Cargo.toml futures@{{ source "futures" }} --precise {{ source "futures" }}
      changedif:
        kind: file/checksum
        spec:
          files:
            - "Cargo.lock"
    disablesourceinput: true
`},
		},
		{
			name:                  "Scenario 6 -- simple_crate_lock -- Cargo and Cargo Upgrade available and lockfile",
			rootDir:               "testdata/simple_crate_lock",
			cargoAvailable:        true,
			cargoUpgradeAvailable: true,
			expectedPipelines: []string{`name: 'deps(cargo): bump dependency "anyhow" for "test-crate" crate'
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
`, `name: 'deps(cargo): bump dependency "rand" for "test-crate" crate'
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
`, `name: 'deps(cargo): bump dev dependency "futures" for "test-crate" crate'
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
    name: 'deps(cargo): bump crate dev dependency "futures" to {{ source "futures" }}'
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
		},
		{
			name:                  "Scenario 7 -- workspace -- Cargo and Cargo Upgrade unavailable, workspace, no lockfile",
			rootDir:               "testdata/workspace",
			cargoAvailable:        false,
			cargoUpgradeAvailable: false,
			expectedPipelines: []string{`name: 'deps(cargo): bump workspace dependency "anyhow"'
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
      Key: 'workspace.dependencies.anyhow'
conditions:
  anyhow:
    name: 'Test if version of "anyhow" {{ source "anyhow-current-version" }} differs from {{ source "anyhow" }}'
    kind: 'shell'
    sourceid: 'anyhow'
    spec:
      command: 'test {{ source "anyhow-current-version" }} != '
targets:
  anyhow:
    name: 'deps(cargo): bump workspace dependency "anyhow" to {{ source "anyhow" }}'
    kind: 'toml'
    spec:
      file: 'Cargo.toml'
      key: 'workspace.dependencies.anyhow'
    sourceid: 'anyhow'
`,
				`name: 'deps(cargo): bump dependency "rand" for "simple_crate" crate'
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
      file: 'crates/simple_crate/Cargo.toml'
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
      file: 'crates/simple_crate/Cargo.toml'
      key: 'dependencies.rand.version'
    sourceid: 'rand'
`, `name: 'deps(cargo): bump dev dependency "futures" for "simple_crate" crate'
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
      file: 'crates/simple_crate/Cargo.toml'
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
    name: 'deps(cargo): bump crate dev dependency "futures" to {{ source "futures" }}'
    kind: 'toml'
    spec:
      file: 'crates/simple_crate/Cargo.toml'
      key: 'dev-dependencies.futures.version'
    sourceid: 'futures'
`},
		}, {

			name:                  "Scenario 8 -- workspace -- Cargo available and Cargo Upgrade unavailable, workspace, no lockfile",
			rootDir:               "testdata/workspace",
			cargoAvailable:        true,
			cargoUpgradeAvailable: false,
			expectedPipelines: []string{`name: 'deps(cargo): bump workspace dependency "anyhow"'
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
      Key: 'workspace.dependencies.anyhow'
conditions:
  anyhow:
    name: 'Test if version of "anyhow" {{ source "anyhow-current-version" }} differs from {{ source "anyhow" }}'
    kind: 'shell'
    sourceid: 'anyhow'
    spec:
      command: 'test {{ source "anyhow-current-version" }} != '
targets:
  anyhow:
    name: 'deps(cargo): bump workspace dependency "anyhow" to {{ source "anyhow" }}'
    kind: 'toml'
    spec:
      file: 'Cargo.toml'
      key: 'workspace.dependencies.anyhow'
    sourceid: 'anyhow'
`,
				`name: 'deps(cargo): bump dependency "rand" for "simple_crate" crate'
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
      file: 'crates/simple_crate/Cargo.toml'
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
      file: 'crates/simple_crate/Cargo.toml'
      key: 'dependencies.rand.version'
    sourceid: 'rand'
`, `name: 'deps(cargo): bump dev dependency "futures" for "simple_crate" crate'
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
      file: 'crates/simple_crate/Cargo.toml'
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
    name: 'deps(cargo): bump crate dev dependency "futures" to {{ source "futures" }}'
    kind: 'toml'
    spec:
      file: 'crates/simple_crate/Cargo.toml'
      key: 'dev-dependencies.futures.version'
    sourceid: 'futures'
`},
		}, {
			name:                  "Scenario 9 -- workspace -- Cargo and Cargo Upgrade available, workspace, no lockfile",
			rootDir:               "testdata/workspace",
			cargoAvailable:        true,
			cargoUpgradeAvailable: true,
			expectedPipelines: []string{`name: 'deps(cargo): bump workspace dependency "anyhow"'
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
      Key: 'workspace.dependencies.anyhow'
conditions:
  anyhow:
    name: 'Test if version of "anyhow" {{ source "anyhow-current-version" }} differs from {{ source "anyhow" }}'
    kind: 'shell'
    sourceid: 'anyhow'
    spec:
      command: 'test {{ source "anyhow-current-version" }} != '
targets:
  anyhow:
    name: 'deps(cargo): bump workspace dependency "anyhow" to {{ source "anyhow" }}'
    kind: 'shell'
    spec:
      command: |
        ARGS=""
        if [ "$DRY_RUN" = "true" ]; then
          ARGS="--dry-run"
        fi
        cargo upgrade $ARGS --manifest-path Cargo.toml --package anyhow@{{ source "anyhow" }}
      changedif:
        kind: file/checksum
        spec:
          files:
            - "Cargo.toml"
    disablesourceinput: true
`, `name: 'deps(cargo): bump dependency "rand" for "simple_crate" crate'
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
      file: 'crates/simple_crate/Cargo.toml'
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
        cargo upgrade $ARGS --manifest-path crates/simple_crate/Cargo.toml --package rand@{{ source "rand" }}
      changedif:
        kind: file/checksum
        spec:
          files:
            - "crates/simple_crate/Cargo.toml"
    disablesourceinput: true
`, `name: 'deps(cargo): bump dev dependency "futures" for "simple_crate" crate'
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
      file: 'crates/simple_crate/Cargo.toml'
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
    name: 'deps(cargo): bump crate dev dependency "futures" to {{ source "futures" }}'
    kind: 'shell'
    spec:
      command: |
        ARGS=""
        if [ "$DRY_RUN" = "true" ]; then
          ARGS="--dry-run"
        fi
        cargo upgrade $ARGS --manifest-path crates/simple_crate/Cargo.toml --package futures@{{ source "futures" }}
      changedif:
        kind: file/checksum
        spec:
          files:
            - "crates/simple_crate/Cargo.toml"
    disablesourceinput: true
`},
		},
		{
			name:                  "Scenario 10 -- workspace -- Cargo and Cargo Upgrade unavailable, workspace, lockfile",
			rootDir:               "testdata/workspace_lock",
			cargoAvailable:        false,
			cargoUpgradeAvailable: false,
			expectedPipelines:     []string{},
		}, {

			name:                  "Scenario 11 -- workspace -- Cargo available and Cargo Upgrade unavailable, workspace, lockfile",
			rootDir:               "testdata/workspace_lock",
			cargoAvailable:        true,
			cargoUpgradeAvailable: false,
			expectedPipelines: []string{`name: 'deps(cargo): bump workspace dependency "anyhow"'
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
      Key: 'workspace.dependencies.anyhow'
conditions:
  anyhow:
    name: 'Test if version of "anyhow" {{ source "anyhow-current-version" }} differs from {{ source "anyhow" }}'
    kind: 'shell'
    sourceid: 'anyhow'
    spec:
      command: 'test {{ source "anyhow-current-version" }} != '
targets:
  anyhow:
    name: 'deps(cargo): bump workspace dependency "anyhow" to {{ source "anyhow" }}'
    kind: 'toml'
    spec:
      file: 'Cargo.toml'
      key: 'workspace.dependencies.anyhow'
    sourceid: 'anyhow'
  lockfile:
    name: 'deps(cargo): update Cargo.lock following bump workspace dependency "anyhow" to {{ source "anyhow" }}'
    kind: 'shell'
    dependson:
      - target#anyhow
    spec:
      command: |
        ARGS=""
        if [ "$DRY_RUN" = "true" ]; then
          ARGS="--dry-run"
        fi
        cargo update $ARGS --manifest-path Cargo.toml anyhow@{{ source "anyhow" }} --precise {{ source "anyhow" }}
      changedif:
        kind: file/checksum
        spec:
          files:
            - "Cargo.lock"
    disablesourceinput: true
`,
				`name: 'deps(cargo): bump dependency "rand" for "simple_crate" crate'
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
      file: 'crates/simple_crate/Cargo.toml'
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
      file: 'crates/simple_crate/Cargo.toml'
      key: 'dependencies.rand.version'
    sourceid: 'rand'
  lockfile:
    name: 'deps(cargo): update Cargo.lock following bump crate dependency "rand" to {{ source "rand" }}'
    kind: 'shell'
    dependson:
      - target#rand
    spec:
      command: |
        ARGS=""
        if [ "$DRY_RUN" = "true" ]; then
          ARGS="--dry-run"
        fi
        cargo update $ARGS --manifest-path crates/simple_crate/Cargo.toml rand@{{ source "rand" }} --precise {{ source "rand" }}
      changedif:
        kind: file/checksum
        spec:
          files:
            - "Cargo.lock"
    disablesourceinput: true
`, `name: 'deps(cargo): bump dev dependency "futures" for "simple_crate" crate'
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
      file: 'crates/simple_crate/Cargo.toml'
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
    name: 'deps(cargo): bump crate dev dependency "futures" to {{ source "futures" }}'
    kind: 'toml'
    spec:
      file: 'crates/simple_crate/Cargo.toml'
      key: 'dev-dependencies.futures.version'
    sourceid: 'futures'
  lockfile:
    name: 'deps(cargo): update Cargo.lock following bump crate dev dependency "futures" to {{ source "futures" }}'
    kind: 'shell'
    dependson:
      - target#futures
    spec:
      command: |
        ARGS=""
        if [ "$DRY_RUN" = "true" ]; then
          ARGS="--dry-run"
        fi
        cargo update $ARGS --manifest-path crates/simple_crate/Cargo.toml futures@{{ source "futures" }} --precise {{ source "futures" }}
      changedif:
        kind: file/checksum
        spec:
          files:
            - "Cargo.lock"
    disablesourceinput: true
`},
		}, {
			name:                  "Scenario 12 -- workspace -- Cargo and Cargo Upgrade available, workspace, lockfile",
			rootDir:               "testdata/workspace_lock",
			expectedScmID:         "git",
			expectedActionID:      "github",
			cargoAvailable:        true,
			cargoUpgradeAvailable: true,
			expectedPipelines: []string{`name: 'deps(cargo): bump workspace dependency "anyhow"'
actions:
  github:
    title: 'deps(cargo): bump workspace dependency "anyhow" to {{ source "anyhow" }}'

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
    scmid: 'git'
    spec:
      file: 'Cargo.toml'
      Key: 'workspace.dependencies.anyhow'
conditions:
  anyhow:
    name: 'Test if version of "anyhow" {{ source "anyhow-current-version" }} differs from {{ source "anyhow" }}'
    kind: 'shell'
    sourceid: 'anyhow'
    spec:
      command: 'test {{ source "anyhow-current-version" }} != '
targets:
  anyhow:
    name: 'deps(cargo): bump workspace dependency "anyhow" to {{ source "anyhow" }}'
    scmid: 'git'
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
`, `name: 'deps(cargo): bump dependency "rand" for "simple_crate" crate'
actions:
  github:
    title: 'deps(cargo): bump crate dependency "rand" to {{ source "rand" }}'

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
    scmid: 'git'
    spec:
      file: 'crates/simple_crate/Cargo.toml'
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
    scmid: 'git'
    kind: 'shell'
    spec:
      command: |
        ARGS=""
        if [ "$DRY_RUN" = "true" ]; then
          ARGS="--dry-run"
        fi
        cargo upgrade $ARGS --manifest-path crates/simple_crate/Cargo.toml --package rand@{{ source "rand" }}
        cargo update $ARGS --manifest-path crates/simple_crate/Cargo.toml rand@{{ source "rand" }} --precise {{ source "rand" }}
      changedif:
        kind: file/checksum
        spec:
          files:
            - "crates/simple_crate/Cargo.toml"
            - "Cargo.lock"
    disablesourceinput: true
`, `name: 'deps(cargo): bump dev dependency "futures" for "simple_crate" crate'
actions:
  github:
    title: 'deps(cargo): bump crate dev dependency "futures" to {{ source "futures" }}'

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
    scmid: 'git'
    spec:
      file: 'crates/simple_crate/Cargo.toml'
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
    name: 'deps(cargo): bump crate dev dependency "futures" to {{ source "futures" }}'
    scmid: 'git'
    kind: 'shell'
    spec:
      command: |
        ARGS=""
        if [ "$DRY_RUN" = "true" ]; then
          ARGS="--dry-run"
        fi
        cargo upgrade $ARGS --manifest-path crates/simple_crate/Cargo.toml --package futures@{{ source "futures" }}
        cargo update $ARGS --manifest-path crates/simple_crate/Cargo.toml futures@{{ source "futures" }} --precise {{ source "futures" }}
      changedif:
        kind: file/checksum
        spec:
          files:
            - "crates/simple_crate/Cargo.toml"
            - "Cargo.lock"
    disablesourceinput: true
`},
		},
	}

	for _, tt := range testdata {

		t.Run(tt.name, func(t *testing.T) {
			c, err := New(
				Spec{}, tt.rootDir, tt.expectedScmID, tt.expectedActionID)

			c.cargoAvailable = tt.cargoAvailable
			c.cargoUpgradeAvailable = tt.cargoUpgradeAvailable

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
