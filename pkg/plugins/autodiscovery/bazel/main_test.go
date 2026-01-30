package bazel

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func TestNew(t *testing.T) {
	testdata := []struct {
		name        string
		spec        Spec
		rootDir     string
		scmID       string
		actionID    string
		expectError bool
	}{
		{
			name: "Valid spec with default version filter",
			spec: Spec{
				RootDir: "testdata",
			},
			rootDir:     ".",
			scmID:       "test-scm",
			actionID:    "test-action",
			expectError: false,
		},
		{
			name: "Valid spec with custom version filter",
			spec: Spec{
				RootDir: "testdata",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "minor",
				},
			},
			rootDir:     ".",
			scmID:       "test-scm",
			actionID:    "test-action",
			expectError: false,
		},
		{
			name: "Valid spec with absolute rootDir",
			spec: Spec{
				RootDir: "/absolute/path",
			},
			rootDir:     ".",
			scmID:       "test-scm",
			actionID:    "test-action",
			expectError: false,
		},
		{
			name: "Empty rootDir",
			spec: Spec{
				RootDir: "",
			},
			rootDir:     "",
			scmID:       "",
			actionID:    "",
			expectError: true,
		},
		{
			name: "Invalid ignore spec with empty rule",
			spec: Spec{
				RootDir: "testdata",
				Ignore: MatchingRules{
					{},
				},
			},
			rootDir:     ".",
			scmID:       "test-scm",
			actionID:    "test-action",
			expectError: true,
		},
		{
			name: "Invalid only spec with empty rule",
			spec: Spec{
				RootDir: "testdata",
				Only: MatchingRules{
					{},
				},
			},
			rootDir:     ".",
			scmID:       "test-scm",
			actionID:    "test-action",
			expectError: true,
		},
		{
			name: "Valid spec with ignore rule",
			spec: Spec{
				RootDir: "testdata",
				Ignore: MatchingRules{
					{Path: "testdata/project1/MODULE.bazel"},
				},
			},
			rootDir:     ".",
			scmID:       "test-scm",
			actionID:    "test-action",
			expectError: false,
		},
		{
			name: "Valid spec with only rule",
			spec: Spec{
				RootDir: "testdata",
				Only: MatchingRules{
					{Modules: map[string]string{"rules_go": ""}},
				},
			},
			rootDir:     ".",
			scmID:       "test-scm",
			actionID:    "test-action",
			expectError: false,
		},
		{
			name: "Invalid spec with empty module name in ignore rule",
			spec: Spec{
				RootDir: "testdata",
				Ignore: MatchingRules{
					{Modules: map[string]string{"": ""}},
				},
			},
			rootDir:     ".",
			scmID:       "test-scm",
			actionID:    "test-action",
			expectError: true,
		},
		{
			name: "Invalid spec with empty module name in only rule",
			spec: Spec{
				RootDir: "testdata",
				Only: MatchingRules{
					{Modules: map[string]string{"": ""}},
				},
			},
			rootDir:     ".",
			scmID:       "test-scm",
			actionID:    "test-action",
			expectError: true,
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			bazel, err := New(tt.spec, tt.rootDir, tt.scmID, tt.actionID)
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.actionID, bazel.actionID)
			assert.Equal(t, tt.scmID, bazel.scmID)
		})
	}
}

func TestDiscoverManifests(t *testing.T) {
	testdata := []struct {
		name              string
		rootDir           string
		spec              Spec
		scmID             string
		actionID          string
		expectedManifests int
		minManifests      int // Minimum number of manifests expected
	}{
		{
			name:    "Discover all modules",
			rootDir: "testdata",
			spec: Spec{
				RootDir: "",
			},
			scmID:        "",
			actionID:     "",
			minManifests: 4, // At least rules_go, gazelle, protobuf, rules_docker
		},
		{
			name:    "Discover with Only filter",
			rootDir: "testdata",
			spec: Spec{
				RootDir: "",
				Only: MatchingRules{
					MatchingRule{
						Modules: map[string]string{
							"rules_go": "",
						},
					},
				},
			},
			scmID:        "",
			actionID:     "",
			minManifests: 2, // rules_go appears in multiple files
		},
		{
			name:    "Discover with Ignore filter",
			rootDir: "testdata",
			spec: Spec{
				RootDir: "",
				Ignore: MatchingRules{
					MatchingRule{
						Modules: map[string]string{
							"gazelle": "",
						},
					},
				},
			},
			scmID:        "",
			actionID:     "",
			minManifests: 3, // Should exclude gazelle
		},
		{
			name:    "Discover with version filter",
			rootDir: "testdata",
			spec: Spec{
				RootDir: "",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "minor",
				},
			},
			scmID:        "",
			actionID:     "",
			minManifests: 4,
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			absRootDir, err := filepath.Abs(tt.rootDir)
			require.NoError(t, err)

			bazel, err := New(tt.spec, absRootDir, tt.scmID, tt.actionID)
			require.NoError(t, err)

			manifests, err := bazel.DiscoverManifests()
			require.NoError(t, err)
			assert.GreaterOrEqual(t, len(manifests), tt.minManifests)

			// Verify manifest structure
			for _, manifest := range manifests {
				manifestStr := string(manifest)
				assert.Contains(t, manifestStr, "name:")
				assert.Contains(t, manifestStr, "sources:")
				assert.Contains(t, manifestStr, "conditions:")
				assert.Contains(t, manifestStr, "targets:")
				assert.Contains(t, manifestStr, "kind: bazelregistry")
				assert.Contains(t, manifestStr, "kind: bazelmod")
			}
		})
	}
}

