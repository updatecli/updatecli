package release

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

// newProjectReleaseServer serves the given releases for most project paths, an empty
// release list for "FOSDEM22" and a 404 for any path containing "nonexistent", mirroring
// the handful of real-world repository states exercised by condition/source tests.
func newProjectReleaseServer(t *testing.T, releases ...string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "nonexistent"):
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"message":"404 Project Not Found"}`)
		case strings.Contains(r.URL.Path, "FOSDEM22"):
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, "[]")
		default:
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, "[%s]", strings.Join(releases, ","))
		}
	}))
}

func TestCondition(t *testing.T) {

	server := newProjectReleaseServer(t, releaseJSON("v0.46.2", daysAgo(1), false), releaseJSON("v0.46.1", daysAgo(10), false))
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
			name: "repository cicd-devroom/FOSDEM22 should exist but no release",
			manifest: struct {
				URL        string
				Token      string
				Owner      string
				Repository string
				Tag        string
			}{
				URL:        server.URL,
				Token:      "",
				Owner:      "cicd-devroom",
				Repository: "FOSDEM22",
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
				URL:        server.URL,
				Token:      "",
				Owner:      "cicd-devroom",
				Repository: "FOSDEM22",
				Tag:        "2.0.0",
			},
			wantResult: false,
		},
		{
			name: "repository should exist with release v0.46.2",
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
				Tag:        "v0.46.2",
			},
			wantResult: true,
			wantErr:    false,
		},
		{
			name: "repository should exist with no release v0.0.99",
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
				Tag:        "v0.0.99",
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
