package app

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSpecFromEnv(t *testing.T) {
	os.Setenv("UPDATECLI_GITHUB_APP_CLIENT_ID", "123456")
	os.Setenv("UPDATECLI_GITHUB_APP_PRIVATE_KEY", "dummykey")
	os.Setenv("UPDATECLI_GITHUB_APP_INSTALLATION_ID", "789012")
	os.Setenv("UPDATECLI_GITHUB_APP_EXPIRATION_TIME", "3600")

	spec := NewSpecFromEnv()
	require.NotNil(t, spec)
	assert.Equal(t, "123456", spec.ClientID)
	assert.Equal(t, "dummykey", spec.PrivateKey)
	assert.Equal(t, "789012", spec.InstallationID)
	assert.Equal(t, "3600", spec.ExpirationTime)
}

func TestValidate(t *testing.T) {
	spec := &Spec{
		ClientID:       "123456",
		PrivateKey:     "dummykey",
		InstallationID: "789012",
		ExpirationTime: "3600",
	}
	assert.NoError(t, spec.Validate())

	invalidSpec := &Spec{}
	assert.Error(t, invalidSpec.Validate())
}

func TestGetPrivateKey(t *testing.T) {
	spec := Spec{PrivateKey: "dummykey"}
	key, err := spec.getPrivateKey()
	assert.NoError(t, err)
	assert.Equal(t, "dummykey", key)

	tmpFile, err := os.CreateTemp("", "privatekey")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	_, err = tmpFile.WriteString("filekey")
	require.NoError(t, err)
	tmpFile.Close()

	spec = Spec{PrivateKeyPath: tmpFile.Name()}
	key, err = spec.getPrivateKey()
	assert.NoError(t, err)
	assert.Equal(t, "filekey", key)

	spec = Spec{}
	_, err = spec.getPrivateKey()
	assert.Error(t, err)
}

func TestGetInstallationID(t *testing.T) {
	spec := Spec{InstallationID: "789012"}
	id, err := spec.getInstallationID()
	assert.NoError(t, err)
	assert.Equal(t, int64(789012), id)

	spec = Spec{InstallationID: "notanumber"}
	_, err = spec.getInstallationID()
	assert.Error(t, err)
}

func TestGetExpirationTime(t *testing.T) {
	spec := Spec{ExpirationTime: ""}
	exp, err := spec.getExpirationTime()
	assert.NoError(t, err)
	assert.Equal(t, DefaultExpirationTime, exp)

	spec = Spec{ExpirationTime: "7200"}
	exp, err = spec.getExpirationTime()
	assert.NoError(t, err)
	assert.Equal(t, int64(7200), exp)

	spec = Spec{ExpirationTime: "notanumber"}
	_, err = spec.getExpirationTime()
	assert.Error(t, err)
}

func TestGetExpirationTimeDuration(t *testing.T) {
	spec := Spec{ExpirationTime: "3600"}
	dur, err := spec.getExpirationTimeDuration()
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(3600)*time.Second, dur)
}

func TestGetoauth2TokenSource_Invalid(t *testing.T) {
	spec := &Spec{}
	_, err := spec.Getoauth2TokenSource()
	assert.Error(t, err)

	spec = &Spec{
		ClientID:       "123456",
		PrivateKey:     "dummykey",
		InstallationID: "notanumber",
		ExpirationTime: "3600",
	}
	_, err = spec.Getoauth2TokenSource()
	assert.Error(t, err)

	spec = &Spec{
		ClientID:       "123456",
		PrivateKey:     "dummykey",
		InstallationID: "789012",
		ExpirationTime: "notanumber",
	}
	_, err = spec.Getoauth2TokenSource()
	assert.Error(t, err)
}
