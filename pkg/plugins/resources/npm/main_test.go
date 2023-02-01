package npm

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseNpmRc(t *testing.T) {
	dir, err := CreateDummyRc()
	if err != nil {
		require.NoError(t, err)
	}
	defer os.RemoveAll(dir)
	t.Run("Test parsing custom npmrc", func(t *testing.T) {
		cfg, err := getNpmrcConfig(filepath.Join(dir, ".npmrc"), "", "")
		require.NoError(t, err)
		assert.Contains(t, cfg.Registries, "default")
		assert.Equal(t, cfg.Registries["default"].AuthToken, "")
		assert.Equal(t, cfg.Registries["default"].Url, "https://registry.npmjs.org/")
		assert.Contains(t, cfg.Registries, "mycustomregistry.updatecli.io/")
		assert.Equal(t, cfg.Registries["mycustomregistry.updatecli.io/"].AuthToken, "mytoken")
		assert.Equal(t, cfg.Registries["mycustomregistry.updatecli.io/"].Url, "https://mycustomregistry.updatecli.io/")
		assert.Contains(t, cfg.Scopes, "@TestScope")
		assert.Equal(t, cfg.Scopes["@TestScope"], "mycustomregistry.updatecli.io/")
	})
	t.Run("Test parsing custom npmrc with auth token", func(t *testing.T) {
		cfg, err := getNpmrcConfig("", "https://mycustomregistry.updatecli.io/", "mytoken")
		require.NoError(t, err)
		assert.Contains(t, cfg.Registries, "default")
		assert.Equal(t, cfg.Registries["default"].AuthToken, "mytoken")
		assert.Equal(t, cfg.Registries["default"].Url, "https://mycustomregistry.updatecli.io/")
	})
}
