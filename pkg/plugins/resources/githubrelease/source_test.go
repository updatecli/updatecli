package githubrelease

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

type mockGhHandler struct {
	github.Github
	releases   []github.ReleaseNode
	releaseErr error
	tags       []string
	tagErr     error
}

func (m *mockGhHandler) SearchReleases(releaseType github.ReleaseType) (releases []github.ReleaseNode, err error) {
	return m.releases, m.releaseErr
}

func (m *mockGhHandler) SearchTags() (releases []string, err error) {
	return m.tags, m.tagErr
}

func TestGitHubRelease_Source(t *testing.T) {
	tests := []struct {
		name            string
		workingDir      string
		releaseKey      string
		mockedGhHandler github.GithubHandler
		versionFilter   version.Filter
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// A version filter is required for all test cases
			require.NotNil(t, tt.versionFilter)

			gr := &GitHubRelease{
				ghHandler:     tt.mockedGhHandler,
				versionFilter: tt.versionFilter,
			}
			if tt.releaseKey != "" {
				gr.spec.Key = tt.releaseKey
			}

			gotResult := result.Source{}

			err := gr.Source(tt.workingDir, &gotResult)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantValue, gotResult.Information)
		})
	}
}
