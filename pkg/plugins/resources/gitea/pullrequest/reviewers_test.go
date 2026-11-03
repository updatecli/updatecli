package pullrequest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	giteasdk "gitea.dev/sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestReviewers(t *testing.T) {
	testData := []struct {
		name              string
		reviewers         []string
		pages             [][]giteasdk.PullReview
		maxResponseItems  int
		totalCount        string
		settingsStatus    int
		listStatus        int
		requestStatus     int
		expectedLimit     string
		expectedReviewers []string
		expectedPages     []int
		expectedError     string
	}{
		{
			name: "No reviewers configured",
		},
		{
			name:             "Add only missing reviewers across pages",
			reviewers:        []string{"alice", "bob", "carol", "CAROL"},
			maxResponseItems: 2,
			totalCount:       "3",
			pages: [][]giteasdk.PullReview{
				{
					{
						Reviewer: &giteasdk.User{UserName: "Alice"},
						State:    giteasdk.ReviewStateApproved,
					},
					{ReviewerTeam: &giteasdk.Team{ID: 7}},
				},
				{
					{
						Reviewer: &giteasdk.User{UserName: "bob"},
						State:    giteasdk.ReviewStateRequestReview,
					},
				},
			},
			expectedReviewers: []string{"carol"},
			expectedPages:     []int{1, 2},
			expectedLimit:     "2",
		},
		{
			name:             "An existing review is not requested again",
			reviewers:        []string{"alice"},
			maxResponseItems: 50,
			totalCount:       "1",
			pages: [][]giteasdk.PullReview{
				{
					{
						Reviewer: &giteasdk.User{UserName: "alice"},
						State:    giteasdk.ReviewStateComment,
					},
				},
			},
			expectedPages: []int{1},
			expectedLimit: "30",
		},
		{
			name:             "An empty page can hide a pending review before an approval",
			reviewers:        []string{"alice", "bob"},
			maxResponseItems: 1,
			totalCount:       "2",
			pages: [][]giteasdk.PullReview{
				{},
				{
					{
						Reviewer: &giteasdk.User{UserName: "alice"},
						State:    giteasdk.ReviewStateApproved,
					},
				},
			},
			expectedReviewers: []string{"bob"},
			expectedPages:     []int{1, 2},
			expectedLimit:     "1",
		},
		{
			name:           "API settings failure",
			reviewers:      []string{"alice"},
			settingsStatus: http.StatusForbidden,
			expectedError:  "listing reviews",
		},
		{
			name:             "Review listing failure",
			reviewers:        []string{"alice"},
			maxResponseItems: 50,
			listStatus:       http.StatusForbidden,
			expectedPages:    []int{1},
			expectedLimit:    "30",
			expectedError:    "listing reviews",
		},
		{
			name:              "Review request failure",
			reviewers:         []string{"alice"},
			maxResponseItems:  50,
			totalCount:        "0",
			requestStatus:     http.StatusUnprocessableEntity,
			expectedReviewers: []string{"alice"},
			expectedPages:     []int{1},
			expectedLimit:     "30",
			expectedError:     "requesting reviewers",
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			var pages []int
			var requested []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch {
				case r.URL.Path == "/api/v1/version":
					_, _ = w.Write([]byte(`{"version":"1.24.0"}`))
				case r.Method == http.MethodGet && r.URL.Path == "/api/v1/settings/api":
					assert.NotEmpty(t, tt.reviewers)
					if tt.settingsStatus != 0 {
						w.WriteHeader(tt.settingsStatus)
						return
					}
					_ = json.NewEncoder(w).Encode(giteasdk.GlobalAPISettings{MaxResponseItems: tt.maxResponseItems})
				case r.Method == http.MethodGet && r.URL.Path == "/api/v1/repos/updatecli/updatecli/pulls/42/reviews":
					assert.Equal(t, tt.expectedLimit, r.URL.Query().Get("limit"))
					page, err := strconv.Atoi(r.URL.Query().Get("page"))
					assert.NoError(t, err)
					pages = append(pages, page)
					if tt.listStatus != 0 {
						w.WriteHeader(tt.listStatus)
						return
					}
					if tt.totalCount != "" {
						w.Header().Set("X-Total-Count", tt.totalCount)
					}
					if page > 0 && page <= len(tt.pages) {
						_ = json.NewEncoder(w).Encode(tt.pages[page-1])
					} else {
						_, _ = w.Write([]byte(`[]`))
					}
				case r.Method == http.MethodPost && r.URL.Path == "/api/v1/repos/updatecli/updatecli/pulls/42/requested_reviewers":
					var opt giteasdk.PullReviewRequestOptions
					assert.NoError(t, json.NewDecoder(r.Body).Decode(&opt))
					assert.NotEmpty(t, opt.Reviewers)
					requested = append(requested, opt.Reviewers...)
					assert.Empty(t, opt.TeamReviewers)
					if tt.requestStatus != 0 {
						w.WriteHeader(tt.requestStatus)
						return
					}
					w.WriteHeader(http.StatusCreated)
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()

			client, err := giteasdk.NewClient(server.URL)
			require.NoError(t, err)

			g := Gitea{
				spec:       Spec{Reviewers: tt.reviewers},
				client:     client,
				Owner:      "updatecli",
				Repository: "updatecli",
			}
			pr := &giteasdk.PullRequest{
				ID:      100,
				Index:   42,
				URL:     "https://gitea.example.com/api/v1/repos/updatecli/updatecli/pulls/42",
				HTMLURL: "https://gitea.example.com/updatecli/updatecli/pulls/42",
			}

			err = g.requestReviewers(t.Context(), pr)
			if tt.expectedError != "" {
				require.ErrorContains(t, err, tt.expectedError+" for Gitea pull request "+strconv.Quote(pr.HTMLURL))
				assert.NotContains(t, err.Error(), pr.URL)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.expectedPages, pages)
			assert.Equal(t, tt.expectedReviewers, requested)
		})
	}
}
