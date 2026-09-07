package branch

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

// newPaginatedBranchServer serves the provided pages of branches, which the GitLab API
// orders by name, and records which pages were requested.
func newPaginatedBranchServer(t *testing.T, pages [][]string) (*httptest.Server, *[]int64) {
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

// TestSearchBranchesPagination checks that a repository holding more branches than a
// single page reports each of them once, in the order GitLab lists them.
func TestSearchBranchesPagination(t *testing.T) {
	pages := [][]string{
		{
			branchJSON("v1.0", daysAgo(30)),
			branchJSON("v2.0", daysAgo(10)),
		},
		{
			branchJSON("v3.0", daysAgo(1)),
		},
	}

	server, requested := newPaginatedBranchServer(t, pages)
	defer server.Close()

	got, err := newBranchResource(t, server.URL, age.Spec{}).SearchBranches(age.Spec{})
	require.NoError(t, err)

	assert.Equal(t, []string{"v1.0", "v2.0", "v3.0"}, got)
	// Every page is visited exactly once, rather than the first one twice.
	assert.Equal(t, []int64{1, 2}, *requested)
}
