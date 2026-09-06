package tag

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

// newTagServer serves a single page of tags, so that the age filtering is exercised
// without reaching the real GitLab API.
func newTagServer(t *testing.T, tags ...string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "[%s]", strings.Join(tags, ","))
	}))
}

func tagJSON(name, created string) string {
	return fmt.Sprintf(`{"name":%q,"created_at":%q,"commit":{"committed_date":%q}}`, name, created, created)
}

func newTagResource(t *testing.T, url string, tagAge age.Spec) *Gitlab {
	t.Helper()
	resource, err := New(map[string]interface{}{
		"url":        url,
		"owner":      "updatecli",
		"repository": "updatecli",
		"age":        map[string]interface{}{"minimum": tagAge.Minimum, "maximum": tagAge.Maximum},
	})
	require.NoError(t, err)
	return resource
}

func TestSearchTagsAge(t *testing.T) {
	tags := []string{
		tagJSON("v1.0.0", daysAgo(30)),
		tagJSON("v2.0.0", daysAgo(10)),
		tagJSON("v3.0.0", daysAgo(1)),
	}

	tests := []struct {
		name     string
		tagAge   age.Spec
		want     []string
		wantSkip bool
	}{
		{
			name: "no age filter keeps every tag",
			want: []string{"v1.0.0", "v2.0.0", "v3.0.0"},
		},
		{
			name:   "the most recent tag is still in cooldown",
			tagAge: age.Spec{Minimum: "7d"},
			want:   []string{"v1.0.0", "v2.0.0"},
		},
		{
			name:   "tags created too long ago are discarded",
			tagAge: age.Spec{Maximum: "20d"},
			want:   []string{"v2.0.0", "v3.0.0"},
		},
		{
			name:     "every tag still in cooldown reports a running cooldown",
			tagAge:   age.Spec{Minimum: "60d"},
			wantSkip: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newTagServer(t, tags...)
			defer server.Close()

			got, err := newTagResource(t, server.URL, tt.tagAge).SearchTags(tt.tagAge)

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

func TestTagDate(t *testing.T) {
	created := time.Now().Add(-24 * time.Hour)
	committed := time.Now().Add(-48 * time.Hour)

	tests := []struct {
		name   string
		tag    gitlab.Tag
		want   time.Time
		wantOk bool
	}{
		{
			name:   "the creation date wins",
			tag:    gitlab.Tag{CreatedAt: &created, Commit: &gitlab.Commit{CommittedDate: &committed}},
			want:   created,
			wantOk: true,
		},
		{
			name:   "the committer date is the fallback",
			tag:    gitlab.Tag{Commit: &gitlab.Commit{CommittedDate: &committed}},
			want:   committed,
			wantOk: true,
		},
		{
			name: "a tag without any date reports none",
			tag:  gitlab.Tag{},
		},
		{
			name: "a commit without a date reports none",
			tag:  gitlab.Tag{Commit: &gitlab.Commit{}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := tagDate(&tt.tag)
			assert.Equal(t, tt.wantOk, ok)
			if tt.wantOk {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
