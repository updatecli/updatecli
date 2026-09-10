package github

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/shurcooL/githubv4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github/client"
)

/*
mockPullRequestClient implements the GitHub v4 client interface to answer the two queries
used while cleaning a pull request:
  - the head commit of the working branch
  - the pull request associated to the working branch

Both answers are provided as a list so each call can return a different value, which is
what happens when Updatecli waits for the GitHub API to catch up with what it just pushed.
*/
type mockPullRequestClient struct {
	// remoteBranchHeadOids contains the head commit returned for each query
	remoteBranchHeadOids []string
	// pullRequests contains the pull request returned for each query.
	// An empty ID means that no open pull request exists.
	pullRequests []PullRequestApi

	remoteBranchHeadOidCounter int
	pullRequestCounter         int
}

func (m *mockPullRequestClient) Query(ctx context.Context, q interface{}, variables map[string]interface{}) error {
	query := reflect.ValueOf(q).Elem()

	repository := query.FieldByName("Repository")
	// Only the rate limit is queried
	if !repository.IsValid() {
		return nil
	}

	if pullRequests := repository.FieldByName("PullRequests"); pullRequests.IsValid() {
		pullRequest := next(m.pullRequests, &m.pullRequestCounter)

		nodes := []PullRequestApi{pullRequest}
		if pullRequest.ID == "" {
			nodes = []PullRequestApi{}
		}

		pullRequests.FieldByName("Nodes").Set(reflect.ValueOf(nodes))

		return nil
	}

	// Otherwise this is the query looking for the head commit of a branch
	defaultBranchRef := repository.FieldByName("DefaultBranchRef")
	defaultBranchRef.Set(reflect.New(defaultBranchRef.Type().Elem()))
	defaultBranchRef.Elem().FieldByName("Target").FieldByName("Oid").SetString("defaultBranchOid")

	headOid := next(m.remoteBranchHeadOids, &m.remoteBranchHeadOidCounter)
	// An empty head commit means that the branch doesn't exist
	if headOid == "" {
		return nil
	}

	ref := repository.FieldByName("Ref")
	ref.Set(reflect.New(ref.Type().Elem()))
	ref.Elem().FieldByName("Target").FieldByName("Oid").SetString(headOid)

	return nil
}

func (m *mockPullRequestClient) Mutate(ctx context.Context, mutation interface{}, input githubv4.Input, variables map[string]interface{}) error {
	return nil
}

// next returns the value matching the current counter, or the last one once the list is
// exhausted, then increments the counter.
func next[T any](values []T, counter *int) T {
	var value T

	if len(values) == 0 {
		return value
	}

	if *counter < len(values) {
		value = values[*counter]
	} else {
		value = values[len(values)-1]
	}

	*counter++

	return value
}

func newTestPullRequest(c client.Client, remotePullRequest PullRequestApi) PullRequest {
	return PullRequest{
		gh: &Github{
			Spec: Spec{
				Owner:      "updatecli",
				Repository: "updatecli",
				Branch:     "main",
			},
			client: c,
		},
		repository: &Repository{
			Owner: "updatecli",
			Name:  "updatecli",
		},
		remotePullRequest: remotePullRequest,
	}
}

func TestIsPullRequestEmpty(t *testing.T) {
	testdata := []struct {
		name                 string
		remotePullRequest    PullRequestApi
		remoteBranchHeadOids []string
		pullRequests         []PullRequestApi
		expectedResult       bool
		expectedSleeps       int
	}{
		{
			name: "pull request up to date with the working branch and without any changed file",
			remotePullRequest: PullRequestApi{
				ID:           "prid",
				HeadRefOid:   "commitA",
				ChangedFiles: 0,
			},
			remoteBranchHeadOids: []string{"commitA"},
			expectedResult:       true,
		},
		{
			name: "pull request up to date with the working branch and with changed files",
			remotePullRequest: PullRequestApi{
				ID:           "prid",
				HeadRefOid:   "commitA",
				ChangedFiles: 2,
			},
			remoteBranchHeadOids: []string{"commitA"},
			expectedResult:       false,
		},
		{
			/*
				GitHub didn't process yet the commit that Updatecli just pushed so the
				pull request temporarily reports no changed file.
				Once the pull request state caught up, it does contain a changed file
				so it must not be closed.
			*/
			name: "pull request state is stale then catch up with the working branch",
			remotePullRequest: PullRequestApi{
				ID:           "prid",
				HeadRefOid:   "commitA",
				ChangedFiles: 0,
			},
			remoteBranchHeadOids: []string{"commitB"},
			pullRequests: []PullRequestApi{
				{
					ID:           "prid",
					HeadRefOid:   "commitB",
					ChangedFiles: 1,
				},
			},
			expectedResult: false,
			expectedSleeps: 1,
		},
		{
			name: "pull request state never catches up with the working branch",
			remotePullRequest: PullRequestApi{
				ID:           "prid",
				HeadRefOid:   "commitA",
				ChangedFiles: 0,
			},
			remoteBranchHeadOids: []string{"commitB"},
			pullRequests: []PullRequestApi{
				{
					ID:           "prid",
					HeadRefOid:   "commitA",
					ChangedFiles: 0,
				},
			},
			expectedResult: false,
			expectedSleeps: 2,
		},
		{
			name: "working branch doesn't exist anymore",
			remotePullRequest: PullRequestApi{
				ID:           "prid",
				HeadRefOid:   "commitA",
				ChangedFiles: 0,
			},
			remoteBranchHeadOids: []string{""},
			pullRequests: []PullRequestApi{
				{
					ID:           "prid",
					HeadRefOid:   "commitA",
					ChangedFiles: 0,
				},
			},
			expectedResult: false,
			expectedSleeps: 2,
		},
		{
			name: "pull request isn't open anymore",
			remotePullRequest: PullRequestApi{
				ID:           "prid",
				HeadRefOid:   "commitA",
				ChangedFiles: 0,
			},
			remoteBranchHeadOids: []string{"commitB"},
			pullRequests:         []PullRequestApi{{}},
			expectedResult:       false,
			expectedSleeps:       1,
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			sleeps := 0
			defaultCleanupSleep := cleanupSleep
			cleanupSleep = func(d time.Duration) { sleeps++ }
			defer func() { cleanupSleep = defaultCleanupSleep }()

			mockClient := mockPullRequestClient{
				remoteBranchHeadOids: tt.remoteBranchHeadOids,
				pullRequests:         tt.pullRequests,
			}

			pullRequest := newTestPullRequest(&mockClient, tt.remotePullRequest)

			gotResult, gotErr := pullRequest.isPullRequestEmpty(context.Background(), cleanupMaxAttempts)

			require.NoError(t, gotErr)
			assert.Equal(t, tt.expectedResult, gotResult)
			assert.Equal(t, tt.expectedSleeps, sleeps)
		})
	}
}
