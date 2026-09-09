package udash

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// useTempConfigDir points os.UserConfigDir at a temporary directory so tests never
// touch the developer's real Updatecli configuration.
func useTempConfigDir(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	// os.UserConfigDir reads XDG_CONFIG_HOME on unix and HOME on darwin.
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("HOME", dir)
	t.Setenv("AppData", dir)

	// APIURLSelector is a package level global, reset it between tests.
	previous := APIURLSelector
	APIURLSelector = ""
	t.Cleanup(func() { APIURLSelector = previous })

	return dir
}

func TestUpdateConfigFileRoundTrip(t *testing.T) {
	useTempConfigDir(t)

	require.NoError(t, updateConfigFile(authData{
		URL:   "https://app.updatecli.io",
		API:   "https://app.updatecli.io/api",
		Token: "udash_pat_first",
	}))

	got, err := readConfigFile()
	require.NoError(t, err)

	assert.Equal(t, "app.updatecli.io/api", got.Default)
	assert.Equal(t, "udash_pat_first", got.Auths["app.updatecli.io/api"].Token)

	// A second service must not evict the first one.
	require.NoError(t, updateConfigFile(authData{
		URL:   "https://udash.example.com",
		API:   "https://udash.example.com/api",
		Token: "udash_pat_second",
	}))

	got, err = readConfigFile()
	require.NoError(t, err)

	assert.Len(t, got.Auths, 2)
	assert.Equal(t, "udash.example.com/api", got.Default, "the most recent login becomes the default")
	assert.Equal(t, "udash_pat_first", got.Auths["app.updatecli.io/api"].Token)
}

func TestWriteConfigFileIsNotWorldReadable(t *testing.T) {
	dir := useTempConfigDir(t)

	require.NoError(t, updateConfigFile(authData{
		URL:   "https://app.updatecli.io",
		API:   "https://app.updatecli.io/api",
		Token: "udash_pat_secret",
	}))

	info, err := os.Stat(filepath.Join(dir, "updatecli", "udash.json"))
	require.NoError(t, err)

	// The file holds a long lived API token.
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
}

func TestGetConfigFromFile(t *testing.T) {
	useTempConfigDir(t)

	require.NoError(t, updateConfigFile(authData{
		URL:   "https://app.updatecli.io",
		API:   "https://app.updatecli.io/api",
		Token: "udash_pat_default",
	}))
	require.NoError(t, updateConfigFile(authData{
		URL:   "https://udash.example.com",
		API:   "https://udash.example.com/api",
		Token: "udash_pat_other",
	}))

	t.Run("empty selector returns the default", func(t *testing.T) {
		_, _, token, err := getConfigFromFile("")
		require.NoError(t, err)
		assert.Equal(t, "udash_pat_other", token)
	})

	t.Run("explicit API URL selects that credential", func(t *testing.T) {
		_, _, token, err := getConfigFromFile("https://app.updatecli.io/api")
		require.NoError(t, err)
		assert.Equal(t, "udash_pat_default", token)
	})

	t.Run("APIURLSelector overrides the argument", func(t *testing.T) {
		APIURLSelector = "https://app.updatecli.io/api"
		defer func() { APIURLSelector = "" }()

		_, _, token, err := getConfigFromFile("")
		require.NoError(t, err)
		assert.Equal(t, "udash_pat_default", token)
	})

	t.Run("unknown API URL errors", func(t *testing.T) {
		_, _, _, err := getConfigFromFile("https://nope.example.com/api")
		require.Error(t, err)
	})
}
