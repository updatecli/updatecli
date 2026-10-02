package golang

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func TestSearchGoModFiles(t *testing.T) {

	dataset := []struct {
		name               string
		rootDir            string
		expectedFoundFiles []string
	}{
		{
			name:    "Default working scenario",
			rootDir: "testdata",
			expectedFoundFiles: []string{
				"testdata/noModule/go.mod",
				"testdata/noSumFile/go.mod",
				"testdata/pseudoVersion/go.mod",
				"testdata/replace/go.mod",
				"testdata/replaceInactive/go.mod",
			},
		},
	}

	for _, d := range dataset {
		t.Run(d.name, func(t *testing.T) {
			foundFiles, err := searchGoModFiles(d.rootDir)
			require.NoError(t, err)

			assert.Equal(t, foundFiles, d.expectedFoundFiles)
		})
	}
}

func TestGetGoModContent(t *testing.T) {
	dataset := []struct {
		name                   string
		goModFile              string
		expectedModules        map[string]string
		expectedReplaceModules []Replace
		expectedApplied        map[string]Replace
		expectedGoVersion      string
	}{
		{
			name:      "Replace go module",
			goModFile: "testdata/replace/go.mod",
			expectedReplaceModules: []Replace{
				{
					OldPath:    "github.com/rancher/saml",
					OldVersion: "",
					NewPath:    "github.com/rancher/saml",
					NewVersion: "v0.2.0",
				},
				{
					OldPath:    "github.com/crewjam/saml",
					OldVersion: "v0.6.0",
					NewPath:    "github.com/crewjam/saml",
					NewVersion: "v0.5.0",
				},
			},
			expectedApplied: map[string]Replace{
				"github.com/rancher/saml": {
					OldPath:    "github.com/rancher/saml",
					NewPath:    "github.com/rancher/saml",
					NewVersion: "v0.2.0",
				},
				"github.com/crewjam/saml": {
					OldPath:    "github.com/crewjam/saml",
					OldVersion: "v0.6.0",
					NewPath:    "github.com/crewjam/saml",
					NewVersion: "v0.5.0",
				},
				"github.com/stretchr/testify": {
					OldPath: "github.com/stretchr/testify",
					NewPath: "../local/testify",
				},
			},
			expectedModules: map[string]string{
				"github.com/rancher/saml":     "v0.3.0",
				"github.com/crewjam/saml":     "v0.6.0",
				"github.com/stretchr/testify": "v1.8.4",
			},
			expectedGoVersion: "1.25.0",
		},
		{
			name:      "Default go modules",
			goModFile: "testdata/noModule/go.mod",
			expectedModules: map[string]string{
				"gopkg.in/yaml.v3": "v3.0.1",
			},
			expectedGoVersion: "1.20",
		},
	}

	for _, d := range dataset {
		t.Run(d.name, func(t *testing.T) {
			foundGoVersion, foundGoModules, foundReplaceGoModules, foundAppliedReplaces, err := getGoModContent(d.goModFile)
			require.NoError(t, err)

			assert.Equal(t, d.expectedModules, foundGoModules)
			assert.Equal(t, d.expectedReplaceModules, foundReplaceGoModules)
			assert.Equal(t, d.expectedApplied, foundAppliedReplaces)
			assert.Equal(t, d.expectedGoVersion, foundGoVersion)
		})
	}
}

