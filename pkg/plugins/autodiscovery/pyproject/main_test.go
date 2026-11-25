package pyproject

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func TestDiscoverManifests(t *testing.T) {
	testdata := []struct {
		name              string
		rootDir           string
		scmID             string
		actionID          string
		spec              Spec
		uvAvailable       bool
		expectedPipelines []string
	}{
		{
			name:        "Scenario 1 -- simple project, uv available",
			rootDir:     "testdata/simple_project",
			uvAvailable: true,
			expectedPipelines: []string{
				`name: 'deps(pypi): bump "flask" for "simple-project" project'
sources:
  flask:
    name: 'Get latest "flask" package version'
    kind: 'pypi'
    spec:
      name: 'flask'
      age:
        minimum: '7d'
      versionfilter:
        kind: 'pep440'
        pattern: '>=3.0'
targets:
  flask:
    name: 'deps(pypi): bump "flask" to {{ source "flask" }}'
    kind: 'shell'
    spec:
      command: 'uv lock --upgrade-package flask=={{ source "flask" }}'
      changedif:
        kind: file/checksum
        spec:
          files:
            - "uv.lock"
      environments:
        - name: PATH
      workdir: '.'
    disablesourceinput: true
`,
				`name: 'deps(pypi): bump "requests" for "simple-project" project'
sources:
  requests:
    name: 'Get latest "requests" package version'
    kind: 'pypi'
    spec:
      name: 'requests'
      age:
        minimum: '7d'
      versionfilter:
        kind: 'pep440'
        pattern: '>=2.28'
targets:
  requests:
    name: 'deps(pypi): bump "requests" to {{ source "requests" }}'
    kind: 'shell'
    spec:
      command: 'uv lock --upgrade-package requests=={{ source "requests" }}'
      changedif:
        kind: file/checksum
        spec:
          files:
            - "uv.lock"
      environments:
        - name: PATH
      workdir: '.'
    disablesourceinput: true
`,
			},
		},
		{
			name:              "Scenario 2 -- simple project, uv NOT available",
			rootDir:           "testdata/simple_project",
			uvAvailable:       false,
			expectedPipelines: []string{},
		},
		{
			name:        "Scenario 3 -- no lock file (source-only manifests)",
			rootDir:     "testdata/no_lockfile",
			uvAvailable: true,
			expectedPipelines: []string{
				`name: 'deps(pypi): bump "requests" for "no-lock-project" project'
sources:
  requests:
    name: 'Get latest "requests" package version'
    kind: 'pypi'
    spec:
      name: 'requests'
      age:
        minimum: '7d'
      versionfilter:
        kind: 'pep440'
        pattern: '>=2.28'
`,
			},
		},
		{
			name:        "Scenario 4 -- optional deps",
			rootDir:     "testdata/optional_deps",
			uvAvailable: true,
			expectedPipelines: []string{
				`name: 'deps(pypi): bump "requests" for "optional-project" project'
sources:
  requests:
    name: 'Get latest "requests" package version'
    kind: 'pypi'
    spec:
      name: 'requests'
      age:
        minimum: '7d'
      versionfilter:
        kind: 'pep440'
        pattern: '>=2.28'
targets:
  requests:
    name: 'deps(pypi): bump "requests" to {{ source "requests" }}'
    kind: 'shell'
    spec:
      command: 'uv lock --upgrade-package requests=={{ source "requests" }}'
      changedif:
        kind: file/checksum
        spec:
          files:
            - "uv.lock"
      environments:
        - name: PATH
      workdir: '.'
    disablesourceinput: true
`,
				`name: 'deps(pypi): bump "pytest" [dev] for "optional-project" project'
sources:
  pytest:
    name: 'Get latest "pytest" package version'
    kind: 'pypi'
    spec:
      name: 'pytest'
      age:
        minimum: '7d'
      versionfilter:
        kind: 'pep440'
        pattern: '>=8.0'
targets:
  pytest:
    name: 'deps(pypi): bump "pytest" to {{ source "pytest" }}'
    kind: 'shell'
    spec:
      command: 'uv lock --upgrade-package pytest=={{ source "pytest" }}'
      changedif:
        kind: file/checksum
        spec:
          files:
            - "uv.lock"
      environments:
        - name: PATH
      workdir: '.'
    disablesourceinput: true
`,
				`name: 'deps(pypi): bump "ruff" [dev] for "optional-project" project'
sources:
  ruff:
    name: 'Get latest "ruff" package version'
    kind: 'pypi'
    spec:
      name: 'ruff'
      age:
        minimum: '7d'
      versionfilter:
        kind: 'pep440'
        pattern: '>=0.4'
targets:
  ruff:
    name: 'deps(pypi): bump "ruff" to {{ source "ruff" }}'
    kind: 'shell'
    spec:
      command: 'uv lock --upgrade-package ruff=={{ source "ruff" }}'
      changedif:
        kind: file/checksum
        spec:
          files:
            - "uv.lock"
      environments:
        - name: PATH
      workdir: '.'
    disablesourceinput: true
`,
			},
		},
		{
			name:        "Scenario 5 -- with scmID and actionID",
			rootDir:     "testdata/simple_project",
			scmID:       "git",
			actionID:    "github",
			uvAvailable: true,
			expectedPipelines: []string{
				`name: 'deps(pypi): bump "flask" for "simple-project" project'
actions:
  github:
    title: 'deps(pypi): bump "flask" to {{ source "flask" }}'

sources:
  flask:
    name: 'Get latest "flask" package version'
    kind: 'pypi'
    spec:
      name: 'flask'
      age:
        minimum: '7d'
      versionfilter:
        kind: 'pep440'
        pattern: '>=3.0'
targets:
  flask:
    name: 'deps(pypi): bump "flask" to {{ source "flask" }}'
    scmid: 'git'
    kind: 'shell'
    spec:
      command: 'uv lock --upgrade-package flask=={{ source "flask" }}'
      changedif:
        kind: file/checksum
        spec:
          files:
            - "uv.lock"
      environments:
        - name: PATH
      workdir: '.'
    disablesourceinput: true
`,
				`name: 'deps(pypi): bump "requests" for "simple-project" project'
actions:
  github:
    title: 'deps(pypi): bump "requests" to {{ source "requests" }}'

sources:
  requests:
    name: 'Get latest "requests" package version'
    kind: 'pypi'
    spec:
      name: 'requests'
      age:
        minimum: '7d'
      versionfilter:
        kind: 'pep440'
        pattern: '>=2.28'
targets:
  requests:
    name: 'deps(pypi): bump "requests" to {{ source "requests" }}'
    scmid: 'git'
    kind: 'shell'
    spec:
      command: 'uv lock --upgrade-package requests=={{ source "requests" }}'
      changedif:
        kind: file/checksum
        spec:
          files:
            - "uv.lock"
      environments:
        - name: PATH
      workdir: '.'
    disablesourceinput: true
`,
			},
		},
		{
			name:        "Scenario 6 -- ignore rule excludes flask",
			rootDir:     "testdata/simple_project",
			uvAvailable: true,
			spec: Spec{
				Ignore: MatchingRules{
					{Packages: map[string]string{"flask": ""}},
				},
			},
			expectedPipelines: []string{
				`name: 'deps(pypi): bump "requests" for "simple-project" project'
sources:
  requests:
    name: 'Get latest "requests" package version'
    kind: 'pypi'
    spec:
      name: 'requests'
      age:
        minimum: '7d'
      versionfilter:
        kind: 'pep440'
        pattern: '>=2.28'
targets:
  requests:
    name: 'deps(pypi): bump "requests" to {{ source "requests" }}'
    kind: 'shell'
    spec:
      command: 'uv lock --upgrade-package requests=={{ source "requests" }}'
      changedif:
        kind: file/checksum
        spec:
          files:
            - "uv.lock"
      environments:
        - name: PATH
      workdir: '.'
    disablesourceinput: true
`,
			},
		},
		{
			name:        "Scenario 7 -- only rule restricts to requests",
			rootDir:     "testdata/simple_project",
			uvAvailable: true,
			spec: Spec{
				Only: MatchingRules{
					{Packages: map[string]string{"requests": ""}},
				},
			},
			expectedPipelines: []string{
				`name: 'deps(pypi): bump "requests" for "simple-project" project'
sources:
  requests:
    name: 'Get latest "requests" package version'
    kind: 'pypi'
    spec:
      name: 'requests'
      age:
        minimum: '7d'
      versionfilter:
        kind: 'pep440'
        pattern: '>=2.28'
targets:
  requests:
    name: 'deps(pypi): bump "requests" to {{ source "requests" }}'
    kind: 'shell'
    spec:
      command: 'uv lock --upgrade-package requests=={{ source "requests" }}'
      changedif:
        kind: file/checksum
        spec:
          files:
            - "uv.lock"
      environments:
        - name: PATH
      workdir: '.'
    disablesourceinput: true
`,
			},
		},
		{
			name:        "Scenario 8 -- env markers are stripped from dependency strings",
			rootDir:     "testdata/markers",
			uvAvailable: true,
			expectedPipelines: []string{
				`name: 'deps(pypi): bump "pywin32" for "markers-project" project'
sources:
  pywin32:
    name: 'Get latest "pywin32" package version'
    kind: 'pypi'
    spec:
      name: 'pywin32'
      age:
        minimum: '7d'
      versionfilter:
        kind: 'pep440'
        pattern: '>=300'
targets:
  pywin32:
    name: 'deps(pypi): bump "pywin32" to {{ source "pywin32" }}'
    kind: 'shell'
    spec:
      command: 'uv lock --upgrade-package pywin32=={{ source "pywin32" }}'
      changedif:
        kind: file/checksum
        spec:
          files:
            - "uv.lock"
      environments:
        - name: PATH
      workdir: '.'
    disablesourceinput: true
`,
				`name: 'deps(pypi): bump "requests" for "markers-project" project'
sources:
  requests:
    name: 'Get latest "requests" package version'
    kind: 'pypi'
    spec:
      name: 'requests'
      age:
        minimum: '7d'
      versionfilter:
        kind: 'pep440'
        pattern: '>=2.28'
targets:
  requests:
    name: 'deps(pypi): bump "requests" to {{ source "requests" }}'
    kind: 'shell'
    spec:
      command: 'uv lock --upgrade-package requests=={{ source "requests" }}'
      changedif:
        kind: file/checksum
        spec:
          files:
            - "uv.lock"
      environments:
        - name: PATH
      workdir: '.'
    disablesourceinput: true
`,
			},
		},
		{
			name:        "Scenario 9 -- no version constraint uses wildcard pattern",
			rootDir:     "testdata/no_version",
			uvAvailable: true,
			expectedPipelines: []string{
				`name: 'deps(pypi): bump "numpy" for "noversion-project" project'
sources:
  numpy:
    name: 'Get latest "numpy" package version'
    kind: 'pypi'
    spec:
      name: 'numpy'
      age:
        minimum: '7d'
      versionfilter:
        kind: 'pep440'
        pattern: '*'
targets:
  numpy:
    name: 'deps(pypi): bump "numpy" to {{ source "numpy" }}'
    kind: 'shell'
    spec:
      command: 'uv lock --upgrade-package numpy=={{ source "numpy" }}'
      changedif:
        kind: file/checksum
        spec:
          files:
            - "uv.lock"
      environments:
        - name: PATH
      workdir: '.'
    disablesourceinput: true
`,
			},
		},
		{
			name:        "Scenario 10 -- custom IndexURL appears in source spec",
			rootDir:     "testdata/simple_project",
			uvAvailable: true,
			spec:        Spec{IndexURL: "https://private.pypi.org/"},
			expectedPipelines: []string{
				`name: 'deps(pypi): bump "flask" for "simple-project" project'
sources:
  flask:
    name: 'Get latest "flask" package version'
    kind: 'pypi'
    spec:
      name: 'flask'
      url: 'https://private.pypi.org/'
      versionfilter:
        kind: 'pep440'
        pattern: '>=3.0'
targets:
  flask:
    name: 'deps(pypi): bump "flask" to {{ source "flask" }}'
    kind: 'shell'
    spec:
      command: 'uv lock --upgrade-package flask=={{ source "flask" }}'
      changedif:
        kind: file/checksum
        spec:
          files:
            - "uv.lock"
      environments:
        - name: PATH
      workdir: '.'
    disablesourceinput: true
`,
				`name: 'deps(pypi): bump "requests" for "simple-project" project'
sources:
  requests:
    name: 'Get latest "requests" package version'
    kind: 'pypi'
    spec:
      name: 'requests'
      url: 'https://private.pypi.org/'
      versionfilter:
        kind: 'pep440'
        pattern: '>=2.28'
targets:
  requests:
    name: 'deps(pypi): bump "requests" to {{ source "requests" }}'
    kind: 'shell'
    spec:
      command: 'uv lock --upgrade-package requests=={{ source "requests" }}'
      changedif:
        kind: file/checksum
        spec:
          files:
            - "uv.lock"
      environments:
        - name: PATH
      workdir: '.'
    disablesourceinput: true
`,
			},
		},
		{
			name:        "Scenario 11 -- custom VersionFilter overrides per-dep constraint",
			rootDir:     "testdata/simple_project",
			uvAvailable: true,
			spec:        Spec{VersionFilter: version.Filter{Kind: "semver", Pattern: "minor"}},
			expectedPipelines: []string{
				`name: 'deps(pypi): bump "flask" for "simple-project" project'
sources:
  flask:
    name: 'Get latest "flask" package version'
    kind: 'pypi'
    spec:
      name: 'flask'
      age:
        minimum: '7d'
      versionfilter:
        kind: 'semver'
        pattern: '>=3.0.0, 3.x'
targets:
  flask:
    name: 'deps(pypi): bump "flask" to {{ source "flask" }}'
    kind: 'shell'
    spec:
      command: 'uv lock --upgrade-package flask=={{ source "flask" }}'
      changedif:
        kind: file/checksum
        spec:
          files:
            - "uv.lock"
      environments:
        - name: PATH
      workdir: '.'
    disablesourceinput: true
`,
				`name: 'deps(pypi): bump "requests" for "simple-project" project'
sources:
  requests:
    name: 'Get latest "requests" package version'
    kind: 'pypi'
    spec:
      name: 'requests'
      age:
        minimum: '7d'
      versionfilter:
        kind: 'semver'
        pattern: '>=2.28.0, 2.x'
targets:
  requests:
    name: 'deps(pypi): bump "requests" to {{ source "requests" }}'
    kind: 'shell'
    spec:
      command: 'uv lock --upgrade-package requests=={{ source "requests" }}'
      changedif:
        kind: file/checksum
        spec:
          files:
            - "uv.lock"
      environments:
        - name: PATH
      workdir: '.'
    disablesourceinput: true
`,
			},
		},
		{
			name:        "Scenario 12 -- nested subdirectory project tracks uv.lock relative to its workdir",
			rootDir:     "testdata/subdir_project",
			uvAvailable: true,
			expectedPipelines: []string{
				`name: 'deps(pypi): bump "requests" for "nested-project" project'
sources:
  requests:
    name: 'Get latest "requests" package version'
    kind: 'pypi'
    spec:
      name: 'requests'
      age:
        minimum: '7d'
      versionfilter:
        kind: 'pep440'
        pattern: '>=2.28'
targets:
  requests:
    name: 'deps(pypi): bump "requests" to {{ source "requests" }}'
    kind: 'shell'
    spec:
      command: 'uv lock --upgrade-package requests=={{ source "requests" }}'
      changedif:
        kind: file/checksum
        spec:
          files:
            - "uv.lock"
      environments:
        - name: PATH
      workdir: 'foo/bar'
    disablesourceinput: true
`,
			},
		},
		{
			name:        "Scenario 13 -- dependency groups, include-group not expanded, ignore applied",
			rootDir:     "testdata/dependency_groups",
			uvAvailable: true,
			spec: Spec{
				Ignore: MatchingRules{
					{Packages: map[string]string{"requests": ""}},
				},
			},
			expectedPipelines: []string{
				`name: 'deps(pypi): bump "rich" [dev] for "groups-project" project'
sources:
  rich:
    name: 'Get latest "rich" package version'
    kind: 'pypi'
    spec:
      name: 'rich'
      age:
        minimum: '7d'
      versionfilter:
        kind: 'pep440'
        pattern: '>=13.0'
targets:
  rich:
    name: 'deps(pypi): bump "rich" to {{ source "rich" }}'
    kind: 'shell'
    spec:
      command: 'uv lock --upgrade-package rich=={{ source "rich" }}'
      changedif:
        kind: file/checksum
        spec:
          files:
            - "uv.lock"
      environments:
        - name: PATH
      workdir: '.'
    disablesourceinput: true
`,
				`name: 'deps(pypi): bump "ruff" [group:dev] for "groups-project" project'
sources:
  ruff:
    name: 'Get latest "ruff" package version'
    kind: 'pypi'
    spec:
      name: 'ruff'
      age:
        minimum: '7d'
      versionfilter:
        kind: 'pep440'
        pattern: '>=0.4'
targets:
  ruff:
    name: 'deps(pypi): bump "ruff" to {{ source "ruff" }}'
    kind: 'shell'
    spec:
      command: 'uv lock --upgrade-package ruff=={{ source "ruff" }}'
      changedif:
        kind: file/checksum
        spec:
          files:
            - "uv.lock"
      environments:
        - name: PATH
      workdir: '.'
    disablesourceinput: true
`,
				`name: 'deps(pypi): bump "pytest" [group:test] for "groups-project" project'
sources:
  pytest:
    name: 'Get latest "pytest" package version'
    kind: 'pypi'
    spec:
      name: 'pytest'
      age:
        minimum: '7d'
      versionfilter:
        kind: 'pep440'
        pattern: '>=8.0'
targets:
  pytest:
    name: 'deps(pypi): bump "pytest" to {{ source "pytest" }}'
    kind: 'shell'
    spec:
      command: 'uv lock --upgrade-package pytest=={{ source "pytest" }}'
      changedif:
        kind: file/checksum
        spec:
          files:
            - "uv.lock"
      environments:
        - name: PATH
      workdir: '.'
    disablesourceinput: true
`,
			},
		},
		{
			name:        "Scenario 14 -- only rule on the dev dependency group, which includes the test group",
			rootDir:     "testdata/dependency_groups",
			uvAvailable: true,
			spec: Spec{
				Only: MatchingRules{
					{DependencyGroups: []string{"dev"}},
				},
			},
			expectedPipelines: []string{
				`name: 'deps(pypi): bump "ruff" [group:dev] for "groups-project" project'
sources:
  ruff:
    name: 'Get latest "ruff" package version'
    kind: 'pypi'
    spec:
      name: 'ruff'
      age:
        minimum: '7d'
      versionfilter:
        kind: 'pep440'
        pattern: '>=0.4'
targets:
  ruff:
    name: 'deps(pypi): bump "ruff" to {{ source "ruff" }}'
    kind: 'shell'
    spec:
      command: 'uv lock --upgrade-package ruff=={{ source "ruff" }}'
      changedif:
        kind: file/checksum
        spec:
          files:
            - "uv.lock"
      environments:
        - name: PATH
      workdir: '.'
    disablesourceinput: true
`,
				`name: 'deps(pypi): bump "pytest" [group:test] for "groups-project" project'
sources:
  pytest:
    name: 'Get latest "pytest" package version'
    kind: 'pypi'
    spec:
      name: 'pytest'
      age:
        minimum: '7d'
      versionfilter:
        kind: 'pep440'
        pattern: '>=8.0'
targets:
  pytest:
    name: 'deps(pypi): bump "pytest" to {{ source "pytest" }}'
    kind: 'shell'
    spec:
      command: 'uv lock --upgrade-package pytest=={{ source "pytest" }}'
      changedif:
        kind: file/checksum
        spec:
          files:
            - "uv.lock"
      environments:
        - name: PATH
      workdir: '.'
    disablesourceinput: true
`,
			},
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			p, err := New(tt.spec, tt.rootDir, tt.scmID, tt.actionID)
			require.NoError(t, err)

			// Override uvAvailable so tests are deterministic regardless of
			// whether the real uv CLI is installed in the test environment.
			p.uvAvailable = tt.uvAvailable

			manifests, err := p.DiscoverManifests()
			require.NoError(t, err)

			assert.Equal(t, len(tt.expectedPipelines), len(manifests))

			for i := range manifests {
				assert.Equal(t, tt.expectedPipelines[i], string(manifests[i]))
			}
		})
	}
}

