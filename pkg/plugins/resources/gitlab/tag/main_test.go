package tag

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
)

// newPaginatedTagServer serves the provided pages of tags, the most recent tag first
// the way the GitLab API orders them, and records which pages were requested.
func newPaginatedTagServer(t *testing.T, pages [][]string) (*httptest.Server, *[]int64) {
	t.Helper()

	var mu sync.Mutex
	requested := []int64{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, err := strconv.ParseInt(r.URL.Query().Get("page"), 10, 64)
		if err != nil {
			page = 1
		}

		mu.Lock()
		requested = append(requested, page)
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Total-Pages", strconv.Itoa(len(pages)))
		if page < int64(len(pages)) {
			w.Header().Set("X-Next-Page", strconv.FormatInt(page+1, 10))
		}

		if page < 1 || page > int64(len(pages)) {
			fmt.Fprint(w, "[]")
			return
		}
		fmt.Fprintf(w, "[%s]", strings.Join(pages[page-1], ","))
	}))

	return server, &requested
}

// TestSearchTagsOrder covers the two assumptions the version filters make about the
// tags returned by SearchTags: the oldest one comes first, and each one appears once.
func TestSearchTagsOrder(t *testing.T) {
	pages := [][]string{
		{
			tagJSON("v3.0.0", daysAgo(1)),
			tagJSON("v2.0.0", daysAgo(10)),
		},
		{
			tagJSON("v1.0.0", daysAgo(30)),
		},
	}

	server, requested := newPaginatedTagServer(t, pages)
	defer server.Close()

	got, err := newTagResource(t, server.URL, age.Spec{}).SearchTags(age.Spec{})
	require.NoError(t, err)

	// The version filters read the latest version off the end of the list.
	assert.Equal(t, []string{"v1.0.0", "v2.0.0", "v3.0.0"}, got)
	// Every page is visited exactly once, rather than the first one twice.
	assert.Equal(t, []int64{1, 2}, *requested)
}
