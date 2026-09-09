package udash

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogoutRemovesTheOnlyCredential(t *testing.T) {
	useTempConfigDir(t)

	require.NoError(t, updateConfigFile(authData{
		URL:   "https://app.updatecli.io",
		API:   "https://app.updatecli.io/api",
		Token: "udash_pat_only",
	}))

	require.NoError(t, Logout("https://app.updatecli.io"))

	// Regression: the write used to be skipped when the last entry was removed,
	// so logging out of the only configured service never persisted.
	got, err := readConfigFile()
	require.NoError(t, err)

	assert.Empty(t, got.Auths)
	assert.Empty(t, got.Default)
}

func TestLogoutKeepsOtherCredentials(t *testing.T) {
	useTempConfigDir(t)

	require.NoError(t, updateConfigFile(authData{
		URL:   "https://app.updatecli.io",
		API:   "https://app.updatecli.io/api",
		Token: "udash_pat_first",
	}))
	require.NoError(t, updateConfigFile(authData{
		URL:   "https://udash.example.com",
		API:   "https://udash.example.com/api",
		Token: "udash_pat_second",
	}))

	require.NoError(t, Logout("https://udash.example.com"))

	got, err := readConfigFile()
	require.NoError(t, err)

	assert.Len(t, got.Auths, 1)
	assert.Equal(t, "udash_pat_first", got.Auths["app.updatecli.io/api"].Token)
	assert.Equal(t, "app.updatecli.io/api", got.Default, "a remaining credential becomes the default")
}

func TestLogoutByAPIURL(t *testing.T) {
	useTempConfigDir(t)

	require.NoError(t, updateConfigFile(authData{
		URL:   "https://app.updatecli.io",
		API:   "https://app.updatecli.io/api",
		Token: "udash_pat_only",
	}))

	require.NoError(t, Logout("https://app.updatecli.io/api"))

	got, err := readConfigFile()
	require.NoError(t, err)
	assert.Empty(t, got.Auths)
}