func TestNewAge(t *testing.T) {
	testdata := []struct {
		name             string
		spec             map[string]interface{}
		expectedPipeline string
		expectedError    bool
	}{
		{
			name: "explicit age replaces the default",
			spec: map[string]interface{}{"age": map[string]interface{}{"minimum": "2w", "maximum": "1y"}},
			expectedPipeline: `name: 'deps(pypi): bump "requests" for "no-lock-project" project'
sources:
  requests:
    name: 'Get latest "requests" package version'
    kind: 'pypi'
    spec:
      name: 'requests'
      age:
        minimum: '2w'
        maximum: '1y'
      versionfilter:
        kind: 'pep440'
        pattern: '>=2.28'
`,
		},
		{
			name: "empty age disables the default",
			spec: map[string]interface{}{"age": map[string]interface{}{}},
			expectedPipeline: `name: 'deps(pypi): bump "requests" for "no-lock-project" project'
sources:
  requests:
    name: 'Get latest "requests" package version'
    kind: 'pypi'
    spec:
      name: 'requests'
      versionfilter:
        kind: 'pep440'
        pattern: '>=2.28'
`,
		},
		{
			name: "the default doesn't apply with indexurl, an explicit age does",
			spec: map[string]interface{}{"indexurl": "https://pypi.internal/simple", "age": map[string]interface{}{"minimum": "7d"}},
			expectedPipeline: `name: 'deps(pypi): bump "requests" for "no-lock-project" project'
sources:
  requests:
    name: 'Get latest "requests" package version'
    kind: 'pypi'
    spec:
      name: 'requests'
      url: 'https://pypi.internal/simple'
      age:
        minimum: '7d'
      versionfilter:
        kind: 'pep440'
        pattern: '>=2.28'
`,
		},
		{
			name:          "invalid age is rejected",
			spec:          map[string]interface{}{"age": map[string]interface{}{"minimum": "seven days"}},
			expectedError: true,
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			p, err := New(tt.spec, "testdata/no_lockfile", "", "")
			if tt.expectedError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			manifests, err := p.DiscoverManifests()
			require.NoError(t, err)
			require.Len(t, manifests, 1)
			assert.Equal(t, tt.expectedPipeline, string(manifests[0]))
		})
	}
}

