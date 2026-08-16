package npm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"

	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
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
      age:
        minimum: '3d'
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
			name:    "Npm lockfile with a release age filter",
			rootDir: "testdata/npmlockfile",
			spec: Spec{
				IgnoreVersionConstraints: boolPtr(true),
				Age: &age.Spec{
					Minimum: "7d",
					Maximum: "1y",
				},
			},
			expectedPipelines: []string{`name: 'Bump "axios" package version'
sources:
  npm:
    name: 'Get "axios" package version'
    kind: 'npm'
    spec:
      name: 'axios'
      age:
        minimum: '7d'
        maximum: '1y'
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
			name:    "Npm lockfile with an empty release age filter disabling the default one",
			rootDir: "testdata/npmlockfile",
			spec: Spec{
				IgnoreVersionConstraints: boolPtr(true),
				Age:                      &age.Spec{},
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
			name:    "Npm lockfile with a maximum release age only",
			rootDir: "testdata/npmlockfile",
			spec: Spec{
				IgnoreVersionConstraints: boolPtr(true),
				Age:                      &age.Spec{Maximum: "1y"},
			},
			expectedPipelines: []string{`name: 'Bump "axios" package version'
sources:
  npm:
    name: 'Get "axios" package version'
    kind: 'npm'
    spec:
      name: 'axios'
      age:
        maximum: '1y'
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
      age:
        minimum: '3d'
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
      age:
        minimum: '3d'
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
      age:
        minimum: '3d'
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
      age:
        minimum: '3d'
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

func TestNewReleaseAge(t *testing.T) {
	tests := []struct {
		name               string
		spec               Spec
		expectedReleaseAge age.Spec
		expectedError      bool
	}{
		{
			name:               "No age specified fallbacks to the default minimum release age",
			spec:               Spec{},
			expectedReleaseAge: age.Spec{Minimum: "3d"},
		},
		{
			name:               "An empty age disables the default minimum release age",
			spec:               Spec{Age: &age.Spec{}},
			expectedReleaseAge: age.Spec{},
		},
		{
			name:               "A specified minimum age overrides the default one",
			spec:               Spec{Age: &age.Spec{Minimum: "7d"}},
			expectedReleaseAge: age.Spec{Minimum: "7d"},
		},
		{
			name:               "A specified maximum age doesn't reintroduce the default minimum one",
			spec:               Spec{Age: &age.Spec{Maximum: "1y"}},
			expectedReleaseAge: age.Spec{Maximum: "1y"},
		},
		{
			name:          "An invalid age is reported",
			spec:          Spec{Age: &age.Spec{Minimum: "notaduration"}},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			npm, err := New(tt.spec, ".", "", "")
			if tt.expectedError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			assert.Equal(t, tt.expectedReleaseAge, npm.releaseAge)
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
	assert.Contains(t, manifestStr, "- name: NPM_CONFIG_USERCONFIG")
	assert.Contains(t, manifestStr, "value: '/custom/.npmrc'")

	var parsedManifest struct {
		Targets map[string]struct {
			Spec struct {
				Environments []struct {
					Name  string `yaml:"name"`
					Value string `yaml:"value"`
				} `yaml:"environments"`
			} `yaml:"spec"`
		} `yaml:"targets"`
	}
	require.NoError(t, yaml.Unmarshal(manifests[0], &parsedManifest))

	foundNpmrcEnv := false
	for _, target := range parsedManifest.Targets {
		for _, env := range target.Spec.Environments {
			if env.Name == "NPM_CONFIG_USERCONFIG" {
				assert.Equal(t, "/custom/.npmrc", env.Value)
				foundNpmrcEnv = true
			}
		}
	}
	assert.True(t, foundNpmrcEnv, "expected NPM_CONFIG_USERCONFIG in at least one target environment")
}
