package github

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchReleases(t *testing.T) {
	tests := []struct {
		name         string
		spec         Spec
		mockedQuery  *releasesQuery
		mockedError  error
		wantReleases []string
		wantErr      bool
	}{
		{
			name: "Nominal case with 4 kubernetes tags",
			spec: Spec{
				Owner:      "updatecli",
				Repository: "updatecli",
				Token:      "superSecretTokenForJoe",
				Username:   "joe",
			},
			mockedQuery: &releasesQuery{
				RateLimit: RateLimit{
					Cost:      1,
					Remaining: 5000,
				},
				Repository: struct {
					Releases repositoryRelease `graphql:"releases(last: 100, before: $before, orderBy: $orderBy)"`
				}{
					Releases: repositoryRelease{
						TotalCount: 4,
						PageInfo: PageInfo{
							HasNextPage: false,
							EndCursor:   "Y3Vyc29yOnYyOpK5MjAyMC0wMi0xOVQyMTozMDoxMyswMTowMM4BYOFb",
						},
						Edges: []releaseEdge{
							{
								Cursor: "Y3Vyc29yOnYyOpK5MjAyMi0wMS0xOFQxOTo0Nzo1MyswMTowMM4Da8vn",
								Node: releaseNode{
									Name:         "v0.18.4 🌈",
									TagName:      "v0.18.4",
									IsDraft:      true,
									IsPrerelease: false,
								},
							},
							{
								Cursor: "Y3Vyc29yOnYyOpK5MjAyMi0wMS0xOFQxOTo0Nzo1MyswMTowMM4Da8vn",
								Node: releaseNode{
									Name:         "v0.18.3",
									TagName:      "v0.18.3",
									IsDraft:      false,
									IsPrerelease: false,
								},
							},
							{
								Cursor: "Y3Vyc29yOnYyOpK5MjAyMS0xMS0xM1QxMjozNzoxMiswMTowMM4DK66W",
								Node: releaseNode{
									Name:         "v1.0.0",
									TagName:      "v1.0.0",
									IsDraft:      false,
									IsPrerelease: true,
								},
							},
							{
								Cursor: "Y3Vyc29yOnYyOpK5MjAyMC0wMi0xOVQyMTozMDoxMyswMTowMM4BYOFb",
								Node: releaseNode{
									Name:         "v0.0.1",
									TagName:      "v0.0.1",
									IsDraft:      false,
									IsPrerelease: false,
								},
							},
						},
					},
				},
			},
			wantReleases: []string{"v0.0.1", "v0.18.3"},
		},
		{
			name:        "Case with error returned from github query",
			mockedQuery: &releasesQuery{},
			mockedError: fmt.Errorf("Random error from GitHub API."),
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotNil(t, tt.mockedQuery)

			sut := Github{
				Spec: tt.spec,
				client: &MockGitHubClient{
					mockedQuery: tt.mockedQuery,
					mockedErr:   tt.mockedError,
				},
			}
			got, err := sut.SearchReleases()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			assert.Equal(t, tt.wantReleases, got)
		})
	}
}