func TestNewAgeNeverDowngrades(t *testing.T) {
	testdata := []struct {
		name string
		spec Spec
		// expectedPatterns maps a dependency to the pattern of its source versionfilter.
		expectedPatterns map[string]string
	}{
		{
			name: "the default age bounds the constraint at the version uv.lock resolves",
			expectedPatterns: map[string]string{
				"requests": ">=2.32.5, >=2.28",
				// Resolved at several versions, numpy is skipped: one upgrade can't target every fork
				// Resolved from another registry, still the version in use
				"corp-utils": ">=1.4.0",
			},
		},
		{
			name: "a user versionfilter is bounded at the version uv.lock resolves",
			spec: Spec{VersionFilter: version.Filter{Kind: "semver", Pattern: "minor"}},
			expectedPatterns: map[string]string{
				"requests":   ">=2.32.5, 2.x",
				"corp-utils": ">=1.4.0, 1.x",
			},
		},
		{
			name: "a user pep440 versionfilter is bounded at the version uv.lock resolves",
			spec: Spec{VersionFilter: version.Filter{Kind: "pep440", Pattern: "<3"}},
			expectedPatterns: map[string]string{
				"requests":   ">=2.32.5, <3",
				"corp-utils": ">=1.4.0, <3",
			},
		},
		{
			name: "allowdowngrade keeps the constraint as declared",
			spec: Spec{Age: &age.AutodiscoverySpec{Spec: age.Spec{Minimum: "7d"}, AllowDowngrade: new(true)}},
			expectedPatterns: map[string]string{
				"requests":   ">=2.28",
				"corp-utils": "*",
			},
		},
		{
			name: "without an age filter, the constraint is kept as declared",
			spec: Spec{Age: &age.AutodiscoverySpec{}},
			expectedPatterns: map[string]string{
				"requests":   ">=2.28",
				"corp-utils": "*",
			},
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			p, err := New(tt.spec, "testdata/locked_versions", "", "")
			require.NoError(t, err)
			p.uvAvailable = true

			manifests, err := p.DiscoverManifests()
			require.NoError(t, err)
			require.Len(t, manifests, len(tt.expectedPatterns), "manifests: %s", manifests)

			for _, manifest := range manifests {
				var found bool
				for dependency, pattern := range tt.expectedPatterns {
					if !strings.Contains(string(manifest), fmt.Sprintf("name: '%s'\n", dependency)) {
						continue
					}
					found = true
					assert.Contains(t, string(manifest), fmt.Sprintf("pattern: '%s'\n", pattern), "dependency %q", dependency)
				}
				assert.True(t, found, "unexpected manifest:\n%s", manifest)
			}
		})
	}
}

