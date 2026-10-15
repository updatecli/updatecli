package branch

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
)

// newProjectBranchServer serves the given branches for any project path, except one
// containing "nonexistent" in its name, which GitLab reports as a 404.
func newProjectBranchServer(t *testing.T, branches ...string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "nonexistent") {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"message":"404 Project Not Found"}`)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		taggedJSON := make([]string, 0, len(branches))
		for _, branch := range branches {
			taggedJSON = append(taggedJSON, branchJSON(branch, daysAgo(1)))
		}
		fmt.Fprintf(w, "[%s]", strings.Join(taggedJSON, ","))
	}))
}

func TestCondition(t *testing.T) {

	server := newProjectBranchServer(t, "main", "develop")
	defer server.Close()

	tests := []struct {
		name     string
		manifest struct {
			URL        string
			Token      string
			Owner      string
			Repository string
			Branch     string
		}
		wantResult bool
		wantErr    bool
	}{
		{
			name: "repository olblak/updatecli-nonexistent should not exist",
			manifest: struct {
				URL        string
				Token      string
				Owner      string
				Repository string
				Branch     string
			}{
				URL:        server.URL,
				Token:      "",
				Owner:      "olblak",
				Repository: "updatecli-nonexistent",
				Branch:     "v2",
			},
			wantResult: false,
			wantErr:    true,
		},
		{
			name: "repository olblak/updatecli-mirror should exist with branches",
			manifest: struct {
				URL        string
				Token      string
				Owner      string
				Repository string
				Branch     string
			}{
				URL:        server.URL,
				Token:      "",
				Owner:      "olblak",
				Repository: "updatecli",
				Branch:     "main",
			},
			wantResult: true,
			wantErr:    false,
		},
		{
			name: "repository olblak/updatecli should not have branch nonexistent",
			manifest: struct {
				URL        string
				Token      string
				Owner      string
				Repository string
				Branch     string
			}{
				URL:        server.URL,
				Token:      "",
				Owner:      "olblak",
				Repository: "updatecli",
				Branch:     "doesnotexist",
			},
			wantResult: false,
			wantErr:    false,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			g, gotErr := New(tt.manifest)
			require.NoError(t, gotErr)

			gotResult, _, gotErr := g.Condition(context.Background(), "", nil, pathresolver.Resolver{})

			if tt.wantErr {
				require.Error(t, gotErr)
			} else {
				require.NoError(t, gotErr)
			}

			assert.Equal(t, tt.wantResult, gotResult)

		})

	}
}
