package mergerequest

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// newMergeRequestServer serves a fixed "main" branch for branch lookups and an empty
// merge request list, so that branch/MR checks are exercised without reaching the real
// GitLab API.
func newMergeRequestServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "/repository/branches"):
			w.Header().Set("X-Total-Pages", "1")
			fmt.Fprint(w, `[{"name":"main"}]`)
		case strings.Contains(r.URL.Path, "/merge_requests"):
			fmt.Fprint(w, "[]")
		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"message":"404 Not Found"}`)
		}
	}))
}

func TestIsRemoteBranchExist(t *testing.T) {

	server := newMergeRequestServer(t)
	defer server.Close()

	testdata := []struct {
		name           string
		spec           map[string]interface{}
		expectedResult bool
	}{
		{
			name: "Existing branch",
			spec: map[string]interface{}{
				"URL":          server.URL,
				"Owner":        "olblak",
				"Repository":   "updatecli",
				"SourceBranch": "main",
				"TargetBranch": "main",
			},
			expectedResult: true,
		},
		{
			name: "Non existing branch",
			spec: map[string]interface{}{
				"URL":          server.URL,
				"Owner":        "olblak",
				"Repository":   "updatecli",
				"SourceBranch": "donotexist",
				"TargetBranch": "donotexist",
			},
			expectedResult: false,
		},
	}

	for _, td := range testdata {
		t.Run(td.name, func(t *testing.T) {
			gitlab, err := New(td.spec, nil)
			if err != nil {
				t.Fatalf("failed to create Gitlab instance: %v", err)
			}

			gotResult, err := gitlab.isRemoteBranchesExist()
			require.NoError(t, err)

			require.Equal(t, td.expectedResult, gotResult)

		})
	}
}
func TestFindExistingMR_Table(t *testing.T) {

	server := newMergeRequestServer(t)
	defer server.Close()

	testdata := []struct {
		name string
		spec map[string]interface{}
	}{
		{
			name: "NoMergeRequest",
			spec: map[string]interface{}{
				"URL":          server.URL,
				"Owner":        "olblak",
				"Repository":   "updatecli",
				"SourceBranch": "donotexist",
				"TargetBranch": "donotexist",
			},
		},
		{
			name: "NoOpenedMergeRequestBetweenMainBranches",
			spec: map[string]interface{}{
				"URL":          server.URL,
				"Owner":        "olblak",
				"Repository":   "updatecli",
				"SourceBranch": "main",
				"TargetBranch": "main",
			},
		},
	}

	for _, td := range testdata {
		tc := td
		t.Run(tc.name, func(t *testing.T) {
			gitlab, err := New(tc.spec, nil)
			if err != nil {
				t.Fatalf("failed to create Gitlab instance: %v", err)
			}

			mr, err := gitlab.findExistingMR()
			require.NoError(t, err)
			require.Nil(t, mr)
		})
	}
}