func TestNewAgeAllowDowngrade(t *testing.T) {
	// "allowdowngrade" alone keeps the default age
	p, err := New(map[string]interface{}{"age": map[string]interface{}{"allowdowngrade": true}}, "testdata/no_lockfile", "", "")
	require.NoError(t, err)
	assert.Equal(t, defaultMinimumReleaseAge, p.ageFilter.Minimum)
	assert.True(t, p.ageFilter.MayDowngrade())

	// "age: {}" disables the default, so nothing is held back and nothing is bounded
	p, err = New(map[string]interface{}{"age": map[string]interface{}{}}, "testdata/no_lockfile", "", "")
	require.NoError(t, err)
	assert.True(t, p.ageFilter.IsZero())
	assert.False(t, p.ageFilter.NeverDowngrade())
}

func TestInstalledVersion(t *testing.T) {
	file := discoveredFile{installed: lockedVersions{"requests": {"2.32.5"}, "numpy": {"2.1.0", "1.26.4", "10.0.0rc1"}}}

	assert.Equal(t, "2.32.5", file.installedVersion(pythonDependency{Name: "Requests", Constraint: ">=2.28"}))
	assert.Equal(t, "10.0.0rc1", file.installedVersion(pythonDependency{Name: "numpy", Constraint: ">=1.20"}))
	assert.Equal(t, "10.0.0rc1", file.installedVersion(pythonDependency{Name: "numpy"}))
	// A declaration under an environment marker only gets the versions within its constraint
	assert.Equal(t, "1.26.4", file.installedVersion(pythonDependency{Name: "numpy", Constraint: ">=1.24,<2"}))
	// uv.lock resolves none within the constraint, which then stays unbounded, even when pinned
	assert.Equal(t, "", file.installedVersion(pythonDependency{Name: "requests", Constraint: "<2"}))
	assert.Equal(t, "", file.installedVersion(pythonDependency{Name: "requests", Constraint: "==2.0.0"}))
	// Without uv.lock, the version pinned with "==" is the version in use
	assert.Equal(t, "8.1.7", file.installedVersion(pythonDependency{Name: "click", Constraint: "==8.1.7"}))
	assert.Equal(t, "", file.installedVersion(pythonDependency{Name: "click", Constraint: ">=8"}))
}