func TestGetGoModContentAppliedReplaces(t *testing.T) {
	goModFile := filepath.Join(t.TempDir(), "go.mod")
	goMod := `module example.com/replaced

go 1.25.0

require (
	github.com/a/unversioned v1.0.0
	github.com/b/matching v1.0.0
	github.com/c/mismatching v1.0.0
	github.com/d/local v1.0.0
	github.com/e/indirect v1.0.0 // indirect
	github.com/f/both v1.0.0
	github.com/g/shadowed v1.0.0
	github.com/i/ambiguous v1.0.0
)

replace (
	github.com/a/unversioned => github.com/a/fork v1.2.0
	github.com/b/matching v1.0.0 => github.com/b/matching v0.9.0
	github.com/c/mismatching v0.1.0 => github.com/c/mismatching v0.2.0
	github.com/d/local => ../local
	github.com/e/indirect => github.com/e/fork v1.0.0
	github.com/f/both => github.com/f/fork v1.0.0
	github.com/f/both v1.0.0 => github.com/f/pinned v2.0.0
	github.com/g/shadowed v1.0.0 => ../shadowed
	github.com/g/shadowed => github.com/g/fork v1.0.0
	github.com/h/unrequired => github.com/h/fork v1.0.0
	github.com/i/ambiguous v0.1.0 => github.com/i/old v0.1.0
	github.com/i/ambiguous => github.com/i/fork v1.0.0
)
`
	require.NoError(t, os.WriteFile(goModFile, []byte(goMod), 0o600))

	goVersion, _, _, appliedReplaces, err := getGoModContent(goModFile)
	require.NoError(t, err)

	assert.Equal(t, "1.25.0", goVersion)

	applied := make(map[string]string, len(appliedReplaces))
	for module, r := range appliedReplaces {
		assert.Equal(t, module, r.OldPath)
		applied[module] = r.OldVersion + "=>" + r.NewPath
	}

	assert.Equal(t, map[string]string{
		// Unversioned replace directive of a direct module
		"github.com/a/unversioned": "=>github.com/a/fork",
		// Replace directive matching the required version
		"github.com/b/matching": "v1.0.0=>github.com/b/matching",
		// Replace directive pointing to a local path
		"github.com/d/local": "=>../local",
		// Replace directive matching the required version takes precedence over the unversioned one, whatever the order
		"github.com/f/both":     "v1.0.0=>github.com/f/pinned",
		"github.com/g/shadowed": "v1.0.0=>../shadowed",
		// Unversioned replace directive applied while another one with a version exists
		"github.com/i/ambiguous": "=>github.com/i/fork",
	}, applied)

	assert.True(t, appliedReplaces["github.com/i/ambiguous"].Ambiguous)
	assert.True(t, appliedReplaces["github.com/d/local"].isLocal())
	for _, module := range []string{"github.com/a/unversioned", "github.com/b/matching", "github.com/d/local", "github.com/f/both", "github.com/g/shadowed"} {
		assert.False(t, appliedReplaces[module].Ambiguous, module)
	}
}

func TestPseudoVersion(t *testing.T) {
	dataset := []struct {
		name           string
		version        string
		expectedResult bool
	}{
		{
			name:           "Valid pseudo-version",
			version:        "v0.0.0-20230215024106-420ad0987b9b",
			expectedResult: true,
		},
		{
			name:           "Invalid pseudo-version",
			version:        "v1.2.3",
			expectedResult: false,
		},
		{
			name:           "Valid pseudo-version with zero patch increment form",
			version:        "v0.0.0-0.20230215024106-420ad0987b9b",
			expectedResult: true,
		},
		{
			name:           "Valid pseudo-version with prerelease form",
			version:        "v0.0.0-beta.0.20230215024106-420ad0987b9b",
			expectedResult: true,
		},
		{
			name:           "Valid pseudo-version with incompatible suffix",
			version:        "v0.0.0-20230215024106-420ad0987b9b+incompatible",
			expectedResult: true,
		},
		{
			name:           "Valid zero patch increment pseudo-version with incompatible suffix",
			version:        "v1.2.4-0.20230215024106-420ad0987b9b+incompatible",
			expectedResult: true,
		},
		{
			name:           "Invalid pseudo-version with short timestamp",
			version:        "v1.2.3-2023021502410-420ad0987b9b",
			expectedResult: false,
		},
		{
			name:           "Invalid pseudo-version with short hash",
			version:        "v1.2.3-20230215024106-420ad0987b9",
			expectedResult: false,
		},
	}

	for _, d := range dataset {
		t.Run(d.name, func(t *testing.T) {
			result := isPseudoVersion(d.version)
			assert.Equal(t, d.expectedResult, result)
		})
	}
}

