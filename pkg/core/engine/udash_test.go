package engine

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/updatecli/updatecli/pkg/core/cmdoptions"
	"github.com/updatecli/updatecli/pkg/core/pipeline"
	"github.com/updatecli/updatecli/pkg/core/udash"
)

func TestPublishToUdashFailureIsTyped(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("HOME", dir)
	t.Setenv(udash.DefaultEnvVariableURL, "")
	t.Setenv(udash.DefaultEnvVariableAccessToken, "")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)
	t.Setenv(udash.DefaultEnvVariableAPIURL, server.URL)

	previous := cmdoptions.Experimental
	cmdoptions.Experimental = true
	t.Cleanup(func() { cmdoptions.Experimental = previous })

	e := Engine{Pipelines: []*pipeline.Pipeline{{Name: "first"}, {Name: "second"}}}

	err := e.publishToUdash()
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUdashPublish)
	assert.Contains(t, err.Error(), `pipeline "first"`)
	assert.Contains(t, err.Error(), `pipeline "second"`)
}

func TestPublishToUdashSkippedWithoutExperimental(t *testing.T) {
	previous := cmdoptions.Experimental
	cmdoptions.Experimental = false
	t.Cleanup(func() { cmdoptions.Experimental = previous })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Udash must not be called without --experimental")
	}))
	t.Cleanup(server.Close)
	t.Setenv(udash.DefaultEnvVariableAPIURL, server.URL)

	e := Engine{Pipelines: []*pipeline.Pipeline{{Name: "first"}}}
	assert.NoError(t, e.publishToUdash())
}
