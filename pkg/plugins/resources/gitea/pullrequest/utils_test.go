package pullrequest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	giteasdk "gitea.dev/sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsRemoteBranchesExist(t *testing.T) {
	pages := [][]string{
		{"main", "feature-1"},
		{"feature-2", "feature-3"},
		{"updatecli", "feature-4"},
	}

	testData := []struct {
		name          string
		sourceBranch  string
		targetBranch  string
		expected      bool
		expectedPages int
	}{
		{
			name:          "Both branches on the first page",
			sourceBranch:  "feature-1",
			targetBranch:  "main",
			expected:      true,
			expectedPages: 1,
		},
		{
			name:          "Source branch on the last page",
			sourceBranch:  "updatecli",
			targetBranch:  "main",
			expected:      true,
			expectedPages: 3,
		},
		{
			name:          "Source branch missing",
			sourceBranch:  "missing",
			targetBranch:  "main",
			expected:      false,
			expectedPages: 3,
		},
		{
			name:          "Target branch missing",
			sourceBranch:  "updatecli",
			targetBranch:  "missing",
			expected:      false,
			expectedPages: 3,
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			requestedPages := 0

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				page, err := strconv.Atoi(r.URL.Query().Get("page"))
				if err != nil || page < 1 || page > len(pages) {
					http.Error(w, "unexpected page", http.StatusBadRequest)
					return
				}
				requestedPages++

				if page < len(pages) {
					w.Header().Set("Link", fmt.Sprintf(`<http://%s%s?page=%d>; rel="next"`, r.Host, r.URL.Path, page+1))
				}

				branches := []giteasdk.Branch{}
				for _, name := range pages[page-1] {
					branches = append(branches, giteasdk.Branch{Name: name})
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(branches)
			}))
			defer server.Close()

			client, err := giteasdk.NewClient(server.URL)
			require.NoError(t, err)

			g := Gitea{
				spec: Spec{
					Owner:        "updatecli",
					Repository:   "updatecli",
					SourceBranch: tt.sourceBranch,
					TargetBranch: tt.targetBranch,
				},
				client: client,
			}

			got, err := g.isRemoteBranchesExist(t.Context())

			require.NoError(t, err)
			assert.Equal(t, tt.expected, got)
			assert.Equal(t, tt.expectedPages, requestedPages)
		})
	}
}
