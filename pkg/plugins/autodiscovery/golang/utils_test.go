package golang

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			foundGoVersion, foundGoModules, foundReplaceGoModules, err := getGoModContent(d.goModFile)
			require.NoError(t, err)

			assert.Equal(t, d.expectedModules, foundGoModules)
			assert.Equal(t, d.expectedReplaceModules, foundReplaceGoModules)
			assert.Equal(t, d.expectedGoVersion, foundGoVersion)
		})
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
