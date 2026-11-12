package release

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"testing"
	"time"

	"github.com/drone/go-scm/scm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func TestChangelogRanges(t *testing.T) {
	date := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	releases := []*scm.Release{
		{Tag: "v3.0.0", Published: date.Add(48 * time.Hour)},
		{Tag: "v2.0.0", Published: date.Add(24 * time.Hour)},
		{Tag: "v1.0.0", Published: date},
	}
	tests := []struct {
		name     string
		from     string
		to       string
		filter   version.Filter
		releases []*scm.Release
		want     []string
	}{
		{name: "inclusive range", from: "v1.0.0", to: "v3.0.0", want: []string{"v3.0.0", "v2.0.0", "v1.0.0"}},
		{name: "same release", from: "v2.0.0", to: "v2.0.0", want: []string{"v2.0.0"}},
		{name: "downgrade", from: "v3.0.0", to: "v1.0.0", want: []string{"v3.0.0"}},
		{name: "missing old release", from: "v0.0.0", to: "v2.0.0", want: []string{"v2.0.0"}},
		{name: "missing new release", from: "v1.0.0", to: "v4.0.0"},
		{name: "no old version", to: "v2.0.0", want: []string{"v2.0.0", "v1.0.0"}},
		{name: "no new version", from: "v2.0.0", want: []string{"v2.0.0"}},
		{name: "no new version and missing old release", from: "v9.9.9"},
		{name: "lex without new version", from: "v1.0.0", filter: version.Filter{Kind: "lex"}},
		{name: "all releases", want: []string{"v3.0.0", "v2.0.0", "v1.0.0"}},
		{name: "no releases", from: "v1.0.0", to: "v2.0.0", releases: []*scm.Release{}},
		{
			name:   "semantic versions override publication order",
			from:   "v1.0.0",
			to:     "v3.0.0",
			filter: version.Filter{Kind: "latest", Strict: true},
			releases: []*scm.Release{
				{Tag: "v1.0.0", Published: date.Add(48 * time.Hour)},
				{Tag: "nightly", Published: date.Add(24 * time.Hour)},
				{Tag: "v3.0.0", Published: date},
			},
			want: []string{"v3.0.0", "v1.0.0"},
		},
		{
			name:     "explicit semver with no bounds",
			filter:   version.Filter{Kind: "semver"},
			releases: []*scm.Release{{Tag: "v1"}, {Tag: "nightly"}, {Tag: "v2"}},
			want:     []string{"v2", "v1"},
		},
		{
			name:     "explicit semver honors strict parsing",
			from:     "1.0.0",
			to:       "2.0.0",
			filter:   version.Filter{Kind: "semver", Strict: true},
			releases: []*scm.Release{{Tag: "2.0.0"}, {Tag: "1.5"}, {Tag: "1.0.0"}},
			want:     []string{"2.0.0", "1.0.0"},
		},
		{
			name:   "lex returns the selected release",
			from:   "release-a",
			to:     "release-c",
			filter: version.Filter{Kind: "lex"},
			releases: []*scm.Release{
				{Tag: "release-a", Published: date.Add(48 * time.Hour)},
				{Tag: "release-b", Published: date.Add(24 * time.Hour)},
				{Tag: "release-c", Published: date},
			},
			want: []string{"release-c"},
		},
		{
			name:   "date tags keep API order",
			from:   "2026-01-01",
			to:     "2026-02-01",
			filter: version.Filter{Kind: "latest"},
			releases: []*scm.Release{
				{Tag: "2026-02-01", Created: date.Add(24 * time.Hour), Published: date.Add(24 * time.Hour)},
				{Tag: "2026-01-01", Created: date, Published: date.Add(48 * time.Hour)},
			},
			want: []string{"2026-02-01", "2026-01-01"},
		},
		{
			name:   "regex keeps API order",
			from:   "v3.0.0",
			to:     "v1.0.0",
			filter: version.Filter{Kind: "regex"},
			releases: []*scm.Release{
				{Tag: "v1.0.0", Published: date},
				{Tag: "nightly", Published: date.Add(24 * time.Hour)},
				{Tag: "v3.0.0", Published: date.Add(48 * time.Hour)},
			},
			want: []string{"v1.0.0", "nightly", "v3.0.0"},
		},
		{
			name: "regex semver overrides publication order",
			from: "release-v1.0.0",
			to:   "release-v2.0.0",
			filter: version.Filter{
				Kind:  "regex/semver",
				Regex: `^release-v(\d+\.\d+\.\d+)$`,
			},
			releases: []*scm.Release{
				{Tag: "release-v1.0.0", Published: date.Add(24 * time.Hour)},
				{Tag: "release-v2.0.0", Published: date},
			},
			want: []string{"release-v2.0.0", "release-v1.0.0"},
		},
		{
			name: "regex semver applies replacement and strict parsing",
			from: "release-v1_0_0",
			to:   "release-v2_0_0",
			filter: version.Filter{
				Kind:       "regex/semver",
				Regex:      `^release-v([^_]+)$`,
				ReplaceAll: version.ReplaceAll{Pattern: "_", Replacement: "."},
				Strict:     true,
				Pattern:    ">=2.0.0",
			},
			releases: []*scm.Release{
				{Tag: "release-v1_0_0", Published: date.Add(48 * time.Hour)},
				{Tag: "1_8_0", Published: date.Add(24 * time.Hour)},
				{Tag: "release-v1_9", Published: date.Add(24 * time.Hour)},
				{Tag: "release-vnightly", Published: date.Add(24 * time.Hour)},
				{Tag: "release-v1_5_0", Published: date.Add(24 * time.Hour)},
				{Tag: "release-v2_0_0-rc_1", Published: date},
				{Tag: "release-v2_0_0", Published: date},
			},
			want: []string{"release-v2_0_0", "release-v2_0_0-rc_1", "release-v1_5_0", "release-v1_0_0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := tt.releases
			if input == nil {
				input = releases
			}
			g := &Gitea{versionFilter: tt.filter}
			got := g.selectReleases(slices.Clone(input), tt.from, tt.to)
			var tags []string
			for _, release := range got {
				tags = append(tags, release.Tag)
			}
			assert.Equal(t, tt.want, tags)
		})
	}
}

