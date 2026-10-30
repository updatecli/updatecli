package pullrequest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	giteasdk "gitea.dev/sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/reports"
)

const (
	testRepoID  int64 = 1
	testForkID  int64 = 2
	testPRIndex int64 = 42
)

func intPtr(v int) *int {
	return &v
}

// cleanServer fakes the Gitea API endpoints used to clean a pull request.
type cleanServer struct {
	// noPullRequest removes the pull request from the list
	noPullRequest bool
	headRepoID    int64
	changedFiles  *int
	// listedSha returns the pull request ref commit for each list request, the last value being repeated
	listedSha []string
	branchSha string
	// replacementIndex is the index of the pull request listed from the second list request
	replacementIndex int64
	// title and body describe the existing pull request
	title string
	body  string

	closeStatus   int
	commentStatus int

	listCount   int
	closeBody   string
	closeCount  int
	commentBody string
}

func (s *cleanServer) handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	pullsPath := "/api/v1/repos/updatecli/updatecli/pulls"
	prPath := fmt.Sprintf("%s/%d", pullsPath, testPRIndex)
	commentPath := fmt.Sprintf("/api/v1/repos/updatecli/updatecli/issues/%d/comments", testPRIndex)

	pullRequest := func(sha string) giteasdk.PullRequest {
		return giteasdk.PullRequest{
			Index:   testPRIndex,
			URL:     "https://gitea.example.com/api/v1/repos/updatecli/updatecli/pulls/42",
			HTMLURL: "https://gitea.example.com/updatecli/updatecli/pulls/42",
			Title:   s.title,
			Body:    s.body,
			State:   giteasdk.StateOpen,
			Head:    &giteasdk.PRBranchInfo{Name: "updatecli", Sha: sha, RepoID: s.headRepoID},
			Base:    &giteasdk.PRBranchInfo{Name: "main", RepoID: testRepoID},
		}
	}

	switch {
	case r.Method == http.MethodGet && r.URL.Path == pullsPath:
		s.listCount++
		pullRequests := []giteasdk.PullRequest{}
		if !s.noPullRequest {
			sha := s.listedSha[min(s.listCount, len(s.listedSha))-1]
			pr := pullRequest(sha)
			if s.replacementIndex != 0 && s.listCount > 1 {
				pr.Index = s.replacementIndex
			}
			pullRequests = append(pullRequests, pr)
		}
		_ = json.NewEncoder(w).Encode(pullRequests)

	case r.Method == http.MethodGet && r.URL.Path == prPath:
		pr := pullRequest(s.branchSha)
		pr.ChangedFiles = s.changedFiles
		_ = json.NewEncoder(w).Encode(pr)

	case r.Method == http.MethodPatch && r.URL.Path == prPath:
		s.closeCount++
		body, _ := io.ReadAll(r.Body)
		s.closeBody = string(body)
		if s.closeStatus != 0 {
			w.WriteHeader(s.closeStatus)
			return
		}
		// Like Gitea, only the fields set in the request are updated
		pr := pullRequest(s.branchSha)
		var opt giteasdk.EditPullRequestOption
		_ = json.Unmarshal(body, &opt)
		if opt.Title != "" {
			pr.Title = opt.Title
		}
		if opt.Body != nil {
			pr.Body = *opt.Body
		}
		_ = json.NewEncoder(w).Encode(pr)

	case r.Method == http.MethodPost && r.URL.Path == commentPath:
		body, _ := io.ReadAll(r.Body)
		s.commentBody = string(body)
		if s.commentStatus != 0 {
			w.WriteHeader(s.commentStatus)
			return
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(giteasdk.Comment{})

	default:
		http.Error(w, "unexpected request "+r.Method+" "+r.URL.Path, http.StatusNotFound)
	}
}

