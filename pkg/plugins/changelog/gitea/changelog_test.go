package gitea

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/drone/go-scm/scm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/resources/gitea/client"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// newChangelog returns a changelog for updatecli/<repository> on a test server.
func newChangelog(serverURL, repository string) *Changelog {
	return &Changelog{URL: serverURL, Owner: "updatecli", Repository: repository}
}

// search runs Search and fails the test on error.
func search(t *testing.T, c *Changelog, from, to string) result.Changelogs {
	t.Helper()
	got, err := c.Search(from, to)
	require.NoError(t, err)
	return got
}

// listReleases lists updatecli/project releases from a test server.
func listReleases(t *testing.T, serverURL string) ([]*scm.Release, error) {
	t.Helper()
	c, err := client.New(client.Spec{URL: serverURL})
	require.NoError(t, err)
	return ListReleases(t.Context(), c, "updatecli", "project")
}

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
		{name: "no old version", to: "v2.0.0", want: []string{"v2.0.0"}},
		{name: "no new version", from: "v2.0.0", want: []string{"v2.0.0"}},
		{name: "no new version and missing old release", from: "v9.9.9"},
		{name: "lex without new version", from: "v1.0.0", filter: version.Filter{Kind: "lex"}},
		{name: "no versions"},
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
			name:     "explicit semver sorts and skips non semver tags",
			from:     "v1",
			to:       "v2",
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
			got := selectReleases(slices.Clone(input), tt.filter, tt.from, tt.to)
			var tags []string
			for _, release := range got {
				tags = append(tags, release.Tag)
			}
			assert.Equal(t, tt.want, tags)
		})
	}
}

// resetCatalog clears the release cache before and after a test.
func resetCatalog(t *testing.T) {
	t.Helper()
	reset := func() {
		catalogMu.Lock()
		defer catalogMu.Unlock()
		catalog = nil
	}
	reset()
	t.Cleanup(reset)
}

func TestChangelogCache(t *testing.T) {
	resetCatalog(t)
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[
			{"tag_name":"v2.0.0","body":"Second"},
			{"tag_name":"v1.0.0","body":"First"}
		]`)
	}))
	defer server.Close()

	for range 3 {
		got := search(t, newChangelog(server.URL, "project"), "v1.0.0", "v2.0.0")
		require.Len(t, got, 2)
	}
	assert.Equal(t, int32(1), requests.Load())

	// Another repository on the same server is fetched separately.
	require.Len(t, search(t, newChangelog(server.URL, "other"), "v1.0.0", "v2.0.0"), 2)
	assert.Equal(t, int32(2), requests.Load())
}

func TestListReleasesStuckPagination(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.Header().Set("Content-Type", "application/json")
		// A next link pointing back at the current page must not loop forever.
		w.Header().Set("Link", fmt.Sprintf(`<http://%s/api/v1/repos/updatecli/project/releases?page=1>; rel="next"`, r.Host))
		fmt.Fprint(w, `[{"tag_name":"v1.0.0"}]`)
	}))
	defer server.Close()

	_, err := listReleases(t, server.URL)
	require.Error(t, err)
	assert.Equal(t, int32(1), requests.Load())
}

func TestChangelogCacheRefreshesMissingRelease(t *testing.T) {
	resetCatalog(t)
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if requests.Add(1) == 1 {
			fmt.Fprint(w, `[{"tag_name":"v2.0.0"},{"tag_name":"v1.0.0"}]`)
			return
		}
		// v3.0.0 was published after the first request filled the cache.
		fmt.Fprint(w, `[{"tag_name":"v3.0.0"},{"tag_name":"v2.0.0"},{"tag_name":"v1.0.0"}]`)
	}))
	defer server.Close()

	c := newChangelog(server.URL, "project")
	require.Len(t, search(t, c, "v1.0.0", "v2.0.0"), 2)
	got := search(t, c, "v2.0.0", "v3.0.0")
	require.Len(t, got, 2)
	assert.Equal(t, "v3.0.0", got[0].Title)
	assert.Equal(t, int32(2), requests.Load())

	// The refreshed list now serves v3.0.0 from the cache.
	require.Len(t, search(t, c, "v1.0.0", "v3.0.0"), 3)
	assert.Equal(t, int32(2), requests.Load())
}

func TestChangelogCacheLocksPerRepository(t *testing.T) {
	resetCatalog(t)
	release := make(chan struct{})
	slowStarted := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/slow/") {
			close(slowStarted)
			<-release
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[{"tag_name":"v2.0.0"},{"tag_name":"v1.0.0"}]`)
	}))
	defer server.Close()
	defer close(release)

	slow, fast := newChangelog(server.URL, "slow"), newChangelog(server.URL, "fast")

	go func() { _, _ = slow.Search("v1.0.0", "v2.0.0") }()
	// Wait until the slow fetch holds its repository lock.
	<-slowStarted

	done := make(chan result.Changelogs)
	go func() {
		got, _ := fast.Search("v1.0.0", "v2.0.0")
		done <- got
	}()
	select {
	case got := <-done:
		assert.Len(t, got, 2)
	case <-time.After(5 * time.Second):
		t.Fatal("a slow fetch for one repository blocked another repository")
	}
}

func TestListReleasesStopsOnEmptyPage(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		page, err := strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil {
			// require must not be used outside the test goroutine.
			http.Error(w, "invalid page", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		// Every response links to a further page, even once the releases run out.
		w.Header().Set("Link", fmt.Sprintf(`<http://%s/api/v1/repos/updatecli/project/releases?page=%d>; rel="next"`, r.Host, page+1))
		if page == 1 {
			fmt.Fprint(w, `[{"tag_name":"v1.0.0"}]`)
			return
		}
		fmt.Fprint(w, `[]`)
	}))
	defer server.Close()

	releases, err := listReleases(t, server.URL)
	require.NoError(t, err)
	require.Len(t, releases, 1)
	assert.Equal(t, int32(2), requests.Load())
}