func TestGetBazelModuleManifest(t *testing.T) {
	bazel := Bazel{
		actionID: "test-action",
		scmID:    "test-scm",
		versionFilter: version.Filter{
			Kind:    "semver",
			Pattern: "*",
		},
	}

	manifest, err := bazel.getBazelModuleManifest(
		"testdata/project1/MODULE.bazel",
		"rules_go",
		">=0.42.0",
	)

	require.NoError(t, err)
	manifestStr := string(manifest)

	// Verify manifest contains expected elements
	assert.Contains(t, manifestStr, "rules_go")
	assert.Contains(t, manifestStr, "bazelregistry")
	assert.Contains(t, manifestStr, "bazelmod")
	assert.Contains(t, manifestStr, "testdata/project1/MODULE.bazel")
	assert.Contains(t, manifestStr, ">=0.42.0")
	assert.Contains(t, manifestStr, "test-action")
	assert.Contains(t, manifestStr, "test-scm")
}

func TestSanitizeID(t *testing.T) {
	testdata := []struct {
		input    string
		expected string
	}{
		{
			input:    "rules-go",
			expected: "rules_go",
		},
		{
			input:    "rules.go",
			expected: "rules_go",
		},
		{
			input:    "rules/go",
			expected: "rules_go",
		},
		{
			input:    "simple_name",
			expected: "simple_name",
		},
	}

	for _, tt := range testdata {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeID(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDiscoverManifestsWithEmptyFiles(t *testing.T) {
	// Create a temporary directory with an empty MODULE.bazel (no dependencies)
	tmpDir := t.TempDir()
	moduleFile := filepath.Join(tmpDir, "MODULE.bazel")
	content := `module(
    name = "test",
    version = "1.0.0",
)
`
	err := os.WriteFile(moduleFile, []byte(content), 0600)
	require.NoError(t, err)

	bazel := Bazel{
		rootDir: tmpDir,
		spec: Spec{
			RootDir: "",
		},
		versionFilter: version.Filter{
			Kind:    "semver",
			Pattern: "*",
		},
	}

	manifests, err := bazel.DiscoverManifests()
	require.NoError(t, err)
	// Should return empty manifests (slice), not error
	// In Go, a nil slice and empty slice are different, but both are valid
	// We just check that the function succeeds without error
	if manifests == nil {
		manifests = [][]byte{} // Normalize to empty slice
	}
	assert.Len(t, manifests, 0)
}

func TestDiscoverManifestsInvalidFile(t *testing.T) {
	absRootDir, err := filepath.Abs("testdata/invalid")
	require.NoError(t, err)

	bazel := Bazel{
		rootDir: absRootDir,
		spec: Spec{
			RootDir: "",
		},
		versionFilter: version.Filter{
			Kind:    "semver",
			Pattern: "*",
		},
	}

	manifests, err := bazel.DiscoverManifests()
	// Should not error, but parser is lenient and may still extract deps
	require.NoError(t, err)
	// Parser may successfully extract dependencies even from invalid syntax
	// So we just check that it doesn't error
	assert.NotNil(t, manifests)
}

func TestManifestTemplateStructure(t *testing.T) {
	bazel := Bazel{
		actionID: "test-action",
		scmID:    "test-scm",
		versionFilter: version.Filter{
			Kind:    "semver",
			Pattern: "*",
		},
	}

	manifest, err := bazel.getBazelModuleManifest(
		"MODULE.bazel",
		"test_module",
		">=1.0.0",
	)

	require.NoError(t, err)
	manifestStr := string(manifest)

	// Verify YAML structure
	lines := strings.Split(manifestStr, "\n")

	// Check for required sections
	hasSources := false
	hasConditions := false
	hasTargets := false

	for _, line := range lines {
		if strings.TrimSpace(line) == "sources:" {
			hasSources = true
		}
		if strings.TrimSpace(line) == "conditions:" {
			hasConditions = true
		}
		if strings.TrimSpace(line) == "targets:" {
			hasTargets = true
		}
	}

	assert.True(t, hasSources, "Manifest should contain sources section")
	assert.True(t, hasConditions, "Manifest should contain conditions section")
	assert.True(t, hasTargets, "Manifest should contain targets section")
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
			name:    "Simple project1 with default version filter",
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
			name:    "Project1 with scmID and actionID",
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
				// We expect manifest generated by the autodiscovery to use the yaml syntax
				manifests = append(manifests, string(rawManifests[i]))
				assert.Equal(t, tt.expectedManifests[i], manifests[i], "Manifest %d should match expected output", i)
			}
		})
	}
}
