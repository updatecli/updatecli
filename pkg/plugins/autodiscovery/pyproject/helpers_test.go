package pyproject

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// sourcePackageRegex captures the package name of the source of a generated manifest.
	sourcePackageRegex = regexp.MustCompile(`(?m)^    spec:\n      name: '([^']+)'$`)
	// sourcePatternRegex captures the versionfilter pattern of the source of a generated manifest.
	sourcePatternRegex = regexp.MustCompile(`(?m)^        pattern: '([^']*)'$`)
)

// manifestsByName maps the generated manifests to their pipeline name, the first line without
// its "name: " key. It fails on two manifests sharing a name, which would open conflicting pull
// requests.
func manifestsByName(t *testing.T, manifests [][]byte) map[string]string {
	t.Helper()

	byName := map[string]string{}
	for _, manifest := range manifests {
		name := strings.TrimPrefix(strings.SplitN(string(manifest), "\n", 2)[0], "name: ")
		require.NotContains(t, byName, name, "two manifests are named %s", name)
		byName[name] = string(manifest)
	}
	return byName
}

// assertSourcePatterns checks that the generated manifests are exactly one per dependency of
// expected, each with its source versionfilter pattern.
func assertSourcePatterns(t *testing.T, manifests [][]byte, expected map[string]string) {
	t.Helper()

	got := map[string]string{}
	for _, manifest := range manifests {
		dependency := sourcePackageRegex.FindStringSubmatch(string(manifest))
		require.NotNil(t, dependency, "no source package in manifest:\n%s", manifest)
		require.NotContains(t, got, dependency[1], "two manifests update %q", dependency[1])

		pattern := sourcePatternRegex.FindStringSubmatch(string(manifest))
		require.NotNil(t, pattern, "no versionfilter pattern in manifest:\n%s", manifest)
		got[dependency[1]] = pattern[1]
	}

	assert.Equal(t, expected, got)
}
