package gitlabsearch

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScmsGenerator(t *testing.T) {
	for _, tt := range []struct {
		name         string
		limit        int
		branch       string
		failurePage  string
		wantErr      string
		wantBranches []string
		wantPages    []string
	}{
		{
			name:         "Stop after the first match",
			limit:        1,
			branch:       "^main$",
			failurePage:  "one/2",
			wantBranches: []string{"one/main"},
			wantPages:    []string{"one/1"},
		},
		{
			name:         "Matching branch on a later page",
			limit:        1,
			branch:       "^stable$",
			wantBranches: []string{"one/stable"},
			wantPages:    []string{"one/1", "one/2"},
		},
		{
			name:         "Limit spans repositories",
			limit:        3,
			branch:       "^(main|stable)$",
			failurePage:  "two/2",
			wantBranches: []string{"one/main", "one/stable", "two/main"},
			wantPages:    []string{"one/1", "one/2", "two/1"},
		},
		{
			name:         "Unlimited",
			branch:       "^(main|stable)$",
			wantBranches: []string{"one/main", "one/stable", "two/main", "two/stable"},
			wantPages:    []string{"one/1", "one/2", "two/1", "two/2"},
		},
		{
			name:         "Negative limit remains unlimited",
			limit:        -1,
			branch:       "^(main|stable)$",
			wantBranches: []string{"one/main", "one/stable", "two/main", "two/stable"},
			wantPages:    []string{"one/1", "one/2", "two/1", "two/2"},
		},
		{
			name:        "Required page fails",
			limit:       1,
			branch:      "^stable$",
			failurePage: "one/2",
			wantErr:     "branch page failed",
			wantPages:   []string{"one/1", "one/2"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var pages []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch r.URL.Path {
				case "/api/v4/groups/updatecli/projects":
					_, _ = w.Write([]byte(`[{"path_with_namespace":"updatecli/one"},{"path_with_namespace":"updatecli/two"}]`))
				case "/api/v4/projects/updatecli/one/repository/branches", "/api/v4/projects/updatecli/two/repository/branches":
					repository := strings.Split(r.URL.Path, "/")[5]
					page := r.URL.Query().Get("page")
					pages = append(pages, repository+"/"+page)
					if repository+"/"+page == tt.failurePage {
						w.WriteHeader(http.StatusBadRequest)
						_, _ = w.Write([]byte(`{"message":"branch page failed"}`))
						return
					}
					if page == "1" {
						w.Header().Set("X-Next-Page", "2")
						_, _ = w.Write([]byte(`[{"name":"develop"},{"name":"main"}]`))
					} else {
						_, _ = w.Write([]byte(`[{"name":"stable"}]`))
					}
				default:
					t.Errorf("unexpected request: %s", r.URL)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()

			search, err := New(map[string]any{
				"url":    server.URL,
				"group":  "updatecli",
				"branch": tt.branch,
				"limit":  tt.limit,
			})
			require.NoError(t, err)
			specs, err := search.ScmsGenerator(context.Background())
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			var branches []string
			for _, spec := range specs {
				branches = append(branches, spec.Repository+"/"+spec.Branch)
			}
			assert.Equal(t, tt.wantBranches, branches)
			assert.Equal(t, tt.wantPages, pages)
		})
	}
}