func TestCleanAction(t *testing.T) {
	testData := []struct {
		name            string
		server          cleanServer
		expectedError   bool
		expectedClosed  bool
		expectedComment bool
		expectedLists   int
		expectedSleeps  int
	}{
		{
			name:          "No pull request found",
			server:        cleanServer{noPullRequest: true},
			expectedLists: 1,
		},
		{
			name: "Pull request with changed files",
			server: cleanServer{
				headRepoID:   testRepoID,
				changedFiles: intPtr(3),
				listedSha:    []string{"abc"},
				branchSha:    "abc",
			},
			expectedLists: 1,
		},
		{
			name: "Empty pull request",
			server: cleanServer{
				headRepoID:   testRepoID,
				changedFiles: intPtr(0),
				listedSha:    []string{"abc"},
				branchSha:    "abc",
			},
			expectedClosed:  true,
			expectedComment: true,
			expectedLists:   1,
		},
		{
			name: "Unknown number of changed files",
			server: cleanServer{
				headRepoID: testRepoID,
				listedSha:  []string{"abc"},
				branchSha:  "abc",
			},
			expectedLists: 1,
		},
		{
			name: "Pull request ref catches up with the working branch",
			server: cleanServer{
				headRepoID:   testRepoID,
				changedFiles: intPtr(0),
				listedSha:    []string{"old", "abc"},
				branchSha:    "abc",
			},
			expectedClosed:  true,
			expectedComment: true,
			expectedLists:   2,
			expectedSleeps:  1,
		},
		{
			name: "Pull request ref never catches up with the working branch",
			server: cleanServer{
				headRepoID:   testRepoID,
				changedFiles: intPtr(0),
				listedSha:    []string{"old"},
				branchSha:    "abc",
			},
			expectedLists:  cleanupMaxAttempts,
			expectedSleeps: cleanupMaxAttempts - 1,
		},
		{
			name: "Pull request replaced while waiting for its ref to catch up",
			server: cleanServer{
				headRepoID:       testRepoID,
				changedFiles:     intPtr(0),
				listedSha:        []string{"old", "abc"},
				branchSha:        "abc",
				replacementIndex: testPRIndex + 1,
			},
			expectedLists:  2,
			expectedSleeps: 1,
		},
		{
			name: "Pull request opened from a fork",
			server: cleanServer{
				headRepoID:   testForkID,
				changedFiles: intPtr(0),
				listedSha:    []string{"abc"},
				branchSha:    "abc",
			},
			expectedLists: 1,
		},
		{
			name: "Failing to comment isn't an error",
			server: cleanServer{
				headRepoID:    testRepoID,
				changedFiles:  intPtr(0),
				listedSha:     []string{"abc"},
				branchSha:     "abc",
				commentStatus: http.StatusInternalServerError,
			},
			expectedClosed:  true,
			expectedComment: true,
			expectedLists:   1,
		},
		{
			name: "Failing to close is an error",
			server: cleanServer{
				headRepoID:   testRepoID,
				changedFiles: intPtr(0),
				listedSha:    []string{"abc"},
				branchSha:    "abc",
				closeStatus:  http.StatusForbidden,
			},
			expectedError: true,
			expectedLists: 1,
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			sleeps := 0
			originalSleep := cleanupSleep
			cleanupSleep = func(time.Duration) { sleeps++ }
			defer func() { cleanupSleep = originalSleep }()

			fake := tt.server
			server := httptest.NewServer(http.HandlerFunc(fake.handler))
			defer server.Close()

			client, err := giteasdk.NewClient(server.URL)
			require.NoError(t, err)

			g := Gitea{
				client:       client,
				Owner:        "updatecli",
				Repository:   "updatecli",
				SourceBranch: "updatecli",
				TargetBranch: "main",
			}

			report := reports.Action{
				Link:        "https://gitea.example.com/updatecli/updatecli/pulls/42",
				Description: "original",
			}

			err = g.CleanAction(t.Context(), &report)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedLists, fake.listCount)
			assert.Equal(t, tt.expectedSleeps, sleeps)

			if tt.expectedClosed {
				assert.Equal(t, 1, fake.closeCount)
				assert.Contains(t, fake.closeBody, `"state":"closed"`)
				assert.Empty(t, report.Link)
				assert.Equal(t, pullRequestClosedMessage, report.Description)
			} else {
				assert.NotEmpty(t, report.Link)
				assert.Equal(t, "original", report.Description)
			}

			if tt.expectedComment {
				assert.Contains(t, fake.commentBody, pullRequestClosedMessage)
			} else {
				assert.Empty(t, fake.commentBody)
			}
		})
	}
}

func TestFindExistingPullRequest(t *testing.T) {
	pages := [][]giteasdk.PullRequest{
		{
			{Index: 1, Head: &giteasdk.PRBranchInfo{Name: "feature-1"}, Base: &giteasdk.PRBranchInfo{Name: "main"}},
		},
		{
			{Index: 2, Head: &giteasdk.PRBranchInfo{Name: "feature-2"}, Base: &giteasdk.PRBranchInfo{Name: "main"}},
		},
		{
			{Index: 3, Head: &giteasdk.PRBranchInfo{Name: "updatecli"}, Base: &giteasdk.PRBranchInfo{Name: "main"}},
		},
	}

	testData := []struct {
		name          string
		sourceBranch  string
		expectedIndex int64
		expectedNil   bool
	}{
		{
			name:          "Pull request on the last page",
			sourceBranch:  "updatecli",
			expectedIndex: 3,
		},
		{
			name:         "Pull request missing",
			sourceBranch: "missing",
			expectedNil:  true,
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			requestedPages := []int{}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				page, err := strconv.Atoi(r.URL.Query().Get("page"))
				if err != nil || page < 1 || page > len(pages) {
					http.Error(w, "unexpected page", http.StatusBadRequest)
					return
				}
				requestedPages = append(requestedPages, page)

				if page < len(pages) {
					w.Header().Set("Link", fmt.Sprintf(`<http://%s%s?page=%d>; rel="next"`, r.Host, r.URL.Path, page+1))
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(pages[page-1])
			}))
			defer server.Close()

			client, err := giteasdk.NewClient(server.URL)
			require.NoError(t, err)

			g := Gitea{
				client:       client,
				Owner:        "updatecli",
				Repository:   "updatecli",
				SourceBranch: tt.sourceBranch,
				TargetBranch: "main",
			}

			got, err := g.findExistingPullRequest(t.Context())
			require.NoError(t, err)

			assert.Equal(t, []int{1, 2, 3}, requestedPages)

			if tt.expectedNil {
				assert.Nil(t, got)
				return
			}

			require.NotNil(t, got)
			assert.Equal(t, tt.expectedIndex, got.Index)
		})
	}
}
