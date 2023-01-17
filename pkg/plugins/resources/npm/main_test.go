package npm

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func CreateDummyRc() (string, error) {
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		return "", err
	}
	config, err := os.Create(filepath.Join(dir, ".npmrc"))
	if err != nil {
		return "", err
	}
	defer config.Close()
	_, err = fmt.Fprintf(config, "//npm.pkg.github.com/:_authToken=1234567890\n")
	if err != nil {
		return "", err
	}
	_, err = fmt.Fprintf(config, "@TestScope:registry=https://npm.pkg.github.com/\n")
	if err != nil {
		return "", err
	}
	return dir, nil
}

func TestParseNpmRc(t *testing.T) {
	dir, err := CreateDummyRc()
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)
	t.Run("Test parsing custom npmrc", func(t *testing.T) {
		cfg, err := getNpmrcConfig(filepath.Join(dir, ".npmrc"), "", "")
		require.NoError(t, err)
		assert.Contains(t, cfg.Registries, "default")
		assert.Equal(t, cfg.Registries["default"].AuthToken, "")
		assert.Equal(t, cfg.Registries["default"].Url, "https://registry.npmjs.org/")
		assert.Contains(t, cfg.Registries, "npm.pkg.github.com/")
		assert.Equal(t, cfg.Registries["npm.pkg.github.com/"].AuthToken, "1234567890")
		assert.Equal(t, cfg.Registries["npm.pkg.github.com/"].Url, "https://npm.pkg.github.com/")
		assert.Contains(t, cfg.Scopes, "@TestScope")
		assert.Equal(t, cfg.Scopes["@TestScope"], "npm.pkg.github.com/")
	})
	t.Run("Test parsing custom npmrc with auth token", func(t *testing.T) {
		cfg, err := getNpmrcConfig("", "https://npm.pkg.github.com/", "1234567890")
		require.NoError(t, err)
		assert.Contains(t, cfg.Registries, "default")
		assert.Equal(t, cfg.Registries["default"].AuthToken, "1234567890")
		assert.Equal(t, cfg.Registries["default"].Url, "https://npm.pkg.github.com/")
	})
}
