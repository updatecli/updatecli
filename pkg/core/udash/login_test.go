package udash

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// clearUdashEnv unsets the Udash environment variables so a developer's own
// configuration cannot leak into the tests.
func clearUdashEnv(t *testing.T) {
	t.Helper()
	t.Setenv(DefaultEnvVariableURL, "")
	t.Setenv(DefaultEnvVariableAPIURL, "")
	t.Setenv(DefaultEnvVariableAccessToken, "")
}

// whoamiServer serves a Udash API at the server root: GET /whoami answers with the
// given status and body, recording the token it was called with, and GET
// /pipeline/scms answers the API base probe.
func whoamiServer(t *testing.T, status int, body string) (*httptest.Server, *string) {
	t.Helper()

	seen := new(string)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/pipeline/scms":
			_, _ = w.Write([]byte(`{"scms":[]}`))
		case "/whoami":
			*seen = r.Header.Get("Authorization")

			w.WriteHeader(status)
			if body != "" {
				_, _ = w.Write([]byte(body))
			}
		default:
			t.Errorf("unexpected request to %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)

	return server, seen
}

func TestLoginWithToken(t *testing.T) {
	useTempConfigDir(t)
	clearUdashEnv(t)

	server, seen := whoamiServer(t, http.StatusOK, `{"subject":"user-1","name":"Jane","permission":"publisher"}`)

	require.NoError(t, Login(server.URL, server.URL, "udash_pat_valid"))

	assert.Equal(t, "Bearer udash_pat_valid", *seen)

	_, _, token, err := getConfigFromFile("")
	require.NoError(t, err)
	assert.Equal(t, "udash_pat_valid", token)
}

func TestLoginRejectedTokenIsNotPersisted(t *testing.T) {
	useTempConfigDir(t)
	clearUdashEnv(t)

	server, _ := whoamiServer(t, http.StatusUnauthorized, "")

	err := Login(server.URL, server.URL, "udash_pat_bogus")
	require.Error(t, err)

	// Nothing must have been written: a bad token has to fail at login time
	// rather than silently later on, when a pipeline tries to publish.
	url, apiURL, token, err := getConfigFromFile("")
	require.Error(t, err)
	assert.Empty(t, url)
	assert.Empty(t, apiURL)
	assert.Empty(t, token)
}

func TestLoginWithTokenAgainstServiceWithoutAuthentication(t *testing.T) {
	useTempConfigDir(t)
	clearUdashEnv(t)

	// /whoami only exists when a mode is configured, so the token has no use there.
	server, _ := whoamiServer(t, http.StatusNotFound, "")

	require.NoError(t, Login(server.URL, server.URL, "udash_pat_valid"))

	_, api, token, err := getConfigFromFile("")
	require.NoError(t, err)
	assert.Equal(t, server.URL, api)
	assert.Empty(t, token, "a token is not stored for a service without authentication")
}

func TestLoginRejectsWrongAPIBase(t *testing.T) {
	// The unversioned base answers 404 on every route, which must not pass for a
	// service without authentication.
	for _, token := range []string{"", "udash_pat_valid"} {
		t.Run("token "+token, func(t *testing.T) {
			useTempConfigDir(t)
			clearUdashEnv(t)

			server := httptest.NewServer(http.NotFoundHandler())
			t.Cleanup(server.Close)

			err := Login(server.URL, server.URL+"/api", token)
			require.ErrorIs(t, err, ErrAPINotFound)
			assert.ErrorContains(t, err, DefaultAPIPath)

			_, _, _, err = getConfigFromFile("")
			require.Error(t, err, "nothing must be stored for a wrong API base")
		})
	}
}

func TestLoginProbeAcceptsPrivateService(t *testing.T) {
	useTempConfigDir(t)
	clearUdashEnv(t)

	// A private service refuses the anonymous probe: the API is still there.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/whoami" && r.Header.Get("Authorization") == "Bearer udash_pat_valid" {
			_, _ = w.Write([]byte(`{"subject":"user-1"}`))
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
	}))
	t.Cleanup(server.Close)

	require.NoError(t, Login(server.URL, server.URL, "udash_pat_valid"))

	_, _, token, err := getConfigFromFile("")
	require.NoError(t, err)
	assert.Equal(t, "udash_pat_valid", token)
}

func TestLoginReadsTokenFromStdin(t *testing.T) {
	useTempConfigDir(t)
	clearUdashEnv(t)

	server, seen := whoamiServer(t, http.StatusOK, `{"subject":"user-1"}`)

	// Support `echo $TOKEN | updatecli udash login <url>`.
	read, write, err := os.Pipe()
	require.NoError(t, err)

	previous := os.Stdin
	os.Stdin = read
	t.Cleanup(func() { os.Stdin = previous })

	_, err = write.WriteString("udash_pat_piped\n")
	require.NoError(t, err)
	require.NoError(t, write.Close())

	require.NoError(t, Login(server.URL, server.URL, ""))

	assert.Equal(t, "Bearer udash_pat_piped", *seen, "trailing newline must be trimmed")
}

func TestLoginAgainstServiceWithoutAuthentication(t *testing.T) {
	useTempConfigDir(t)
	clearUdashEnv(t)

	// /whoami only exists when a mode is configured, so a 404 means the service is
	// open. This is the default deployment and must be registrable with no token
	// and no prompt.
	server, seen := whoamiServer(t, http.StatusNotFound, "")

	require.NoError(t, Login(server.URL, server.URL, ""))
	assert.Empty(t, *seen)

	url, api, token, err := getConfigFromFile("")
	require.NoError(t, err)
	assert.Equal(t, server.URL, url)
	assert.Equal(t, server.URL, api)
	assert.Empty(t, token, "there is no token to store against an open service")
}

func TestLoginRequiresURL(t *testing.T) {
	useTempConfigDir(t)
	clearUdashEnv(t)

	require.ErrorContains(t, Login("", "", "udash_pat_valid"), "service URL is required")
}

func TestLoginDefaultsAPIEndpoint(t *testing.T) {
	useTempConfigDir(t)
	clearUdashEnv(t)

	var paths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		_, _ = w.Write([]byte(`{"subject":"user-1"}`))
	}))
	t.Cleanup(server.Close)

	require.NoError(t, Login(server.URL, "", "udash_pat_valid"))

	assert.Equal(t, []string{"/api/v1beta1/pipeline/scms", "/api/v1beta1/whoami"}, paths,
		"the API endpoint defaults to <url>/api/v1beta1")
}

func TestLoginTokenFromEnv(t *testing.T) {
	useTempConfigDir(t)
	clearUdashEnv(t)

	server, seen := whoamiServer(t, http.StatusOK, `{"subject":"user-1"}`)
	t.Setenv(DefaultEnvVariableAccessToken, "udash_pat_from_env")

	require.NoError(t, Login(server.URL, server.URL, ""))

	assert.Equal(t, "Bearer udash_pat_from_env", *seen)
}