func TestNewAgeNeverDowngradesPEP440Versions(t *testing.T) {
	testdata := []struct {
		name string
		spec Spec
		// expectedPatterns maps a dependency to the pattern of its source versionfilter.
		// A dependency missing from it must be skipped.
		expectedPatterns map[string]string
	}{
		{
			name: "local, post and pre-release versions bound the declared constraint",
			expectedPatterns: map[string]string{
				// A specifier can't name a local version
				"torch": ">=2.1.0, >=2.0",
				"foo":   ">=1.0.post1, >=1.0",
				// A pre-release in use doesn't make newer pre-releases match
				"bar": ">=2.0, >=1.0",
			},
		},
		{
			name: "a semver versionfilter skips the versions in use it can't parse",
			spec: Spec{VersionFilter: version.Filter{Kind: "semver", Pattern: "*"}},
			expectedPatterns: map[string]string{
				"torch": ">=2.1.0+cpu",
			},
		},
		{
			name: "allowdowngrade doesn't need a bound",
			spec: Spec{
				VersionFilter: version.Filter{Kind: "semver", Pattern: "*"},
				Age:           &age.AutodiscoverySpec{Spec: age.Spec{Minimum: "7d"}, AllowDowngrade: new(true)},
			},
			expectedPatterns: map[string]string{
				"torch": "*",
				"foo":   "*",
				"bar":   "*",
			},
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			p, err := New(tt.spec, "testdata/locked_pep440_versions", "", "")
			require.NoError(t, err)
			p.uvAvailable = true

			manifests, err := p.DiscoverManifests()
			require.NoError(t, err)
			require.Len(t, manifests, len(tt.expectedPatterns), "manifests: %s", manifests)

			for _, manifest := range manifests {
				var found bool
				for dependency, pattern := range tt.expectedPatterns {
					if !strings.Contains(string(manifest), fmt.Sprintf("name: '%s'\n", dependency)) {
						continue
					}
					found = true
					assert.Contains(t, string(manifest), fmt.Sprintf("pattern: '%s'\n", pattern), "dependency %q", dependency)
				}
				assert.True(t, found, "unexpected manifest:\n%s", manifest)
			}
		})
	}
}

