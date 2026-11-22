package release

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
)

func TestSearchReleasesOrder(t *testing.T) {
	tests := []struct {
		name     string
		lastLink bool
	}{
		{name: "next and last links", lastLink: true},
		{name: "next link only"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var pages []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/repos/updatecli/project/releases", r.URL.Path)
				page := r.URL.Query().Get("page")
				pages = append(pages, page)
				w.Header().Set("Content-Type", "application/json")
				switch page {
				case "0", "1":
					link := fmt.Sprintf(`<http://%s/api/v1/repos/updatecli/project/releases?page=2>; rel="next"`, r.Host)
					if tt.lastLink {
						link += fmt.Sprintf(`, <http://%s/api/v1/repos/updatecli/project/releases?page=2>; rel="last"`, r.Host)
					}
					w.Header().Set("Link", link)
					fmt.Fprint(w, `[{"tag_name":"v3.0.0"},{"tag_name":"v2.0.0"}]`)
				case "2":
					fmt.Fprint(w, `[{"tag_name":"v1.0.0"}]`)
				default:
					http.Error(w, "unexpected page", http.StatusBadRequest)
				}
			}))
			defer server.Close()

			g, err := New(map[string]interface{}{
				"url":        server.URL,
				"owner":      "updatecli",
				"repository": "project",
				"token":      "test-token",
			})
			require.NoError(t, err)

			got, err := g.SearchReleases(t.Context(), age.Spec{})
			require.NoError(t, err)
			assert.Equal(t, []string{"v1.0.0", "v2.0.0", "v3.0.0"}, got)
			assert.Equal(t, []string{"1", "2"}, pages)
		})
	}
}
