package branch

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// daysAgo formats a date n days in the past the way the GitLab API reports one.
func daysAgo(n int) string {
	return time.Now().Add(-time.Duration(n) * 24 * time.Hour).Format(time.RFC3339)
}

// newBranchServer serves a single page of branches, so that the age filtering is
// exercised without reaching the real GitLab API.
func newBranchServer(t *testing.T, branches ...string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "[%s]", strings.Join(branches, ","))
	}))
}

func branchJSON(name, committed string) string {
	return fmt.Sprintf(`{"name":%q,"commit":{"committed_date":%q}}`, name, committed)
}

func newBranchResource(t *testing.T, url string, branchAge age.Spec) *Gitlab {
	t.Helper()
	resource, err := New(map[string]interface{}{
		"url":        url,
		"owner":      "updatecli",
		"repository": "updatecli",
		"age":        map[string]interface{}{"minimum": branchAge.Minimum, "maximum": branchAge.Maximum},
	})
	require.NoError(t, err)
	return resource
}

func TestSearchBranchesAge(t *testing.T) {
	branches := []string{
		branchJSON("v1.0", daysAgo(30)),
		branchJSON("v2.0", daysAgo(10)),
		branchJSON("v3.0", daysAgo(1)),
	}

	tests := []struct {
		name      string
		branchAge age.Spec
		want      []string
		wantSkip  bool
	}{
		{
			name: "no age filter keeps every branch",
			want: []string{"v1.0", "v2.0", "v3.0"},
		},
		{
			name:      "the most recently updated branch is still in cooldown",
			branchAge: age.Spec{Minimum: "7d"},
			want:      []string{"v1.0", "v2.0"},
		},
		{
			name:      "branches updated too long ago are discarded",
			branchAge: age.Spec{Maximum: "20d"},
			want:      []string{"v2.0", "v3.0"},
		},
		{
			name:      "every branch still in cooldown reports a running cooldown",
			branchAge: age.Spec{Minimum: "60d"},
			wantSkip:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newBranchServer(t, branches...)
			defer server.Close()

			got, err := newBranchResource(t, server.URL, tt.branchAge).SearchBranches(tt.branchAge)

			if tt.wantSkip {
				require.Error(t, err)
				assert.ErrorIs(t, err, age.ErrNoVersionMatchingAge)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBranchDate(t *testing.T) {
	committed := time.Now().Add(-24 * time.Hour)
	authored := time.Now().Add(-48 * time.Hour)

	tests := []struct {
		name   string
		branch gitlab.Branch
		want   time.Time
		wantOk bool
	}{
		{
			name:   "the committer date wins",
			branch: gitlab.Branch{Commit: &gitlab.Commit{CommittedDate: &committed, AuthoredDate: &authored}},
			want:   committed,
			wantOk: true,
		},
		{
			name:   "the author date is the fallback",
			branch: gitlab.Branch{Commit: &gitlab.Commit{AuthoredDate: &authored}},
			want:   authored,
			wantOk: true,
		},
		{
			name:   "a branch without a commit reports no date",
			branch: gitlab.Branch{},
		},
		{
			name:   "a commit without a date reports none",
			branch: gitlab.Branch{Commit: &gitlab.Commit{}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := branchDate(&tt.branch)
			assert.Equal(t, tt.wantOk, ok)
			if tt.wantOk {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
