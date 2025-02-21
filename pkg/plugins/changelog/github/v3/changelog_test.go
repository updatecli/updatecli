package changelog

import (
	"os"
	"testing"

	"github.com/google/go-github/v69/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func TestSearch(t *testing.T) {
	testdata := []struct {
		name           string
		changelog      Changelog
		from           string
		to             string
		expectedResult result.Changelogs
	}{
		{
			name: "get all changelogs",
			changelog: Changelog{
				Owner:      "updatecli-test",
				Repository: "updatecli",
				VersionFilter: version.Filter{
					Kind: "semver",
				},
			},
			expectedResult: result.Changelogs{
				{Title: "v0.99.3"},
				{Title: "v0.99.2"},
				{Title: "v0.99.1"},
			},
		},
		{
			name: "get one changelog",
			changelog: Changelog{
				Owner:      "updatecli-test",
				Repository: "updatecli",
				VersionFilter: version.Filter{
					Kind: "semver",
				},
			},
			from: "v0.99.3",
			to:   "v0.99.3",
			expectedResult: result.Changelogs{
				{Title: "v0.99.3"},
			},
		},
		{
			name: "reverse changelog",
			changelog: Changelog{
				Owner:      "updatecli-test",
				Repository: "updatecli",
				VersionFilter: version.Filter{
					Kind: "semver",
				},
			},
			from: "v0.99.3",
			to:   "v0.99.2",
			expectedResult: result.Changelogs{
				{Title: "v0.99.3"},
			},
		},
		{
			name: "filter changelog",
			changelog: Changelog{
				Owner:      "updatecli-test",
				Repository: "updatecli",
				VersionFilter: version.Filter{
					Kind: "semver",
				},
			},
			from: "v0.99.2",
			to:   "v0.99.3",
			expectedResult: result.Changelogs{
				{Title: "v0.99.3"},
				{Title: "v0.99.2"},
			},
		},
		{
			name: "get filter changelog from cache",
			changelog: Changelog{
				Owner:      "updatecli-test",
				Repository: "updatecli",
				VersionFilter: version.Filter{
					Kind: "semver",
				},
			},
			from: "v0.99.2",
			to:   "v0.99.3",
			expectedResult: result.Changelogs{
				{Title: "v0.99.3"},
				{Title: "v0.99.2"},
			},
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {

			if os.Getenv("GITHUB_TOKEN") != "" {
				tt.changelog.Token = os.Getenv("GITHUB_TOKEN")
			}

			gotResults, err := tt.changelog.Search(tt.from, tt.to)
			require.NoError(t, err)

			require.Equal(t, len(tt.expectedResult), len(gotResults))

			for i := range tt.expectedResult {
				assert.Equal(t, tt.expectedResult[i].Title, gotResults[i].Title)
			}
		})
	}
}

func TestSortBySemver(t *testing.T) {
	testdata := []struct {
		name                   string
		releases               []*github.RepositoryRelease
		expectedSortedReleases []*github.RepositoryRelease
	}{
		{
			releases: []*github.RepositoryRelease{
				{TagName: github.Ptr("v0.4.0")},
				{TagName: github.Ptr("v0.5.0")},
				{TagName: github.Ptr("v0.3.0")},
			},
			expectedSortedReleases: []*github.RepositoryRelease{
				{TagName: github.Ptr("v0.5.0")},
				{TagName: github.Ptr("v0.4.0")},
				{TagName: github.Ptr("v0.3.0")},
			},
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			sortReleasesBySemver(&tt.releases)

			require.Equal(t, len(tt.expectedSortedReleases), len(tt.releases))

			for i := range tt.expectedSortedReleases {
				assert.Equal(t, *tt.expectedSortedReleases[i].TagName, *tt.releases[i].TagName)
			}
		})
	}
}
