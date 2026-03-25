package pyproject

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
      versionfilter:
        kind: 'pep440'
        pattern: '>=3.0'
targets:
  flask:
    name: 'deps(pypi): bump "flask" to {{ source "flask" }}'
    kind: 'shell'
    spec:
      command: 'uv add "flask>={{ source "flask" }}"'
      changedif:
        kind: file/checksum
        spec:
          files:
            - "pyproject.toml"
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
      versionfilter:
        kind: 'pep440'
        pattern: '>=2.28'
targets:
  requests:
    name: 'deps(pypi): bump "requests" to {{ source "requests" }}'
    kind: 'shell'
    spec:
      command: 'uv add "requests>={{ source "requests" }}"'
      changedif:
        kind: file/checksum
        spec:
          files:
            - "pyproject.toml"
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
      versionfilter:
        kind: 'pep440'
        pattern: '>=2.28'
targets:
  requests:
    name: 'deps(pypi): bump "requests" to {{ source "requests" }}'
    kind: 'shell'
    spec:
      command: 'uv add "requests>={{ source "requests" }}"'
      changedif:
        kind: file/checksum
        spec:
          files:
            - "pyproject.toml"
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
      versionfilter:
        kind: 'pep440'
        pattern: '>=8.0'
targets:
  pytest:
    name: 'deps(pypi): bump "pytest" to {{ source "pytest" }}'
    kind: 'shell'
    spec:
      command: 'uv add --optional dev "pytest>={{ source "pytest" }}"'
      changedif:
        kind: file/checksum
        spec:
          files:
            - "pyproject.toml"
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
      versionfilter:
        kind: 'pep440'
        pattern: '>=0.4'
targets:
  ruff:
    name: 'deps(pypi): bump "ruff" to {{ source "ruff" }}'
    kind: 'shell'
    spec:
      command: 'uv add --optional dev "ruff>={{ source "ruff" }}"'
      changedif:
        kind: file/checksum
        spec:
          files:
            - "pyproject.toml"
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
      versionfilter:
        kind: 'pep440'
        pattern: '>=3.0'
targets:
  flask:
    name: 'deps(pypi): bump "flask" to {{ source "flask" }}'
    scmid: 'git'
    kind: 'shell'
    spec:
      command: 'uv add "flask>={{ source "flask" }}"'
      changedif:
        kind: file/checksum
        spec:
          files:
            - "pyproject.toml"
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
      versionfilter:
        kind: 'pep440'
        pattern: '>=2.28'
targets:
  requests:
    name: 'deps(pypi): bump "requests" to {{ source "requests" }}'
    scmid: 'git'
    kind: 'shell'
    spec:
      command: 'uv add "requests>={{ source "requests" }}"'
      changedif:
        kind: file/checksum
        spec:
          files:
            - "pyproject.toml"
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
      versionfilter:
        kind: 'pep440'
        pattern: '>=2.28'
targets:
  requests:
    name: 'deps(pypi): bump "requests" to {{ source "requests" }}'
    kind: 'shell'
    spec:
      command: 'uv add "requests>={{ source "requests" }}"'
      changedif:
        kind: file/checksum
        spec:
          files:
            - "pyproject.toml"
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
      versionfilter:
        kind: 'pep440'
        pattern: '>=2.28'
targets:
  requests:
    name: 'deps(pypi): bump "requests" to {{ source "requests" }}'
    kind: 'shell'
    spec:
      command: 'uv add "requests>={{ source "requests" }}"'
      changedif:
        kind: file/checksum
        spec:
          files:
            - "pyproject.toml"
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
      versionfilter:
        kind: 'pep440'
        pattern: '>=300'
targets:
  pywin32:
    name: 'deps(pypi): bump "pywin32" to {{ source "pywin32" }}'
    kind: 'shell'
    spec:
      command: 'uv add "pywin32>={{ source "pywin32" }}"'
      changedif:
        kind: file/checksum
        spec:
          files:
            - "pyproject.toml"
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
      versionfilter:
        kind: 'pep440'
        pattern: '>=2.28'
targets:
  requests:
    name: 'deps(pypi): bump "requests" to {{ source "requests" }}'
    kind: 'shell'
    spec:
      command: 'uv add "requests>={{ source "requests" }}"'
      changedif:
        kind: file/checksum
        spec:
          files:
            - "pyproject.toml"
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
      versionfilter:
        kind: 'pep440'
        pattern: '*'
targets:
  numpy:
    name: 'deps(pypi): bump "numpy" to {{ source "numpy" }}'
    kind: 'shell'
    spec:
      command: 'uv add "numpy>={{ source "numpy" }}"'
      changedif:
        kind: file/checksum
        spec:
          files:
            - "pyproject.toml"
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
      command: 'uv add "flask>={{ source "flask" }}"'
      changedif:
        kind: file/checksum
        spec:
          files:
            - "pyproject.toml"
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
      command: 'uv add "requests>={{ source "requests" }}"'
      changedif:
        kind: file/checksum
        spec:
          files:
            - "pyproject.toml"
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
      versionfilter:
        kind: 'semver'
        pattern: 'minor'
targets:
  flask:
    name: 'deps(pypi): bump "flask" to {{ source "flask" }}'
    kind: 'shell'
    spec:
      command: 'uv add "flask>={{ source "flask" }}"'
      changedif:
        kind: file/checksum
        spec:
          files:
            - "pyproject.toml"
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
      versionfilter:
        kind: 'semver'
        pattern: 'minor'
targets:
  requests:
    name: 'deps(pypi): bump "requests" to {{ source "requests" }}'
    kind: 'shell'
    spec:
      command: 'uv add "requests>={{ source "requests" }}"'
      changedif:
        kind: file/checksum
        spec:
          files:
            - "pyproject.toml"
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
