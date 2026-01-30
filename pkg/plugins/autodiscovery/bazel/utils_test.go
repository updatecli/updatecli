package bazel

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func TestFindModuleFiles(t *testing.T) {
	testdata := []struct {
		name          string
		rootDir       string
		expectedFiles []string
		expectError   bool
	}{
		{
			name:    "Find MODULE.bazel files in testdata",
			rootDir: "testdata",
			expectedFiles: []string{
				"testdata/project1/MODULE.bazel",
				"testdata/project2/MODULE.bazel",
				"testdata/project2/subdir/MODULE.bazel",
			},
			expectError: false,
		},
		{
			name:          "Non-existent directory",
			rootDir:       "nonexistent",
			expectedFiles: []string{},
			expectError:   true,
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			files, err := findModuleFiles(tt.rootDir)
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.GreaterOrEqual(t, len(files), len(tt.expectedFiles))

			// Normalize paths for comparison
			normalizedFiles := make([]string, len(files))
			for i, f := range files {
				rel, err := filepath.Rel(".", f)
				if err == nil {
					normalizedFiles[i] = rel
				} else {
					normalizedFiles[i] = f
				}
			}

			// Check that all expected files are found
			for _, expected := range tt.expectedFiles {
				found := false
				for _, f := range normalizedFiles {
					if filepath.Base(f) == filepath.Base(expected) {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected file %q not found", expected)
			}
		})
	}
}

func TestShouldIgnore(t *testing.T) {
	testdata := []struct {
		name        string
		moduleName  string
		version     string
		filePath    string
		rootDir     string
		rules       MatchingRules
		expectError bool
	}{
		{
			name:        "No rules - should not ignore",
			moduleName:  "rules_go",
			version:     "0.42.0",
			filePath:    "testdata/project1/MODULE.bazel",
			rootDir:     "testdata",
			rules:       MatchingRules{},
			expectError: false,
		},
		{
			name:       "Ignore by module name",
			moduleName: "rules_go",
			version:    "0.42.0",
			filePath:   "testdata/project1/MODULE.bazel",
			rootDir:    "testdata",
			rules: MatchingRules{
				MatchingRule{
					Modules: map[string]string{
						"rules_go": "",
					},
				},
			},
			expectError: true,
		},
		{
			name:       "Ignore by path pattern",
			moduleName: "rules_go",
			version:    "0.42.0",
			filePath:   "testdata/project1/MODULE.bazel",
			rootDir:    "testdata",
			rules: MatchingRules{
				MatchingRule{
					Path: "project1/*",
				},
			},
			expectError: true,
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			// Create absolute paths for testing
			absRoot, err := filepath.Abs(tt.rootDir)
			require.NoError(t, err)
			absPath, err := filepath.Abs(tt.filePath)
			require.NoError(t, err)

			result := shouldIgnore(tt.moduleName, tt.version, absPath, absRoot, tt.rules)
			assert.Equal(t, tt.expectError, result)
		})
	}
}