func TestChangelogOrder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/repos/updatecli/project/releases", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		// The API orders by creation date, even if an older release is published later.
		fmt.Fprint(w, `[
			{"tag_name":"release-2","created_at":"2026-01-02T00:00:00Z","published_at":"2026-01-02T00:00:00Z"},
			{"tag_name":"release-1","created_at":"2026-01-01T00:00:00Z","published_at":"2026-01-03T00:00:00Z"}
		]`)
	}))
	defer server.Close()

	tests := []struct {
		name   string
		filter version.Filter
	}{
		{name: "default latest"},
		{name: "regex", filter: version.Filter{Kind: "regex", Pattern: "^release-"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := map[string]interface{}{
				"url":           server.URL,
				"owner":         "updatecli",
				"repository":    "project",
				"token":         "test-token",
				"versionfilter": tt.filter,
			}
			g, err := New(spec)
			require.NoError(t, err)
			var source result.Source
			require.NoError(t, g.Source(t.Context(), pathresolver.Resolver{}, &source))
			require.Equal(t, "release-2", source.Information)

			g, err = New(spec)
			require.NoError(t, err)
			got := g.Changelog("release-1", source.Information)
			require.NotNil(t, got)
			require.Len(t, *got, 2)
			assert.Equal(t, "release-2", (*got)[0].Title)
			assert.Equal(t, "release-1", (*got)[1].Title)
		})
	}
}

