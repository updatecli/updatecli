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
				IgnoreVersionConstraints: boolPtr(true),
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
				IgnoreVersionConstraints: boolPtr(true),
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
				IgnoreVersionConstraints: boolPtr(false),
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
		{
			name:    "Scenario 2 -- pnpm",
			rootDir: "testdata/pnpmlockfile",
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
  pnpm-lock.yaml:
    name: 'Bump "@mdi/font" package version to {{ source "npm" }}'
    disablesourceinput: true
    kind: shell
    spec:
      command: |-
        pnpm add --lockfile-only @mdi/font@{{ source "npm" }}
      changedif:
        kind: file/checksum
        spec:
          files:
            - "pnpm-lock.yaml"
            - "package.json"
      environments:
       - name: PATH
      workdir: '.'

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

func TestNew_WithResourceConfig(t *testing.T) {
	tests := []struct {
		name              string
		spec              map[string]interface{}
		expectedNpmrcPath string
		expectedURL       string
		expectedToken     string
	}{
		{
			name: "All resource config fields set",
			spec: map[string]interface{}{
				"rootdir":       ".",
				"npmrcpath":     "/custom/.npmrc",
				"url":           "https://npm.example.com",
				"registrytoken": "test-token",
			},
			expectedNpmrcPath: "/custom/.npmrc",
			expectedURL:       "https://npm.example.com",
			expectedToken:     "test-token",
		},
		{
			name: "Only npmrcpath set",
			spec: map[string]interface{}{
				"rootdir":   ".",
				"npmrcpath": "/custom/.npmrc",
			},
			expectedNpmrcPath: "/custom/.npmrc",
			expectedURL:       "",
			expectedToken:     "",
		},
		{
			name: "No resource config fields",
			spec: map[string]interface{}{
				"rootdir": ".",
			},
			expectedNpmrcPath: "",
			expectedURL:       "",
			expectedToken:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			npm, err := New(tt.spec, ".", "", "")
			require.NoError(t, err)

			assert.Equal(t, tt.expectedNpmrcPath, npm.npmrcPath)
			assert.Equal(t, tt.expectedURL, npm.url)
			assert.Equal(t, tt.expectedToken, npm.registryToken)
		})
	}
}

func TestDiscoverManifests_WithResourceConfig(t *testing.T) {
	spec := Spec{
		NpmrcPath:     "/custom/.npmrc",
		URL:           "https://npm.example.com",
		RegistryToken: "test-token-123",
	}

	resource, err := New(spec, "testdata/npmlockfile", "", "")
	require.NoError(t, err)

	manifests, err := resource.DiscoverManifests()
	require.NoError(t, err)
	require.Greater(t, len(manifests), 0)

	// Verify first manifest contains resource-level config
	manifestStr := string(manifests[0])
	assert.Contains(t, manifestStr, "npmrcpath: '/custom/.npmrc'")
	assert.Contains(t, manifestStr, "url: 'https://npm.example.com'")
	assert.Contains(t, manifestStr, "registrytoken: 'test-token-123'")
}
