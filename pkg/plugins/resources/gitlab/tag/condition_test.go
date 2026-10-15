package tag

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

// newProjectTagServer serves the given tags for any project path, except one
// containing "nonexistent" in its name, which GitLab reports as a 404.
func newProjectTagServer(t *testing.T, tags ...string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "nonexistent") {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"message":"404 Project Not Found"}`)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		taggedJSON := make([]string, 0, len(tags))
		for _, tag := range tags {
			taggedJSON = append(taggedJSON, tagJSON(tag, daysAgo(1)))
		}
		fmt.Fprintf(w, "[%s]", strings.Join(taggedJSON, ","))
	}))
}

func TestCondition(t *testing.T) {

	server := newProjectTagServer(t, "v0.1.0", "v0.2.0", "v0.3.0", "v1.0.0")
	defer server.Close()

	tests := []struct {
		name     string
		manifest struct {
			URL        string
			Token      string
			Owner      string
			Repository string
			Tag        string
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
				Tag        string
			}{
				URL:        server.URL,
				Token:      "",
				Owner:      "olblak",
				Repository: "updatecli-nonexistent",
			},
			wantResult: false,
			wantErr:    true,
		},
		{
			name: "repository olblak/updatecli should exist with tags",
			manifest: struct {
				URL        string
				Token      string
				Owner      string
				Repository string
				Tag        string
			}{
				URL:        server.URL,
				Token:      "",
				Owner:      "olblak",
				Repository: "updatecli",
			},
			wantResult: false,
		},
		{
			name: "repository should exist with no tag v0.1.11",
			manifest: struct {
				URL        string
				Token      string
				Owner      string
				Repository string
				Tag        string
			}{
				URL:        server.URL,
				Token:      "",
				Owner:      "olblak",
				Repository: "updatecli",
				Tag:        "v0.1.11",
			},
			wantResult: false,
		},
		{
			name: "repository should exist with tag v0.3.0",
			manifest: struct {
				URL        string
				Token      string
				Owner      string
				Repository string
				Tag        string
			}{
				URL:        server.URL,
				Token:      "",
				Owner:      "olblak",
				Repository: "updatecli",
				Tag:        "v0.3.0",
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
				require.Error(t, gotErr)
			} else {
				require.NoError(t, gotErr)
			}

			assert.Equal(t, tt.wantResult, gotResult)

		})

	}
}