func TestRoutineVersionFilterPEP440WithoutInstalledVersion(t *testing.T) {
	// The declared version of "foo<3.0" is 3.0, an upper bound: a pep440 filter isn't bounded at it.
	p, err := New(map[string]interface{}{"versionfilter": map[string]interface{}{"kind": "pep440", "pattern": "<3.0"}}, "testdata/no_lockfile", "", "")
	require.NoError(t, err)
	require.True(t, p.ageFilter.NeverDowngrade())

	dep, err := parsePEP508("foo<3.0")
	require.NoError(t, err)

	kind, pattern, _, ok := p.routineVersionFilter(discoveredFile{}, dep)
	require.True(t, ok)
	assert.Equal(t, "pep440", kind)
	assert.Equal(t, "<3.0", pattern)

	// An installed version, here pinned, still bounds it
	dep, err = parsePEP508("foo==2.5.0")
	require.NoError(t, err)
	_, pattern, _, ok = p.routineVersionFilter(discoveredFile{}, dep)
	require.True(t, ok)
	assert.Equal(t, ">=2.5.0, <3.0", pattern)
}

func TestSearchUvLock(t *testing.T) {
	root := "testdata/uv_workspace"

	// The workspace root and its members share the root uv.lock
	assert.Equal(t, root, searchUvLock(root, root))
	assert.Equal(t, root, searchUvLock(root+"/packages/api", root))
	// As in uv, a project nested in another project is its own workspace root, even when an
	// outer workspace matches it: the closest pyproject.toml decides
	assert.Equal(t, "", searchUvLock(root+"/libs/x/sub", root))
	assert.Equal(t, root+"/libs/x", searchUvLock(root+"/libs/x", root))
	// A project the workspace doesn't include, or excludes, has no uv.lock
	assert.Equal(t, "", searchUvLock(root+"/tools/standalone", root))
	assert.Equal(t, "", searchUvLock(root+"/packages/legacy", root))
	// The search doesn't go above the root directory
	assert.Equal(t, "", searchUvLock(root+"/packages/api", root+"/packages"))
	// A uv.lock next to the pyproject.toml always applies
	assert.Equal(t, "testdata/simple_project", searchUvLock("testdata/simple_project", "testdata"))
}

