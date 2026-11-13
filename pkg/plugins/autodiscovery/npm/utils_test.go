package npm

import (
	"strings"
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
			rootDir: "testdata",
			expectedFoundFiles: []string{
				"testdata/nolockfile/package.json",
				"testdata/npmlockfile/package.json",
				"testdata/npmworkspace/package.json",
				"testdata/npmworkspace/packages/app/package.json",
				"testdata/pnpmlockfile/package.json",
				"testdata/yarnlockfile/package.json",
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

func TestConvertVersionConstraintToVersion(t *testing.T) {
	dataset := []struct {
		version          string
		expectedResult   string
		expectedError    bool
		expectedErrorMsg string
	}{
		{expectedResult: "1.0.0", version: "1.0.0"},
		{expectedResult: "1.0.0", version: ">=1.0.0"},
		{expectedResult: "1.0.0", version: "1.0.0-alpha"},
		{expectedResult: "1.0.0", version: "1.0.0+alpha"},
		{expectedResult: "", version: "1.0.0_alpha", expectedError: true, expectedErrorMsg: "parsing version constraint \"1.0.0_alpha\": improper constraint: \"1.0.0_alpha\""},
		{expectedResult: "1.0.0", version: "1.0"},
		{expectedResult: "1.0.0", version: "1"},
		{expectedResult: "1.0.0", version: "~1.0"},
		{expectedResult: "1.0.0", version: "1.x"},
		{expectedResult: "1.0.0", version: ">1.0.0"},
		{expectedResult: "1.0.0", version: ">=1.0.0"},
		{expectedResult: "1.0.0", version: "<1.0.0"},
		{expectedResult: "1.0.0", version: "<=1.0.0"},
		{expectedResult: "", version: "file://../dyl", expectedError: true, expectedErrorMsg: "parsing version constraint \"file://../dyl\": improper constraint: \"file://../dyl\""},
		{expectedResult: "1.0.0", version: "<1.0.0 || >= 2.3.1 < 2.4.5 || >=2.5.2 < 3.0.0"},
		{expectedResult: "", version: "latest"},
	}

	for _, d := range dataset {
		t.Run(d.version, func(t *testing.T) {
			gotResult, err := convertSemverVersionConstraintToVersion(d.version)
			if d.expectedError {
				require.Error(t, err)
				if d.expectedErrorMsg != "" {
					assert.Equal(t, d.expectedErrorMsg, err.Error())
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, d.expectedResult, gotResult)
		})
	}
}

func TestGetTargetCommandExactVersion(t *testing.T) {
	// The yarn command depends on the installed npm version, so only the exact flag is checked for it
	testdata := []struct {
		name          string
		cmd           string
		exact         bool
		expectedFlags []string
		absentFlags   []string
	}{
		{name: "npm exact", cmd: npmIdentifier, exact: true, expectedFlags: []string{"--save-exact"}},
		{name: "npm range", cmd: npmIdentifier, exact: false, absentFlags: []string{"--save-exact"}},
		{name: "pnpm exact", cmd: "pnpm", exact: true, expectedFlags: []string{"--save-exact"}},
		{name: "pnpm range", cmd: "pnpm", exact: false, absentFlags: []string{"--save-exact"}},
		{name: "yarn range", cmd: "yarn", exact: false, absentFlags: []string{"--save-exact", "--exact"}},
	}

	for _, d := range testdata {
		t.Run(d.name, func(t *testing.T) {
			got := getTargetCommand(d.cmd, "axios", d.exact)
			for _, flag := range d.expectedFlags {
				assert.Contains(t, got, " "+flag+" ")
			}
			for _, flag := range d.absentFlags {
				assert.NotContains(t, got, flag)
			}
		})
	}

	yarnCommand := getTargetCommand("yarn", "axios", true)
	assert.True(t,
		strings.Contains(yarnCommand, " --save-exact ") || strings.Contains(yarnCommand, " --exact "),
		"yarn command %q must keep an exact version exact", yarnCommand)
}
