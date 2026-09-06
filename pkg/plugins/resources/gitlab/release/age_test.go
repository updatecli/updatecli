package release

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

// newReleaseServer serves a single page of releases, so that the age filtering is
// exercised without reaching the real GitLab API.
func newReleaseServer(t *testing.T, releases ...string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "[%s]", strings.Join(releases, ","))
	}))
}

func releaseJSON(tag, released string, upcoming bool) string {
	return fmt.Sprintf(`{"tag_name":%q,"released_at":%q,"created_at":%q,"upcoming_release":%t}`,
		tag, released, released, upcoming)
}

func newReleaseResource(t *testing.T, url string, releaseAge age.Spec) *Gitlab {
	t.Helper()
	resource, err := New(map[string]interface{}{
		"url":        url,
		"owner":      "updatecli",
		"repository": "updatecli",
		"age":        map[string]interface{}{"minimum": releaseAge.Minimum, "maximum": releaseAge.Maximum},
	})
	require.NoError(t, err)
	return resource
}

func TestSearchReleasesAge(t *testing.T) {
	// GitLab returns the most recent release first and SearchReleases reverses that,
	// so the expected values below read oldest first.
	releases := []string{
		releaseJSON("v3.0.0", daysAgo(1), false),
		releaseJSON("v2.0.0", daysAgo(10), false),
		releaseJSON("v1.0.0", daysAgo(30), false),
	}

	tests := []struct {
		name       string
		releaseAge age.Spec
		releases   []string
		want       []string
		wantSkip   bool
	}{
		{
			name:     "no age filter keeps every release",
			releases: releases,
			want:     []string{"v1.0.0", "v2.0.0", "v3.0.0"},
		},
		{
			name:       "the most recent release is still in cooldown",
			releaseAge: age.Spec{Minimum: "7d"},
			releases:   releases,
			want:       []string{"v1.0.0", "v2.0.0"},
		},
		{
			name:       "releases published too long ago are discarded",
			releaseAge: age.Spec{Maximum: "20d"},
			releases:   releases,
			want:       []string{"v2.0.0", "v3.0.0"},
		},
		{
			name:       "every release still in cooldown reports a running cooldown",
			releaseAge: age.Spec{Minimum: "60d"},
			releases:   releases,
			wantSkip:   true,
		},
		{
			name:       "an upcoming release is not mistaken for a cooling down one",
			releaseAge: age.Spec{Minimum: "7d"},
			releases: []string{
				releaseJSON("v4.0.0", daysAgo(-1), true),
				releaseJSON("v1.0.0", daysAgo(30), false),
			},
			want: []string{"v1.0.0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newReleaseServer(t, tt.releases...)
			defer server.Close()

			got, err := newReleaseResource(t, server.URL, tt.releaseAge).SearchReleases(tt.releaseAge)

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

func TestReleaseDate(t *testing.T) {
	released := time.Now().Add(-24 * time.Hour)
	created := time.Now().Add(-48 * time.Hour)

	tests := []struct {
		name    string
		release gitlab.Release
		want    time.Time
		wantOk  bool
	}{
		{
			name:    "the release date wins",
			release: gitlab.Release{ReleasedAt: &released, CreatedAt: &created},
			want:    released,
			wantOk:  true,
		},
		{
			name:    "the creation date is the fallback",
			release: gitlab.Release{CreatedAt: &created},
			want:    created,
			wantOk:  true,
		},
		{
			name:    "a release without any date reports none",
			release: gitlab.Release{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := releaseDate(&tt.release)
			assert.Equal(t, tt.wantOk, ok)
			if tt.wantOk {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