func TestModuleVersionPattern(t *testing.T) {
	const (
		zeroPseudoVersion = "v0.0.0-20230215024106-420ad0987b9b"
		pseudoVersion     = "v1.2.4-0.20230215024106-420ad0987b9b"
	)

	tests := []struct {
		name            string
		filter          version.Filter
		version         string
		expectedKind    string
		expectedPattern string
	}{
		{
			name:            "release with patch pattern",
			filter:          version.Filter{Kind: "semver", Pattern: "patch"},
			version:         "v1.2.3",
			expectedKind:    "semver",
			expectedPattern: "1.2.x",
		},
		{
			name:            "zero pseudo version with default pattern",
			filter:          version.Filter{Kind: "semver", Pattern: "*"},
			version:         zeroPseudoVersion,
			expectedKind:    "semver",
			expectedPattern: ">=0.0.0-20230215024106-420ad0987b9b",
		},
		{
			name:            "pseudo version with patch pattern",
			filter:          version.Filter{Kind: "semver", Pattern: "patch"},
			version:         pseudoVersion,
			expectedKind:    "semver",
			expectedPattern: ">=1.2.4-0.20230215024106-420ad0987b9b, 1.2.x-0",
		},
		{
			name:            "pseudo version with minor pattern",
			filter:          version.Filter{Kind: "semver", Pattern: "minor"},
			version:         pseudoVersion,
			expectedKind:    "semver",
			expectedPattern: ">=1.2.4-0.20230215024106-420ad0987b9b, 1.x.x-0",
		},
		{
			name:            "pseudo version with major pattern",
			filter:          version.Filter{Kind: "semver", Pattern: "major"},
			version:         pseudoVersion,
			expectedKind:    "semver",
			expectedPattern: ">=1.2.4-0.20230215024106-420ad0987b9b, >=1.x.x-0",
		},
		{
			name:            "pseudo version with minoronly pattern",
			filter:          version.Filter{Kind: "semver", Pattern: "minoronly"},
			version:         pseudoVersion,
			expectedKind:    "semver",
			expectedPattern: pseudoVersion + " || >=1.2.4-0.20230215024106-420ad0987b9b, >1.2.x-0 < 2",
		},
		{
			name:            "pseudo version with majoronly pattern",
			filter:          version.Filter{Kind: "semver", Pattern: "majoronly"},
			version:         pseudoVersion,
			expectedKind:    "semver",
			expectedPattern: ">=1.2.4-0.20230215024106-420ad0987b9b, >1",
		},
		{
			name:            "zero pseudo version with majoronly pattern",
			filter:          version.Filter{Kind: "semver", Pattern: "majoronly"},
			version:         zeroPseudoVersion,
			expectedKind:    "semver",
			expectedPattern: ">=0.0.0-20230215024106-420ad0987b9b, >0",
		},
		{
			name:            "pseudo version with a custom constraint",
			filter:          version.Filter{Kind: "semver", Pattern: "~1.2"},
			version:         pseudoVersion,
			expectedKind:    "semver",
			expectedPattern: ">=1.2.4-0.20230215024106-420ad0987b9b, ~1.2",
		},
		{
			name:            "pseudo version with a custom constraint made of several groups",
			filter:          version.Filter{Kind: "semver", Pattern: "~1.2 || ~1.3"},
			version:         pseudoVersion,
			expectedKind:    "semver",
			expectedPattern: ">=1.2.4-0.20230215024106-420ad0987b9b, ~1.2 || >=1.2.4-0.20230215024106-420ad0987b9b, ~1.3",
		},
		{
			name:            "pseudo version with latest kind",
			filter:          version.Filter{Kind: "latest", Pattern: "latest"},
			version:         pseudoVersion,
			expectedKind:    "latest",
			expectedPattern: "latest",
		},
		{
			name:            "pseudo version with regex kind",
			filter:          version.Filter{Kind: "regex", Pattern: `^v1\.`},
			version:         pseudoVersion,
			expectedKind:    "regex",
			expectedPattern: `^v1\.`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := Golang{versionFilter: tt.filter}

			gotKind, gotPattern, err := g.moduleVersionPattern(tt.version)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedKind, gotKind)
			assert.Equal(t, tt.expectedPattern, gotPattern)
		})
	}
}
