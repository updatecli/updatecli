package maven

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/mavenmetadata"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// unavailable is the daysAgo of a version whose pom file request fails with HTTP 403.
const unavailable = -2

// fakeRelease is a version served by fakeRepository, deployed daysAgo days ago.
// A daysAgo of -1 serves its pom file without Last-Modified header.
type fakeRelease struct {
	version string
	daysAgo int
}

// fakeRepository serves the maven-metadata.xml and pom files of the artifact org.example:lib.
// The pom requests are counted, and HEAD requests are rejected when headNotAllowed is set.
type fakeRepository struct {
	releases       []fakeRelease
	headNotAllowed bool
	pomRequests    atomic.Int64
}

func (f *fakeRepository) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const prefix = "/org/example/lib/"

	if r.URL.Path == prefix+"maven-metadata.xml" {
		var versions strings.Builder
		for _, rel := range f.releases {
			fmt.Fprintf(&versions, "<version>%s</version>", rel.version)
		}
		fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>
<metadata><groupId>org.example</groupId><artifactId>lib</artifactId><versioning><latest>%s</latest><release>%s</release><versions>%s</versions></versioning></metadata>`,
			f.releases[len(f.releases)-1].version, f.releases[len(f.releases)-1].version, versions.String())
		return
	}

	for _, rel := range f.releases {
		if r.URL.Path != fmt.Sprintf("%s%s/lib-%s.pom", prefix, rel.version, rel.version) {
			continue
		}

		f.pomRequests.Add(1)
		if f.headNotAllowed && r.Method == http.MethodHead {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if rel.daysAgo == unavailable {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if rel.daysAgo >= 0 {
			w.Header().Set("Last-Modified", time.Now().AddDate(0, 0, -rel.daysAgo).UTC().Format(http.TimeFormat))
		}
		return
	}

	w.WriteHeader(http.StatusNotFound)
}

// newAgeTestMaven returns a maven resource looking up the artifact org.example:lib on the fake repository.
func newAgeTestMaven(t *testing.T, repo *fakeRepository, filter version.Filter, releaseAge age.Spec) *Maven {
	t.Helper()

	server := httptest.NewServer(repo)
	t.Cleanup(server.Close)

	m, err := New(Spec{
		Repository:    server.URL,
		GroupID:       "org.example",
		ArtifactID:    "lib",
		VersionFilter: filter,
		Age:           releaseAge,
	})
	require.NoError(t, err)

	return m
}

func TestSourceAge(t *testing.T) {
	tests := []struct {
		name           string
		releases       []fakeRelease
		filter         version.Filter
		age            age.Spec
		headNotAllowed bool
		want           string
		wantResult     string
		wantErr        string
	}{
		{
			name:       "latest version old enough",
			releases:   []fakeRelease{{"1.0.0", 30}, {"1.1.0", 10}},
			age:        age.Spec{Minimum: "7d"},
			want:       "1.1.0",
			wantResult: result.SUCCESS,
		},
		{
			name:       "latest version too recent falls back to the most recent one deployed before it",
			releases:   []fakeRelease{{"1.0.0", 30}, {"1.1.0", 10}, {"1.2.0", 1}},
			age:        age.Spec{Minimum: "7d"},
			want:       "1.1.0",
			wantResult: result.SUCCESS,
		},
		{
			name:       "latest version without release date falls back to a dated one",
			releases:   []fakeRelease{{"1.0.0", 30}, {"1.1.0-SNAPSHOT", -1}},
			age:        age.Spec{Minimum: "7d"},
			want:       "1.0.0",
			wantResult: result.SUCCESS,
		},
		{
			name:       "semver filter skips the version too recent",
			releases:   []fakeRelease{{"1.1.0", 10}, {"2.0.0", 1}, {"1.0.0", 30}},
			filter:     version.Filter{Kind: "semver", Pattern: "*"},
			age:        age.Spec{Minimum: "7d"},
			want:       "1.1.0",
			wantResult: result.SUCCESS,
		},
		{
			name:       "maximum age discards older versions",
			releases:   []fakeRelease{{"1.0.0", 400}, {"1.1.0", 10}},
			filter:     version.Filter{Kind: "semver", Pattern: "~1.0"},
			age:        age.Spec{Maximum: "1y"},
			wantResult: result.SKIPPED,
		},
		{
			name:       "every version too recent skips the source",
			releases:   []fakeRelease{{"1.0.0", 2}, {"1.1.0", 1}},
			age:        age.Spec{Minimum: "7d"},
			wantResult: result.SKIPPED,
		},
		{
			name:       "repository without release date skips the source",
			releases:   []fakeRelease{{"1.0.0", -1}, {"1.1.0", -1}},
			age:        age.Spec{Minimum: "7d"},
			wantResult: result.SKIPPED,
		},
		{
			name:       "semver versions without release date are ignored",
			releases:   []fakeRelease{{"1.0.0", 30}, {"1.1.0", -1}},
			filter:     version.Filter{Kind: "semver", Pattern: "*"},
			age:        age.Spec{Minimum: "7d"},
			want:       "1.0.0",
			wantResult: result.SUCCESS,
		},
		{
			name:     "repository error doesn't fall back to an older version",
			releases: []fakeRelease{{"1.0.0", 30}, {"1.1.0", unavailable}},
			age:      age.Spec{Minimum: "7d"},
			wantErr:  "403",
		},
		{
			name:           "HEAD not allowed falls back to GET",
			releases:       []fakeRelease{{"1.0.0", 30}, {"1.1.0", 10}},
			age:            age.Spec{Minimum: "7d"},
			headNotAllowed: true,
			want:           "1.1.0",
			wantResult:     result.SUCCESS,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newAgeTestMaven(t, &fakeRepository{releases: tt.releases, headNotAllowed: tt.headNotAllowed}, tt.filter, tt.age)

			got := result.Source{}
			err := m.Source(context.Background(), pathresolver.Resolver{}, &got)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantResult, got.Result)
			assert.Equal(t, tt.want, got.Information)
		})
	}
}

func TestSourceAgeBoundsTheChecks(t *testing.T) {
	var releases []fakeRelease
	for i := range 2 * maxAgeChecks {
		releases = append(releases, fakeRelease{fmt.Sprintf("1.%d.0", i), 1})
	}
	repo := &fakeRepository{releases: releases}

	for _, filter := range []version.Filter{{}, {Kind: "semver", Pattern: "*"}} {
		repo.pomRequests.Store(0)
		m := newAgeTestMaven(t, repo, filter, age.Spec{Minimum: "7d"})

		got := result.Source{}
		require.NoError(t, m.Source(context.Background(), pathresolver.Resolver{}, &got))
		assert.Equal(t, result.SKIPPED, got.Result)
		assert.Equal(t, int64(maxAgeChecks), repo.pomRequests.Load(), "versionfilter kind %q", filter.Kind)
	}
}

func TestSourceAgeWithoutReleaseDateExplainsTheSkip(t *testing.T) {
	m := newAgeTestMaven(t, &fakeRepository{releases: []fakeRelease{{"1.0.0", -1}}}, version.Filter{}, age.Spec{Minimum: "7d"})

	got := result.Source{}
	require.NoError(t, m.Source(context.Background(), pathresolver.Resolver{}, &got))
	assert.Equal(t, result.SKIPPED, got.Result)
	assert.Contains(t, got.Description, "reports no release date")
}

func TestSourceAgeIgnoresRepositoriesWithoutTheArtifact(t *testing.T) {
	m := newAgeTestMaven(t, &fakeRepository{releases: []fakeRelease{{"1.0.0", 1}}}, version.Filter{}, age.Spec{Minimum: "7d"})
	m.metadataHandlers = append([]mavenmetadata.Handler{&mavenmetadata.MockMetadataHandler{Err: fmt.Errorf("private repository: %w", text.ErrNotFound)}}, m.metadataHandlers...)

	got := result.Source{}
	require.NoError(t, m.Source(context.Background(), pathresolver.Resolver{}, &got))
	assert.Equal(t, result.SKIPPED, got.Result)
}

func TestSourceAgeDoesNotHideOtherFailures(t *testing.T) {
	m := newAgeTestMaven(t, &fakeRepository{releases: []fakeRelease{{"1.0.0", 1}}}, version.Filter{}, age.Spec{Minimum: "7d"})
	m.metadataHandlers = append([]mavenmetadata.Handler{&mavenmetadata.MockMetadataHandler{Err: errors.New("401 Unauthorized")}}, m.metadataHandlers...)

	got := result.Source{}
	err := m.Source(context.Background(), pathresolver.Resolver{}, &got)
	require.ErrorContains(t, err, "401 Unauthorized")
	require.ErrorIs(t, err, age.ErrNoVersionMatchingAge)
}

func TestConditionAge(t *testing.T) {
	releases := []fakeRelease{{"1.0.0", 30}, {"1.1.0", 1}, {"1.2.0", -1}, {"1.3.0", unavailable}}

	tests := []struct {
		name    string
		version string
		want    bool
		wantErr string
	}{
		{name: "version old enough", version: "1.0.0", want: true},
		{name: "version too recent", version: "1.1.0", want: false},
		{name: "version not published", version: "2.0.0", want: false},
		{name: "version without release date", version: "1.2.0", want: false},
		{name: "repository error", version: "1.3.0", wantErr: "403"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newAgeTestMaven(t, &fakeRepository{releases: releases}, version.Filter{}, age.Spec{Minimum: "7d"})

			got, _, err := m.Condition(context.Background(), tt.version, nil, pathresolver.Resolver{})
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConditionAgeTriesEveryRepository(t *testing.T) {
	failing := &fakeRepository{releases: []fakeRelease{{"1.0.0", unavailable}}}
	m := newAgeTestMaven(t, &fakeRepository{releases: []fakeRelease{{"1.0.0", 30}}}, version.Filter{}, age.Spec{Minimum: "7d"})
	m.metadataHandlers = append(newAgeTestMaven(t, failing, version.Filter{}, age.Spec{Minimum: "7d"}).metadataHandlers, m.metadataHandlers...)

	got, _, err := m.Condition(context.Background(), "1.0.0", nil, pathresolver.Resolver{})
	require.NoError(t, err)
	assert.True(t, got)
	assert.Equal(t, int64(1), failing.pomRequests.Load())
}

func TestNewInvalidAge(t *testing.T) {
	_, err := New(Spec{GroupID: "org.example", ArtifactID: "lib", Age: age.Spec{Minimum: "soon"}})
	require.ErrorContains(t, err, "wrong age spec")
}

func TestOlderVersions(t *testing.T) {
	tests := []struct {
		name     string
		versions []string
		latest   string
		limit    int
		expected []string
	}{
		{
			name:     "backport deployed after a newer release",
			versions: []string{"2.16.0", "2.17.0", "2.12.4", "2.17.1-rc1", "2.17.1"},
			latest:   "2.17.1",
			limit:    10,
			expected: []string{"2.17.0", "2.16.0", "2.12.4", "2.17.1-rc1"},
		},
		{
			name:     "pre-releases kept in order when latest is one",
			versions: []string{"3.0.0-beta1", "2.24.0", "3.0.0-beta2", "3.0.0-beta3"},
			latest:   "3.0.0-beta3",
			limit:    10,
			expected: []string{"3.0.0-beta2", "3.0.0-beta1", "2.24.0"},
		},
		{
			name:     "limited",
			versions: []string{"1.0.0", "1.1.0", "1.2.0", "1.3.0"},
			latest:   "1.3.0",
			limit:    2,
			expected: []string{"1.2.0", "1.1.0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, olderVersions(tt.versions, tt.latest, tt.limit))
		})
	}
}
