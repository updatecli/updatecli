package pyproject

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeName(t *testing.T) {
	assert.Equal(t, "jinja2", normalizeName("Jinja2"))
	assert.Equal(t, "zope-interface", normalizeName("zope.interface"))
	assert.Equal(t, "my-package", normalizeName("My__Package"))
	assert.Equal(t, "a-b", normalizeName("a-._b"))
}

func TestExactPinnedVersion(t *testing.T) {
	testdata := map[string]string{
		"==1.2.3":       "1.2.3",
		"== 1.2.3":      "1.2.3",
		"==1.0b2":       "1.0b2",
		"==1.*":         "",
		"===1.2.3":      "",
		">=1.2":         "",
		"~=1.2":         "",
		"==1.2,<2":      "",
		"":              "",
		"==1.0.post1":   "1.0.post1",
		"==2.0.0+local": "2.0.0+local",
	}

	for constraint, expected := range testdata {
		assert.Equal(t, expected, exactPinnedVersion(constraint), constraint)
	}
}

func TestLoadUvLockVersions(t *testing.T) {
	lockedPackages, err := loadUvLockVersions("testdata/vulnerability", trustedRegistries(nil))
	require.NoError(t, err)
	locked := lockedPackages.versions
	assert.True(t, lockedPackages.notFromRegistry["internal-utils"])
	assert.True(t, lockedPackages.notFromRegistry["vulnerable-project"])
	assert.False(t, lockedPackages.notFromRegistry["jinja2"])
	// Other registries are not trusted by default
	assert.Empty(t, locked["markupsafe"])
	assert.Equal(t, "https://artifactory.example.com/api/pypi/pypi-remote/simple/", lockedPackages.untrustedRegistry["markupsafe"])
	assert.Equal(t, "https://pypi.internal.example.com/simple", lockedPackages.untrustedRegistry["corp-utils"])

	// A trusted mirror is compared without its trailing "/"
	mirrored, err := loadUvLockVersions("testdata/vulnerability",
		trustedRegistries([]string{"https://artifactory.example.com/api/pypi/pypi-remote/simple"}))
	require.NoError(t, err)
	assert.Equal(t, []string{"2.1.0"}, mirrored.versions["markupsafe"])
	assert.Empty(t, mirrored.untrustedRegistry["markupsafe"])
	assert.NotEmpty(t, mirrored.untrustedRegistry["corp-utils"])

	assert.Equal(t, []string{"3.1.4"}, locked["jinja2"])
	assert.Equal(t, []string{"1.26.4", "2.1.0"}, locked["numpy"])
	assert.Empty(t, locked["requests"])
	// Workspace members are not PyPI releases
	assert.Empty(t, locked["internal-utils"])
	// The project itself has a virtual source
	assert.Empty(t, locked["vulnerable-project"])

	_, err = loadUvLockVersions("testdata/no_lockfile", trustedRegistries(nil))
	assert.Error(t, err)
}
