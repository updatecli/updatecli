package udash

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/updatecli/updatecli/pkg/core/reports"
)

// publishRequest records what a fake Udash received.
type publishRequest struct {
	path          string
	authorization string
	body          map[string]any
}

// publishServer serves POST /pipeline/reports with the given answer.
func publishServer(t *testing.T, status int, body string) (*httptest.Server, *publishRequest) {
	t.Helper()

	seen := &publishRequest{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.path = r.URL.Path
		seen.authorization = r.Header.Get("Authorization")

		data, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.NoError(t, json.Unmarshal(data, &seen.body))

		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)

	return server, seen
}

func TestPublishSuccess(t *testing.T) {
	useTempConfigDir(t)
	clearUdashEnv(t)

	server, seen := publishServer(t, http.StatusCreated, `{"ReportID":"abc"}`)
	t.Setenv(DefaultEnvVariableAPIURL, server.URL+"/api")
	t.Setenv(DefaultEnvVariableURL, "https://udash.example.com")
	t.Setenv(DefaultEnvVariableAccessToken, "udash_pat_valid")

	r := reports.Report{Name: "pipeline"}
	require.NoError(t, Publish(&r))

	assert.Equal(t, "/api/pipeline/reports", seen.path)
	assert.Equal(t, "Bearer udash_pat_valid", seen.authorization)
	assert.Equal(t, "https://udash.example.com/pipeline/reports/abc", r.ReportURL)
}

func TestPublishWithoutTokenSendsNoAuthorization(t *testing.T) {
	useTempConfigDir(t)
	clearUdashEnv(t)

	server, seen := publishServer(t, http.StatusCreated, `{"ReportID":"abc"}`)
	t.Setenv(DefaultEnvVariableAPIURL, server.URL)

	r := reports.Report{Name: "pipeline"}
	require.NoError(t, Publish(&r))

	assert.Empty(t, seen.authorization)
	assert.Empty(t, r.ReportURL)
}

func TestPublishErrors(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		body     string
		contains []string
	}{
		{
			name:     "unauthorized",
			status:   http.StatusUnauthorized,
			body:     "invalid token",
			contains: []string{"401", "invalid token", "updatecli udash login https://udash.example.com"},
		},
		{
			name:     "server error",
			status:   http.StatusInternalServerError,
			body:     "boom",
			contains: []string{"500", "boom"},
		},
		{
			name:     "malformed response",
			status:   http.StatusCreated,
			body:     "not json",
			contains: []string{"decoding response"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useTempConfigDir(t)
			clearUdashEnv(t)

			server, _ := publishServer(t, tt.status, tt.body)
			t.Setenv(DefaultEnvVariableAPIURL, server.URL)
			t.Setenv(DefaultEnvVariableURL, "https://udash.example.com")

			r := reports.Report{Name: "pipeline"}
			err := Publish(&r)
			require.Error(t, err)
			for _, s := range tt.contains {
				assert.Contains(t, err.Error(), s)
			}
			assert.Empty(t, r.ReportURL)
		})
	}
}

func TestPublishWithoutEndpoint(t *testing.T) {
	useTempConfigDir(t)
	clearUdashEnv(t)

	r := reports.Report{Name: "pipeline"}
	assert.ErrorIs(t, Publish(&r), ErrNoUdashAPIURL)
}

func TestPublishSelectsStoredCredential(t *testing.T) {
	useTempConfigDir(t)
	clearUdashEnv(t)

	wanted, seen := publishServer(t, http.StatusCreated, `{"ReportID":"abc"}`)
	other, otherSeen := publishServer(t, http.StatusCreated, `{"ReportID":"abc"}`)

	require.NoError(t, updateConfigFile(authData{URL: wanted.URL, API: wanted.URL, Token: "wanted"}))
	// Stored last, so it is the default one
	require.NoError(t, updateConfigFile(authData{URL: other.URL, API: other.URL, Token: "other"}))

	APIURLSelector = wanted.URL

	r := reports.Report{Name: "pipeline"}
	require.NoError(t, Publish(&r))

	assert.Equal(t, "Bearer wanted", seen.authorization)
	assert.Empty(t, otherSeen.path)
}

func TestIsConfigured(t *testing.T) {
	t.Run("nothing configured", func(t *testing.T) {
		useTempConfigDir(t)
		clearUdashEnv(t)
		assert.False(t, IsConfigured())
	})

	t.Run("environment variable", func(t *testing.T) {
		useTempConfigDir(t)
		clearUdashEnv(t)
		t.Setenv(DefaultEnvVariableAPIURL, "https://udash.example.com/api")
		assert.True(t, IsConfigured())
	})

	t.Run("config file", func(t *testing.T) {
		useTempConfigDir(t)
		clearUdashEnv(t)
		require.NoError(t, updateConfigFile(authData{
			URL: "https://udash.example.com",
			API: "https://udash.example.com/api",
		}))
		assert.True(t, IsConfigured())
	})
}