func TestChangelogPagination(t *testing.T) {
	var pages []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/forge/api/v1/repos/updatecli/project/releases", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		page := r.URL.Query().Get("page")
		pages = append(pages, page)
		w.Header().Set("Content-Type", "application/json")
		base := "http://" + r.Host + "/forge"
		switch page {
		case "1":
			// A next link is sufficient; a last link is not always provided.
			w.Header().Set("Link", fmt.Sprintf(`<%s/api/v1/repos/updatecli/project/releases?page=2>; rel="next"`, base))
			fmt.Fprintf(w, `[
				{"tag_name":"v4.0.0","draft":true,"published_at":"2026-01-02T12:00:00Z"},
				{"tag_name":"v3.0.0","name":"Third release","body":"### Changes\nNew feature",
				 "url":%q,"html_url":%q,"created_at":"2026-01-02T00:00:00Z","published_at":"2026-01-03T00:00:00Z"}
			]`, base+"/api/v1/repos/updatecli/project/releases/3", base+"/updatecli/project/releases/tag/v3.0.0")
		case "2":
			fmt.Fprintf(w, `[
				{"tag_name":"v2.0.0-rc.1","prerelease":true,"body":"Preview","published_at":"2026-01-02T00:00:00Z",
				 "url":%q},
				{"tag_name":"release/1.0+build","body":"First release","created_at":"2026-01-01T00:00:00Z"}
			]`, "http://"+r.Host+"/api/v1/repos/updatecli/project/2")
		default:
			http.Error(w, "unexpected page", http.StatusBadRequest)
		}
	}))
	defer server.Close()

	baseURL, err := url.Parse(server.URL + "/forge/")
	require.NoError(t, err)
	baseURL.User = url.UserPassword("user", "password")
	g, err := New(map[string]interface{}{
		"url":        baseURL.String(),
		"owner":      "updatecli",
		"repository": "project",
		"token":      "test-token",
		"age":        map[string]interface{}{"minimum": "100y"},
	})
	require.NoError(t, err)

	got := g.Changelog("release/1.0+build", "v3.0.0")
	require.NotNil(t, got)
	assert.Equal(t, result.Changelogs{
		{
			Title:       "v3.0.0",
			Body:        "### Changes\nNew feature",
			PublishedAt: "2026-01-03 00:00:00 +0000 UTC",
			URL:         server.URL + "/forge/updatecli/project/releases/tag/v3.0.0",
		},
		{
			Title:       "v2.0.0-rc.1",
			Body:        "Preview",
			PublishedAt: "2026-01-02 00:00:00 +0000 UTC",
			URL:         server.URL + "/updatecli/project/releases/tag/v2.0.0-rc.1",
		},
		{
			Title:       "release/1.0+build",
			Body:        "First release",
			PublishedAt: "2026-01-01 00:00:00 +0000 UTC",
			URL:         server.URL + "/forge/updatecli/project/releases/tag/release%2F1.0+build",
		},
	}, *got)
	assert.Equal(t, []string{"1", "2"}, pages)

	// The age filter affects source selection, not the intervening release notes.
	var source result.Source
	require.NoError(t, g.Source(t.Context(), pathresolver.Resolver{}, &source))
	assert.Equal(t, result.SKIPPED, source.Result)
}

func TestChangelogErrors(t *testing.T) {
	tests := []struct {
		name string
		page string
		body string
		code int
	}{
		{name: "first page fails", page: "1", body: `{"message":"not found"}`, code: http.StatusNotFound},
		{name: "later page fails", page: "2", body: `{"message":"not found"}`, code: http.StatusNotFound},
		{name: "malformed later page", page: "2", body: `{`, code: http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.URL.Query().Get("page") == tt.page {
					w.WriteHeader(tt.code)
					fmt.Fprint(w, tt.body)
					return
				}
				w.Header().Set("Link", fmt.Sprintf(`<http://%s/api/v1/repos/updatecli/project/releases?page=2>; rel="next"`, r.Host))
				fmt.Fprint(w, `[{"tag_name":"v2.0.0","body":"Partial notes"}]`)
			}))
			defer server.Close()

			g, err := New(map[string]interface{}{
				"url":        server.URL,
				"owner":      "updatecli",
				"repository": "project",
				"token":      "test-token",
			})
			require.NoError(t, err)
			assert.Nil(t, g.Changelog("v1.0.0", "v2.0.0"))
		})
	}
}
