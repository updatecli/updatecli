package npm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func boolPtr(b bool) *bool {
	return &b
}

func TestDiscoverManifests(t *testing.T) {

	testdata := []struct {
		name              string
		rootDir           string
		spec              Spec
		expectedPipelines []string
	}{
		{
			name:    "Npm lockfile without respect version constraint with minor version update",
			rootDir: "testdata/npmlockfile",
			spec: Spec{
				ignoreVersionConstraint: boolPtr(true),
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "minoronly",
				},
			},
			expectedPipelines: []string{`name: 'Bump "axios" package version'
sources:
  npm:
    name: 'Get "axios" package version'
    kind: 'npm'
    spec:
      name: 'axios'
      versionfilter:
        kind: 'semver'
        pattern: '1.0.0 || >1.0 < 2'
targets:
  package-lock.json:
    name: 'Bump "axios" package version to {{ source "npm" }}'
    disablesourceinput: true
    kind: shell
    spec:
      command: |-
        npm install --package-lock-only --dry-run=$DRY_RUN axios@{{ source "npm" }}
      changedif:
        kind: file/checksum
        spec:
          files:
            - "package-lock.json"
            - "package.json"
      environments:
       - name: PATH
      workdir: '.'

`,
			},
		},
		{
			name:    "Npm lockfile without respect version constraint",
			rootDir: "testdata/npmlockfile",
			spec: Spec{
				ignoreVersionConstraint: boolPtr(true),
			},
			expectedPipelines: []string{`name: 'Bump "axios" package version'
sources:
  npm:
    name: 'Get "axios" package version'
    kind: 'npm'
    spec:
      name: 'axios'
      versionfilter:
        kind: 'semver'
        pattern: '*'
targets:
  package-lock.json:
    name: 'Bump "axios" package version to {{ source "npm" }}'
    disablesourceinput: true
    kind: shell
    spec:
      command: |-
        npm install --package-lock-only --dry-run=$DRY_RUN axios@{{ source "npm" }}
      changedif:
        kind: file/checksum
        spec:
          files:
            - "package-lock.json"
            - "package.json"
      environments:
       - name: PATH
      workdir: '.'

`,
			},
		},
		{
			name:    "Npm lockfile with respect version constraint",
			rootDir: "testdata/npmlockfile",
			spec: Spec{
				ignoreVersionConstraint: boolPtr(false),
			},
			expectedPipelines: []string{`name: 'Bump "axios" package version'
sources:
  npm:
    name: 'Get "axios" package version'
    kind: 'npm'
    spec:
      name: 'axios'
      versionfilter:
        kind: 'semver'
        pattern: '^1.0.0'
targets:
  package-lock.json:
    name: 'Bump "axios" package version to {{ source "npm" }}'
    disablesourceinput: true
    kind: shell
    spec:
      command: |-
        npm install --package-lock-only --dry-run=$DRY_RUN axios@{{ source "npm" }}
      changedif:
        kind: file/checksum
        spec:
          files:
            - "package-lock.json"
            - "package.json"
      environments:
       - name: PATH
      workdir: '.'

`,
			},
		},
		{
			name:    "Scenario 1",
			rootDir: "testdata/nolockfile",
			expectedPipelines: []string{`name: 'Bump "@mdi/font" package version'
sources:
  npm:
    name: 'Get "@mdi/font" package version'
    kind: 'npm'
    spec:
      name: '@mdi/font'
      versionfilter:
        kind: 'semver'
        pattern: '>=5.9.55'
targets:
  npm:
    name: 'Bump "@mdi/font" package version to {{ source "npm" }}'
    kind: 'json'
    spec:
      file: 'package.json'
      key: 'dependencies.@mdi/font'
    sourceid: 'npm'

`,
			},
		},
	}

	for _, tt := range testdata {

		t.Run(tt.name, func(t *testing.T) {
			resource, err := New(
				tt.spec, tt.rootDir, "", "")
			require.NoError(t, err)

			pipelines, err := resource.DiscoverManifests()
			require.NoError(t, err)

			assert.Equal(t, len(tt.expectedPipelines), len(pipelines))

			for i, expectedPipeline := range tt.expectedPipelines {
				assert.Equal(t, expectedPipeline, string(pipelines[i]))
			}
		})
	}

}
