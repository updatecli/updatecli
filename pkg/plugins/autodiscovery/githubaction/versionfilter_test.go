package githubaction

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// TestVersionFilterOnNonSemverTag checks a versionfilter that can't be applied to the
// current tag, such as "latest", doesn't crash the crawler: no source spec is built
// for such a tag, so the fallback pattern must not be written to it.
func TestVersionFilterOnNonSemverTag(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Dir(filepath.Join(dir, ".github/workflows/ci.yaml")), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".github/workflows/ci.yaml"), []byte("on: push\njobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n      - uses: docker://nginx:latest\n"), 0o600))
	crawler, err := New(Spec{
		VersionFilter: version.Filter{Kind: "semver", Pattern: "minor"},
	}, dir, "", "")
	require.NoError(t, err)

	require.NotPanics(t, func() {
		_, err = crawler.DiscoverManifests()
	})
	require.NoError(t, err)
}