func TestMatchGlobSegments(t *testing.T) {
	assert.True(t, matchGlobSegments([]string{"packages", "*"}, []string{"packages", "api"}))
	assert.False(t, matchGlobSegments([]string{"packages", "*"}, []string{"packages", "api", "sub"}))
	assert.True(t, matchGlobSegments([]string{"libs", "**", "sub"}, []string{"libs", "sub"}))
	assert.True(t, matchGlobSegments([]string{"libs", "**", "sub"}, []string{"libs", "x", "y", "sub"}))
	assert.False(t, matchGlobSegments([]string{"libs", "**", "sub"}, []string{"libs", "x"}))
}

func TestDiscoverManifestsUvWorkspace(t *testing.T) {
	p, err := New(Spec{}, "testdata/uv_workspace", "", "")
	require.NoError(t, err)
	p.uvAvailable = true

	manifests, err := p.DiscoverManifests()
	require.NoError(t, err)

	byName := map[string]string{}
	for _, manifest := range manifests {
		name := strings.SplitN(string(manifest), "\n", 2)[0]
		byName[name] = string(manifest)
	}

	// requests, declared by the root and a member, is re-locked by a single manifest, the
	// first one found, rather than by one per project opening conflicting pull requests. Its
	// version must satisfy both declarations, ">=2.28" and ">=2.28,<2.33".
	require.Len(t, byName, 3, "manifests: %v", byName)
	apiRequests, found := byName[`name: 'deps(pypi): bump "requests" for "api" project'`]
	require.True(t, found, "manifests: %v", byName)
	assert.Contains(t, apiRequests, "pattern: '>=2.32.5, >=2.28,>=2.28,<2.33'\n")
	assert.NotContains(t, byName, `name: 'deps(pypi): bump "requests" for "ws-root" project'`)

	// A workspace member is bounded at the version the root uv.lock resolves, and uv runs at the root
	api := byName[`name: 'deps(pypi): bump "fastapi" for "api" project'`]
	assert.Contains(t, api, "pattern: '>=0.115.0, >=0.100'\n")
	assert.Contains(t, api, "workdir: '.'\n")

	// A project the workspace doesn't include gets no uv target and no bound
	standalone := byName[`name: 'deps(pypi): bump "flask" for "standalone" project'`]
	assert.Contains(t, standalone, "pattern: '>=3.0'\n")
	assert.NotContains(t, standalone, "targets:")
}

func TestDiscoverSecurityManifestsUvWorkspace(t *testing.T) {
	p, err := New(map[string]interface{}{"vulnerability": map[string]interface{}{}}, "testdata/uv_workspace", "", "")
	require.NoError(t, err)
	p.uvAvailable = true

	manifests, err := p.DiscoverManifests()
	require.NoError(t, err)

	// requests, declared by the root and a member, is updated once in the shared uv.lock
	var names []string
	for _, manifest := range manifests {
		names = append(names, strings.SplitN(string(manifest), "\n", 2)[0])
		assert.Contains(t, string(manifest), "workdir: '.'\n")
	}
	assert.ElementsMatch(t, []string{
		`name: 'deps(pypi): bump "fastapi" to {{ source "fastapi" }}'`,
		`name: 'deps(pypi): bump "requests" to {{ source "requests" }}'`,
	}, names)

	for _, manifest := range manifests {
		if strings.Contains(string(manifest), `"fastapi"`) {
			assert.Contains(t, string(manifest), "version: '0.115.0'\n")
		}
	}
}

