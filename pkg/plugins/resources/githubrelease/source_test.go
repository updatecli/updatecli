package githubrelease

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/shurcooL/githubv4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// publishedDaysAgo builds the publication date of a release published n days ago.
func publishedDaysAgo(n int) githubv4.DateTime {
	return githubv4.DateTime{Time: time.Now().Add(-time.Duration(n) * 24 * time.Hour)}
}

type mockGhHandler struct {
	github.Github
	releases   []github.ReleaseNode
	releaseErr error
	tags       []string
	tagErr     error
}

func (m *mockGhHandler) SearchReleases(_ context.Context, releaseType github.ReleaseType, retry int) (releases []github.ReleaseNode, err error) {
	return m.releases, m.releaseErr
}

func (m *mockGhHandler) SearchTags(_ context.Context, retry int) (releases []string, err error) {
	return m.tags, m.tagErr
}

func (m *mockGhHandler) SearchReleasesByTagName(_ context.Context, releaseType github.ReleaseType) (releases []string, err error) {
	var result []string
	for _, r := range m.releases {
		result = append(result, r.TagName)
	}
	return result, m.releaseErr
}

func (m *mockGhHandler) SearchReleasesByTagHash(_ context.Context, releaseType github.ReleaseType) (releases []string, err error) {
	var result []string
	for _, r := range m.releases {
		result = append(result, r.TagCommit.Oid)
	}
	return result, m.releaseErr
}

func (m *mockGhHandler) SearchReleasesByTitle(_ context.Context, releaseType github.ReleaseType) (releases []string, err error) {
	var result []string
	for _, r := range m.releases {
		result = append(result, r.Name)
	}
	return result, m.releaseErr
}

