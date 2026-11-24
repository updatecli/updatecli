package release

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
)

func TestCondition(t *testing.T) {

	tests := []struct {
		name     string
		manifest struct {
			URL        string
			Token      string
			Owner      string
			Repository string
			Tag        string
		}
		wantResult     bool
		wantErr        bool
		wantErrMessage error
	}{
		{
			name: "repository olblak/updatecli should not exist",
			manifest: struct {
				URL        string
				Token      string
				Owner      string
				Repository string
				Tag        string
			}{
				URL:        "codeberg.org",
				Token:      "",
				Owner:      "updatecli",
				Repository: "updatecli-nonexistent",
			},
			wantResult:     false,
			wantErr:        true,
			wantErrMessage: fmt.Errorf("looking for Gitea release: Not Found"),
		},
		{
			name: "repository olblak/updatecli-mirror should exist but no release",
			manifest: struct {
				URL        string
				Token      string
				Owner      string
				Repository string
				Tag        string
			}{
				URL:        "codeberg.org",
				Token:      "",
				Owner:      "updatecli",
				Repository: "demo-terminal",
			},
			wantResult: false,
		},
		{
			name: "repository should exist with no release 2.0.0",
			manifest: struct {
				URL        string
				Token      string
				Owner      string
				Repository string
				Tag        string
			}{
				URL:        "codeberg.org",
				Token:      "",
				Owner:      "updatecli",
				Repository: "demo-terminal",
				Tag:        "2.0.0",
			},
			wantResult: false,
		},
		{
			name: "repository should exist with release v2.15.0",
			manifest: struct {
				URL        string
				Token      string
				Owner      string
				Repository string
				Tag        string
			}{
				URL:        "codeberg.org",
				Token:      "",
				Owner:      "updatecli",
				Repository: "updatecli-action",
				Tag:        "v2.15.0",
			},
			wantResult: true,
			wantErr:    false,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			g, gotErr := New(tt.manifest)
			require.NoError(t, gotErr)

			gotResult, _, gotErr := g.Condition(context.Background(), "", nil, pathresolver.Resolver{})

			if tt.wantErr {
				if assert.Error(t, gotErr) {
					assert.Equal(t, gotErr.Error(), tt.wantErrMessage.Error())
				}
			} else {
				require.NoError(t, gotErr)
			}

			assert.Equal(t, tt.wantResult, gotResult)

		})

	}
}

func TestConditionReleasesDisabled(t *testing.T) {
	const repoPath = "/forge/api/v1/repos/updatecli/project"
	settings := func(body string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
			fmt.Fprint(w, body)
		}
	}
	disabled := settings(`{"has_releases":false}`)
	otherForge := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("followed a metadata redirect to another forge")
	}))
	defer otherForge.Close()

	for _, tt := range []struct {
		name    string
		status  int
		repo    http.Handler
		wantErr bool
	}{
		{"disabled on Forgejo", http.StatusNotFound, disabled, false},
		{"disabled on Gitea after a rename", http.StatusForbidden, http.RedirectHandler(repoPath+"-new", http.StatusMovedPermanently), false},
		{"enabled but forbidden", http.StatusForbidden, settings(`{"has_releases":true}`), true},
		{"missing setting", http.StatusNotFound, settings(`{}`), true},
		{"redirect to another forge", http.StatusForbidden, http.RedirectHandler(otherForge.URL, http.StatusMovedPermanently), true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("GET "+repoPath+"/releases", func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
			})
			mux.Handle("GET "+repoPath, tt.repo)
			mux.Handle("GET "+repoPath+"-new", disabled)
			server := httptest.NewServer(mux)
			defer server.Close()

			g, err := New(map[string]any{
				"url":        server.URL + "/forge",
				"owner":      "updatecli",
				"repository": "project",
				"token":      "test-token",
			})
			require.NoError(t, err)
			pass, _, err := g.Condition(t.Context(), "v1.0.0", nil, pathresolver.Resolver{})
			assert.False(t, pass)
			if tt.wantErr {
				require.EqualError(t, err, "looking for Gitea release: "+http.StatusText(tt.status))
			} else {
				require.NoError(t, err)
			}
		})
	}
}
