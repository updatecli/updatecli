package nomad

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
	require.NoError(t, os.MkdirAll(filepath.Dir(filepath.Join(dir, "nomad.hcl")), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "nomad.hcl"), []byte("job \"web\" {\n  group \"web\" {\n    task \"nginx\" {\n      driver = \"docker\"\n      config {\n        image = \"nginx:latest\"\n      }\n    }\n  }\n}\n"), 0o600))
	crawler, err := New(Spec{
		VersionFilter: version.Filter{Kind: "semver", Pattern: "minor"},
	}, dir, "", "")
	require.NoError(t, err)

	require.NotPanics(t, func() {
		_, err = crawler.DiscoverManifests()
	})
	require.NoError(t, err)
}