func TestGitHubRelease_Source(t *testing.T) {
	tests := []struct {
		name            string
		workingDir      string
		releaseKey      string
		mockedGhHandler github.GithubHandler
		versionFilter   version.Filter
		releaseAge      age.Spec
		wantValue       string
		wantErr         bool
	}{
		{
			name: "3 releases found, filter with latest",
			mockedGhHandler: &mockGhHandler{
				releases: []github.ReleaseNode{{TagName: "1.0.0"}, {TagName: "2.0.0"}, {TagName: "3.0.0"}},
			},
			versionFilter: version.Filter{
				Kind:    "latest",
				Pattern: "latest",
			},
			wantValue: "3.0.0",
		},
		{
			name: "3 releases found, filter with latest, get hash",
			mockedGhHandler: &mockGhHandler{
				releases: []github.ReleaseNode{
					{TagName: "1.0.0", TagCommit: github.TagCommit{Oid: "11111111"}},
					{TagName: "2.0.0", TagCommit: github.TagCommit{Oid: "22222222"}},
					{TagName: "3.0.0", TagCommit: github.TagCommit{Oid: "33333333"}},
				},
			},
			releaseKey: "hash",
			versionFilter: version.Filter{
				Kind:    "latest",
				Pattern: "latest",
			},
			wantValue: "33333333",
		},
		{
			name: "0 releases found, 3 tags found, filter with latest",
			mockedGhHandler: &mockGhHandler{
				tags: []string{"1.0.0", "2.0.0", "3.0.0"},
			},
			versionFilter: version.Filter{
				Kind:    "latest",
				Pattern: "latest",
			},
			wantValue: "3.0.0",
		},
		{
			name:            "Error: 0 releases found, O tags found, filter with latest",
			mockedGhHandler: &mockGhHandler{},
			versionFilter: version.Filter{
				Kind:    "latest",
				Pattern: "latest",
			},
			wantErr: true,
		},
		{
			name: "Error: 3 releases found, filter with semver on 2.1.y",
			mockedGhHandler: &mockGhHandler{
				releases: []github.ReleaseNode{{TagName: "1.0.0"}, {TagName: "2.0.0"}, {TagName: "3.0.0"}},
			},
			versionFilter: version.Filter{
				Kind:    "semver",
				Pattern: "~2.1",
			},
			wantErr: true,
		},
		{
			name: "Error: error while retrieving releases",
			mockedGhHandler: &mockGhHandler{
				releaseErr: fmt.Errorf("Unexpected error while retrieving releases from GitHub."),
			},
			versionFilter: version.Filter{
				Kind:    "latest",
				Pattern: "latest",
			},
			wantErr: true,
		},
		{
			name: "Error: error while retrieving releases",
			mockedGhHandler: &mockGhHandler{
				tagErr: fmt.Errorf("Unexpected error while retrieving tags from GitHub."),
			},
			versionFilter: version.Filter{
				Kind:    "latest",
				Pattern: "latest",
			},
			wantErr: true,
		},
		{
			name: "3 releases found, the most recent one is still in cooldown",
			mockedGhHandler: &mockGhHandler{
				releases: []github.ReleaseNode{
					{TagName: "1.0.0", PublishedAt: publishedDaysAgo(30)},
					{TagName: "2.0.0", PublishedAt: publishedDaysAgo(10)},
					{TagName: "3.0.0", PublishedAt: publishedDaysAgo(1)},
				},
			},
			versionFilter: version.Filter{
				Kind:    "latest",
				Pattern: "latest",
			},
			releaseAge: age.Spec{Minimum: "7d"},
			wantValue:  "2.0.0",
		},
		{
			name: "3 releases found, only the oldest one is not expired",
			mockedGhHandler: &mockGhHandler{
				releases: []github.ReleaseNode{
					{TagName: "1.0.0", PublishedAt: publishedDaysAgo(30)},
					{TagName: "2.0.0", PublishedAt: publishedDaysAgo(10)},
					{TagName: "3.0.0", PublishedAt: publishedDaysAgo(1)},
				},
			},
			versionFilter: version.Filter{
				Kind:    "latest",
				Pattern: "latest",
			},
			releaseAge: age.Spec{Minimum: "20d"},
			wantValue:  "1.0.0",
		},
		{
			name: "Draft releases fall back on their creation date",
			mockedGhHandler: &mockGhHandler{
				releases: []github.ReleaseNode{
					{TagName: "1.0.0", CreatedAt: publishedDaysAgo(30)},
					{TagName: "2.0.0", CreatedAt: publishedDaysAgo(1)},
				},
			},
			versionFilter: version.Filter{
				Kind:    "latest",
				Pattern: "latest",
			},
			releaseAge: age.Spec{Minimum: "7d"},
			wantValue:  "1.0.0",
		},
		{
			name: "Error: every release is still in cooldown",
			mockedGhHandler: &mockGhHandler{
				releases: []github.ReleaseNode{
					{TagName: "1.0.0", PublishedAt: publishedDaysAgo(2)},
					{TagName: "2.0.0", PublishedAt: publishedDaysAgo(1)},
				},
			},
			versionFilter: version.Filter{
				Kind:    "latest",
				Pattern: "latest",
			},
			releaseAge: age.Spec{Minimum: "7d"},
			wantErr:    true,
		},
		{
			name: "Error: the git tag fallback cannot honor an age filter",
			mockedGhHandler: &mockGhHandler{
				tags: []string{"1.0.0", "2.0.0", "3.0.0"},
			},
			versionFilter: version.Filter{
				Kind:    "latest",
				Pattern: "latest",
			},
			releaseAge: age.Spec{Minimum: "7d"},
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// A version filter is required for all test cases
			require.NotNil(t, tt.versionFilter)

			gr, err := New(Spec{
				Owner:         "owner",
				Repository:    "repository",
				Token:         "ghp_example",
				Key:           tt.releaseKey,
				VersionFilter: tt.versionFilter,
				Age:           tt.releaseAge,
			})

			require.NoError(t, err)

			gr.ghHandler = tt.mockedGhHandler

			gotResult := result.Source{}

			err = gr.Source(context.Background(), tt.workingDir, &gotResult)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantValue, gotResult.Information)
		})
	}
}