func TestChangelogCacheRemembersMissingRelease(t *testing.T) {
	resetCatalog(t)
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[{"tag_name":"v2.0.0"},{"tag_name":"v1.0.0"}]`)
	}))
	defer server.Close()

	c := newChangelog(server.URL, "project")
	// Registries often drop the "v" prefix, so callers try "2.0.0" before "v2.0.0".
	for range 3 {
		assert.Empty(t, search(t, c, "1.0.0", "2.0.0"))
		assert.Len(t, search(t, c, "v1.0.0", "v2.0.0"), 2)
	}
	assert.Equal(t, int32(1), requests.Load())

	// A different missing tag still refreshes the list once.
	assert.Empty(t, search(t, c, "v2.0.0", "v3.0.0"))
	assert.Empty(t, search(t, c, "v2.0.0", "v3.0.0"))
	assert.Equal(t, int32(2), requests.Load())
}

func TestChangelogCacheRemembersSeveralMissingReleases(t *testing.T) {
	resetCatalog(t)
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[{"tag_name":"v2.0.0"},{"tag_name":"v1.0.0"}]`)
	}))
	defer server.Close()

	c := newChangelog(server.URL, "project")
	// The version exists on the registry but has no forge release, with or without "v".
	for range 3 {
		assert.Empty(t, search(t, c, "1.0.0", "3.0.0"))
		assert.Empty(t, search(t, c, "v1.0.0", "v3.0.0"))
	}
	// One fetch to fill the cache, then one refresh for "v3.0.0".
	assert.Equal(t, int32(2), requests.Load())
}

func TestChangelogCacheSeparatesCredentials(t *testing.T) {
	resetCatalog(t)
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.Header().Set("Content-Type", "application/json")
		if r.Header.Get("Authorization") == "" {
			// The repository is private: an anonymous caller sees nothing.
			http.Error(w, `{"message":"not found"}`, http.StatusNotFound)
			return
		}
		fmt.Fprint(w, `[{"tag_name":"v2.0.0","body":"Private notes"},{"tag_name":"v1.0.0"}]`)
	}))
	defer server.Close()

	authenticated := newChangelog(server.URL, "project")
	authenticated.Token = "secret"
	require.Len(t, search(t, authenticated, "v1.0.0", "v2.0.0"), 2)

	anonymous := newChangelog(server.URL+"/", "project")
	_, err := anonymous.Search("v1.0.0", "v2.0.0")
	require.Error(t, err)
	assert.Equal(t, int32(2), requests.Load())
}

func TestChangelogCacheNormalisesURL(t *testing.T) {
	resetCatalog(t)
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[{"tag_name":"v2.0.0"},{"tag_name":"v1.0.0"}]`)
	}))
	defer server.Close()

	require.Len(t, search(t, newChangelog(server.URL, "project"), "v1.0.0", "v2.0.0"), 2)
	require.Len(t, search(t, newChangelog(server.URL+"/", "project"), "v1.0.0", "v2.0.0"), 2)
	assert.Equal(t, int32(1), requests.Load())
}

func TestChangelogCacheRemembersFailure(t *testing.T) {
	resetCatalog(t)
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		http.Error(w, `{"message":"not found"}`, http.StatusNotFound)
	}))
	defer server.Close()

	c := newChangelog(server.URL, "deleted")
	for range 3 {
		_, err := c.Search("1.0.0", "2.0.0")
		require.Error(t, err)
	}
	assert.Equal(t, int32(1), requests.Load())
}

func TestChangelogCacheKeepsListWhenRefreshFails(t *testing.T) {
	resetCatalog(t)
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if requests.Add(1) > 1 {
			http.Error(w, `{"message":"not found"}`, http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[{"tag_name":"v2.0.0"},{"tag_name":"v1.0.0"}]`)
	}))
	defer server.Close()

	c := newChangelog(server.URL, "project")
	require.Len(t, search(t, c, "v1.0.0", "v2.0.0"), 2)
	// "v3.0.0" triggers a refresh that fails: the cached list stays usable.
	assert.Empty(t, search(t, c, "v2.0.0", "v3.0.0"))
	require.Len(t, search(t, c, "v1.0.0", "v2.0.0"), 2)
	assert.Empty(t, search(t, c, "v2.0.0", "v3.0.0"))
	assert.Equal(t, int32(2), requests.Load())
}
