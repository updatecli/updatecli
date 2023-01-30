package npm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchPackageJsonFiles(t *testing.T) {

	dataset := []struct {
		name               string
		rootDir            string
		expectedFoundFiles []string
	}{
		{
			name:    "Default working scenario",
			rootDir: "test/testdata",
			expectedFoundFiles: []string{
				"test/testdata/package.json",
			},
		},
	}

	for _, d := range dataset {
		t.Run(d.name, func(t *testing.T) {
			foundFiles, err := searchPackageJsonFiles(d.rootDir)
			require.NoError(t, err)

			assert.Equal(t, foundFiles, d.expectedFoundFiles)
		})
	}
}

func TestIsVersionConstraintSpecified(t *testing.T) {

	dataset := []struct {
		version        string
		strictSemver   bool
		expectedResult bool
	}{
		{expectedResult: false, version: "1.0.0"},
		{expectedResult: false, version: "1.0.0-alpha"},
		{expectedResult: false, version: "1.0.0+alpha"},
		{expectedResult: true, version: "1.0.0_alpha"},
		{expectedResult: true, version: "1.0"},
		{expectedResult: true, version: "1.0"},
		{expectedResult: true, version: "1"},
		{expectedResult: true, version: "~1.0"},
		{expectedResult: true, version: "1.x"},
		{expectedResult: true, version: ">1.0.0"},
		{expectedResult: true, version: ">=1.0.0"},
		{expectedResult: true, version: "<1.0.0"},
		{expectedResult: true, version: "<=1.0.0"},
		{expectedResult: true, version: "<=1.0.0"},
		{expectedResult: true, version: "file://../dyl"},
		{expectedResult: true, version: "<1.0.0 || >= 2.3.1 < 2.4.5 || >=2.5.2 < 3.0.0"},
		{expectedResult: true, version: "latest"},
	}

	for _, d := range dataset {
		t.Run(d.version, func(t *testing.T) {
			gotResult := isVersionConstraintSpecified("foo", d.version)
			assert.Equal(t, gotResult, d.expectedResult)
		})
	}
}

func TestIsVersionConstraintSupported(t *testing.T) {

	dataset := []struct {
		version        string
		strictSemver   bool
		expectedResult bool
	}{
		{expectedResult: true, version: "1.0.0"},
		{expectedResult: true, version: "1.0.0-alpha"},
		{expectedResult: true, version: "1.0.0+alpha"},
		{expectedResult: false, version: "1.0.0_alpha"},
		{expectedResult: true, version: "1.0"},
		{expectedResult: true, version: "1.0"},
		{expectedResult: true, version: "1"},
		{expectedResult: true, version: "~1.0"},
		{expectedResult: true, version: "1.x"},
		{expectedResult: true, version: ">1.0.0"},
		{expectedResult: true, version: ">=1.0.0"},
		{expectedResult: true, version: "<1.0.0"},
		{expectedResult: true, version: "<=1.0.0"},
		{expectedResult: true, version: "<=1.0.0"},
		{expectedResult: false, version: "file://../dyl"},
		{expectedResult: false, version: "https://../dyl"},
		{expectedResult: true, version: "<1.0.0 || >= 2.3.1 < 2.4.5 || >=2.5.2 < 3.0.0"},
		{expectedResult: true, version: "latest"},
	}

	for _, d := range dataset {
		t.Run(d.version, func(t *testing.T) {
			gotResult := isVersionConstraintSupported("foo", d.version)
			assert.Equal(t, gotResult, d.expectedResult)
		})
	}
}