func TestShouldInclude(t *testing.T) {
	testdata := []struct {
		name        string
		moduleName  string
		version     string
		filePath    string
		rootDir     string
		rules       MatchingRules
		expectError bool
	}{
		{
			name:        "No rules - should include",
			moduleName:  "rules_go",
			version:     "0.42.0",
			filePath:    "testdata/project1/MODULE.bazel",
			rootDir:     "testdata",
			rules:       MatchingRules{},
			expectError: true,
		},
		{
			name:       "Include by module name",
			moduleName: "rules_go",
			version:    "0.42.0",
			filePath:   "testdata/project1/MODULE.bazel",
			rootDir:    "testdata",
			rules: MatchingRules{
				MatchingRule{
					Modules: map[string]string{
						"rules_go": "",
					},
				},
			},
			expectError: true,
		},
		{
			name:       "Exclude by module name",
			moduleName: "gazelle",
			version:    "0.34.0",
			filePath:   "testdata/project1/MODULE.bazel",
			rootDir:    "testdata",
			rules: MatchingRules{
				MatchingRule{
					Modules: map[string]string{
						"rules_go": "",
					},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			// Create absolute paths for testing
			absRoot, err := filepath.Abs(tt.rootDir)
			require.NoError(t, err)
			absPath, err := filepath.Abs(tt.filePath)
			require.NoError(t, err)

			result := shouldInclude(tt.moduleName, tt.version, absPath, absRoot, tt.rules)
			assert.Equal(t, tt.expectError, result)
		})
	}
}

func TestFindModuleFilesSkipsHiddenDirs(t *testing.T) {
	// Create a temporary directory structure with hidden directories
	tmpDir := t.TempDir()

	// Create normal directory
	normalDir := filepath.Join(tmpDir, "normal")
	err := os.MkdirAll(normalDir, 0755)
	require.NoError(t, err)

	// Create MODULE.bazel in normal directory
	normalFile := filepath.Join(normalDir, "MODULE.bazel")
	err = os.WriteFile(normalFile, []byte("module(name = \"test\")\n"), 0600)
	require.NoError(t, err)

	// Create hidden directory
	hiddenDir := filepath.Join(tmpDir, ".hidden")
	err = os.MkdirAll(hiddenDir, 0755)
	require.NoError(t, err)

	// Create MODULE.bazel in hidden directory
	hiddenFile := filepath.Join(hiddenDir, "MODULE.bazel")
	err = os.WriteFile(hiddenFile, []byte("module(name = \"hidden\")\n"), 0600)
	require.NoError(t, err)

	// Find files
	files, err := findModuleFiles(tmpDir)
	require.NoError(t, err)

	// Should only find the file in the normal directory, not the hidden one
	assert.Len(t, files, 1)
	assert.Contains(t, files[0], "normal")
	assert.NotContains(t, files[0], ".hidden")
}

func TestDiscoverManifestsFullOutput(t *testing.T) {
	testdata := []struct {
		name              string
		rootDir           string
		spec              Spec
		scmID             string
		actionID          string
		expectedManifests []string
	}{
		{
			name:    "Scenario 1 - simple project1 with default version filter",
			rootDir: "testdata/project1",
			spec: Spec{
				RootDir: "",
			},
			scmID:    "",
			actionID: "",
			expectedManifests: []string{`name: 'Update Bazel module rules_go'
sources:
  rules_go:
    name: 'Get latest version of Bazel module rules_go'
    kind: bazelregistry
    spec:
      module: rules_go
      versionfilter:
        kind: 'semver'
        pattern: '>=0.42.0'
conditions:
  rules_go:
    name: 'Check if Bazel module rules_go is up to date'
    kind: bazelmod
    spec:
      file: 'MODULE.bazel'
      module: rules_go
    disablesourceinput: true
targets:
  rules_go:
    name: 'Bump Bazel module rules_go to {{ source "rules_go" }}'
    kind: bazelmod
    spec:
      file: 'MODULE.bazel'
      module: rules_go
    sourceid: 'rules_go'
`, `name: 'Update Bazel module gazelle'
sources:
  gazelle:
    name: 'Get latest version of Bazel module gazelle'
    kind: bazelregistry
    spec:
      module: gazelle
      versionfilter:
        kind: 'semver'
        pattern: '>=0.34.0'
conditions:
  gazelle:
    name: 'Check if Bazel module gazelle is up to date'
    kind: bazelmod
    spec:
      file: 'MODULE.bazel'
      module: gazelle
    disablesourceinput: true
targets:
  gazelle:
    name: 'Bump Bazel module gazelle to {{ source "gazelle" }}'
    kind: bazelmod
    spec:
      file: 'MODULE.bazel'
      module: gazelle
    sourceid: 'gazelle'
`, `name: 'Update Bazel module protobuf'
sources:
  protobuf:
    name: 'Get latest version of Bazel module protobuf'
    kind: bazelregistry
    spec:
      module: protobuf
      versionfilter:
        kind: 'semver'
        pattern: '>=21.7.0'
conditions:
  protobuf:
    name: 'Check if Bazel module protobuf is up to date'
    kind: bazelmod
    spec:
      file: 'MODULE.bazel'
      module: protobuf
    disablesourceinput: true
targets:
  protobuf:
    name: 'Bump Bazel module protobuf to {{ source "protobuf" }}'
    kind: bazelmod
    spec:
      file: 'MODULE.bazel'
      module: protobuf
    sourceid: 'protobuf'
`,
			},
		},
		{
			name:    "Scenario 2 - project1 with scmID and actionID",
			rootDir: "testdata/project1",
			spec: Spec{
				RootDir: "",
			},
			scmID:    "defaultscmid",
			actionID: "defaultactionid",
			expectedManifests: []string{`name: 'Update Bazel module rules_go'
actions:
  defaultactionid:
    title: 'Bump Bazel module rules_go to {{ source "rules_go" }}'

sources:
  rules_go:
    name: 'Get latest version of Bazel module rules_go'
    kind: bazelregistry
    spec:
      module: rules_go
      versionfilter:
        kind: 'semver'
        pattern: '>=0.42.0'
conditions:
  rules_go:
    name: 'Check if Bazel module rules_go is up to date'
    kind: bazelmod
    scmid: 'defaultscmid'

    spec:
      file: 'MODULE.bazel'
      module: rules_go
    disablesourceinput: true
targets:
  rules_go:
    name: 'Bump Bazel module rules_go to {{ source "rules_go" }}'
    kind: bazelmod
    scmid: 'defaultscmid'

    spec:
      file: 'MODULE.bazel'
      module: rules_go
    sourceid: 'rules_go'
`, `name: 'Update Bazel module gazelle'
actions:
  defaultactionid:
    title: 'Bump Bazel module gazelle to {{ source "gazelle" }}'

sources:
  gazelle:
    name: 'Get latest version of Bazel module gazelle'
    kind: bazelregistry
    spec:
      module: gazelle
      versionfilter:
        kind: 'semver'
        pattern: '>=0.34.0'
conditions:
  gazelle:
    name: 'Check if Bazel module gazelle is up to date'
    kind: bazelmod
    scmid: 'defaultscmid'

    spec:
      file: 'MODULE.bazel'
      module: gazelle
    disablesourceinput: true
targets:
  gazelle:
    name: 'Bump Bazel module gazelle to {{ source "gazelle" }}'
    kind: bazelmod
    scmid: 'defaultscmid'

    spec:
      file: 'MODULE.bazel'
      module: gazelle
    sourceid: 'gazelle'
`, `name: 'Update Bazel module protobuf'
actions:
  defaultactionid:
    title: 'Bump Bazel module protobuf to {{ source "protobuf" }}'

sources:
  protobuf:
    name: 'Get latest version of Bazel module protobuf'
    kind: bazelregistry
    spec:
      module: protobuf
      versionfilter:
        kind: 'semver'
        pattern: '>=21.7.0'
conditions:
  protobuf:
    name: 'Check if Bazel module protobuf is up to date'
    kind: bazelmod
    scmid: 'defaultscmid'

    spec:
      file: 'MODULE.bazel'
      module: protobuf
    disablesourceinput: true
targets:
  protobuf:
    name: 'Bump Bazel module protobuf to {{ source "protobuf" }}'
    kind: bazelmod
    scmid: 'defaultscmid'

    spec:
      file: 'MODULE.bazel'
      module: protobuf
    sourceid: 'protobuf'
`,
			},
		},
		{
			name:    "Scenario 3 - project2/subdir with rules_docker",
			rootDir: "testdata/project2/subdir",
			spec: Spec{
				RootDir: "",
			},
			scmID:    "",
			actionID: "",
			expectedManifests: []string{`name: 'Update Bazel module rules_docker'
sources:
  rules_docker:
    name: 'Get latest version of Bazel module rules_docker'
    kind: bazelregistry
    spec:
      module: rules_docker
      versionfilter:
        kind: 'semver'
        pattern: '>=0.26.0'
conditions:
  rules_docker:
    name: 'Check if Bazel module rules_docker is up to date'
    kind: bazelmod
    spec:
      file: 'MODULE.bazel'
      module: rules_docker
    disablesourceinput: true
targets:
  rules_docker:
    name: 'Bump Bazel module rules_docker to {{ source "rules_docker" }}'
    kind: bazelmod
    spec:
      file: 'MODULE.bazel'
      module: rules_docker
    sourceid: 'rules_docker'
`,
			},
		},
		{
			name:    "Scenario 4 - custom version filter minor",
			rootDir: "testdata/project1",
			spec: Spec{
				RootDir: "",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "minor",
				},
			},
			scmID:    "",
			actionID: "",
			expectedManifests: []string{`name: 'Update Bazel module rules_go'
sources:
  rules_go:
    name: 'Get latest version of Bazel module rules_go'
    kind: bazelregistry
    spec:
      module: rules_go
      versionfilter:
        kind: 'semver'
        pattern: '0.x'
conditions:
  rules_go:
    name: 'Check if Bazel module rules_go is up to date'
    kind: bazelmod
    spec:
      file: 'MODULE.bazel'
      module: rules_go
    disablesourceinput: true
targets:
  rules_go:
    name: 'Bump Bazel module rules_go to {{ source "rules_go" }}'
    kind: bazelmod
    spec:
      file: 'MODULE.bazel'
      module: rules_go
    sourceid: 'rules_go'
`, `name: 'Update Bazel module gazelle'
sources:
  gazelle:
    name: 'Get latest version of Bazel module gazelle'
    kind: bazelregistry
    spec:
      module: gazelle
      versionfilter:
        kind: 'semver'
        pattern: '0.x'
conditions:
  gazelle:
    name: 'Check if Bazel module gazelle is up to date'
    kind: bazelmod
    spec:
      file: 'MODULE.bazel'
      module: gazelle
    disablesourceinput: true
targets:
  gazelle:
    name: 'Bump Bazel module gazelle to {{ source "gazelle" }}'
    kind: bazelmod
    spec:
      file: 'MODULE.bazel'
      module: gazelle
    sourceid: 'gazelle'
`, `name: 'Update Bazel module protobuf'
sources:
  protobuf:
    name: 'Get latest version of Bazel module protobuf'
    kind: bazelregistry
    spec:
      module: protobuf
      versionfilter:
        kind: 'semver'
        pattern: '21.x'
conditions:
  protobuf:
    name: 'Check if Bazel module protobuf is up to date'
    kind: bazelmod
    spec:
      file: 'MODULE.bazel'
      module: protobuf
    disablesourceinput: true
targets:
  protobuf:
    name: 'Bump Bazel module protobuf to {{ source "protobuf" }}'
    kind: bazelmod
    spec:
      file: 'MODULE.bazel'
      module: protobuf
    sourceid: 'protobuf'
`,
			},
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			absRootDir, err := filepath.Abs(tt.rootDir)
			require.NoError(t, err)

			bazel, err := New(tt.spec, absRootDir, tt.scmID, tt.actionID)
			require.NoError(t, err)

			rawManifests, err := bazel.DiscoverManifests()
			require.NoError(t, err)

			if len(rawManifests) == 0 {
				t.Errorf("No manifests found for %s", tt.name)
			}

			var manifests []string
			assert.Equal(t, len(tt.expectedManifests), len(rawManifests), "Number of manifests should match")

			for i := range rawManifests {
				manifests = append(manifests, string(rawManifests[i]))
				assert.Equal(t, tt.expectedManifests[i], manifests[i], "Manifest %d should match expected output", i)
			}
		})
	}
}