func TestRoutineVersionFilterSemverWithoutLowerBound(t *testing.T) {
	// "foo<3.0" sets no lower bound, so a semver filter isn't anchored at 3.0
	p, err := New(map[string]interface{}{"versionfilter": map[string]interface{}{"kind": "semver", "pattern": "minor"}}, "testdata/no_lockfile", "", "")
	require.NoError(t, err)

	dep, err := parsePEP508("foo<3.0")
	require.NoError(t, err)
	_, pattern, _, ok := p.routineVersionFilter(discoveredFile{}, dep)
	assert.False(t, ok, "a relative pattern without a version to start from must be skipped, not %q", pattern)

	// A pattern which isn't relative applies as is
	p.versionFilter.Pattern = ">=1.0.0"
	_, pattern, _, ok = p.routineVersionFilter(discoveredFile{}, dep)
	require.True(t, ok)
	assert.Equal(t, ">=1.0.0", pattern)
	p.versionFilter.Pattern = "minor"

	// ">=2.28,<3" starts from 2.28
	dep, err = parsePEP508("foo>=2.28,<3")
	require.NoError(t, err)
	_, pattern, _, ok = p.routineVersionFilter(discoveredFile{}, dep)
	require.True(t, ok)
	assert.Equal(t, ">=2.28.0, 2.x", pattern)

	// A post release starts from its release
	dep, err = parsePEP508("foo>=2.28.post1")
	require.NoError(t, err)
	_, pattern, _, ok = p.routineVersionFilter(discoveredFile{}, dep)
	require.True(t, ok)
	assert.Equal(t, ">=2.28.0, 2.x", pattern)

	// A start semver can't parse, such as one with an epoch, skips a relative pattern
	dep, err = parsePEP508("foo>=1!2.0")
	require.NoError(t, err)
	_, pattern, _, ok = p.routineVersionFilter(discoveredFile{}, dep)
	assert.False(t, ok, "a relative pattern that can't start from the constraint must be skipped, not %q", pattern)

	// With uv.lock, the installed version is the start
	dep, err = parsePEP508("foo<3.0")
	require.NoError(t, err)
	_, pattern, _, ok = p.routineVersionFilter(discoveredFile{installed: lockedVersions{"foo": {"2.5.0"}}}, dep)
	require.True(t, ok)
	assert.Equal(t, ">=2.5.0, 2.x", pattern)
}

func TestDiscoverManifestsUvWorkspaceMemberRootDir(t *testing.T) {
	// "rootdir" pointing at a workspace member still finds the uv.lock at the workspace root
	p, err := New(Spec{RootDir: "packages/api"}, "testdata/uv_workspace", "", "")
	require.NoError(t, err)
	p.uvAvailable = true

	manifests, err := p.DiscoverManifests()
	require.NoError(t, err)
	require.Len(t, manifests, 2)

	for _, manifest := range manifests {
		assert.Contains(t, string(manifest), "workdir: '.'\n")
		assert.Contains(t, string(manifest), "uv lock --upgrade-package")
	}
	assert.Contains(t, string(manifests[0])+string(manifests[1]), "pattern: '>=0.115.0, >=0.100'\n")
}

func TestLockStateCombinedConstraint(t *testing.T) {
	lock := newLockState()
	lock.addConstraints(pyprojectTOML{})
	data, err := loadPyprojectData("testdata/uv_workspace/packages/api/pyproject.toml")
	require.NoError(t, err)
	lock.addConstraints(data)
	data, err = loadPyprojectData("testdata/uv_workspace/pyproject.toml")
	require.NoError(t, err)
	lock.addConstraints(data)

	// Every distinct declaration counts, in a stable order, so the source picks a version all accept
	assert.Equal(t, ">=2.28,>=2.28,<2.33", lock.combinedConstraint("Requests", ""))
	assert.Equal(t, ">=2.28,>=2.28,<2.33", lock.combinedConstraint("Requests", "2.32.5"))
	assert.Equal(t, ">=0.100", lock.combinedConstraint("fastapi", ""))
	assert.Equal(t, "", lock.combinedConstraint("unknown", ""))

	// A declaration uv doesn't apply, which the locked version doesn't satisfy, is left out
	lock.constraints["foo"] = []string{">=2", "<2"}
	assert.Equal(t, ">=2", lock.combinedConstraint("foo", "2.5"))
	assert.Equal(t, "<2,>=2", lock.combinedConstraint("foo", ""))

	// Relative patterns start from the highest lower bound of the combined constraint
	assert.Equal(t, "2.10", declaredLowerBound(">=2.0,<2.30,>=2.10"))
}
