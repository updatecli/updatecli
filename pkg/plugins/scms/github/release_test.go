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
		releaseType  ReleaseType
		mockedQuery  *releasesQuery
		mockedError  error
		wantReleases []string
		wantErr      bool
	}{
		{
			name: "Case with a mix of releases (draft, prerelease, release) and default ReleaseType filter",
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
									Name:    "v0.18.4 🌈",
									TagName: "v0.18.4",
									IsDraft: true,
								},
							},
							{
								Cursor: "Y3Vyc29yOnYyOpK5MjAyMi0wMS0xOFQxOTo0Nzo1MyswMTowMM4Da8vn",
								Node: releaseNode{
									Name:     "v0.18.3",
									TagName:  "v0.18.3",
									IsLatest: true,
								},
							},
							{
								Cursor: "Y3Vyc29yOnYyOpK5MjAyMS0xMS0xM1QxMjozNzoxMiswMTowMM4DK66W",
								Node: releaseNode{
									Name:         "v1.0.0",
									TagName:      "v1.0.0",
									IsPrerelease: true,
								},
							},
							{
								Cursor: "Y3Vyc29yOnYyOpK5MjAyMC0wMi0xOVQyMTozMDoxMyswMTowMM4BYOFb",
								Node: releaseNode{
									Name:    "v0.0.1",
									TagName: "v0.0.1",
								},
							},
						},
					},
				},
			},
			wantReleases: []string{"v0.0.1", "v0.18.3"},
		},
		{
			name: "Case with a mix of releases (draft, prerelease, release) and a ReleaseType with draft, prerelease and releases",
			spec: Spec{
				Owner:      "updatecli",
				Repository: "updatecli",
				Token:      "superSecretTokenForJoe",
				Username:   "joe",
			},
			releaseType: ReleaseType{
				Draft:      true,
				PreRelease: true,
				Release:    true,
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
									Name:    "v0.18.4 🌈",
									TagName: "v0.18.4",
									IsDraft: true,
								},
							},
							{
								Cursor: "Y3Vyc29yOnYyOpK5MjAyMi0wMS0xOFQxOTo0Nzo1MyswMTowMM4Da8vn",
								Node: releaseNode{
									Name:     "v0.18.3",
									TagName:  "v0.18.3",
									IsLatest: true,
								},
							},
							{
								Cursor: "Y3Vyc29yOnYyOpK5MjAyMS0xMS0xM1QxMjozNzoxMiswMTowMM4DK66W",
								Node: releaseNode{
									Name:         "v1.0.0",
									TagName:      "v1.0.0",
									IsPrerelease: true,
								},
							},
							{
								Cursor: "Y3Vyc29yOnYyOpK5MjAyMC0wMi0xOVQyMTozMDoxMyswMTowMM4BYOFb",
								Node: releaseNode{
									Name:    "v0.0.1",
									TagName: "v0.0.1",
								},
							},
						},
					},
				},
			},
			wantReleases: []string{"v0.0.1", "v1.0.0", "v0.18.3", "v0.18.4"},
		},
		{
			name: "Case with a mix of releases (draft, prerelease, release) and a ReleaseType with 'Latest'",
			spec: Spec{
				Owner:      "updatecli",
				Repository: "updatecli",
				Token:      "superSecretTokenForJoe",
				Username:   "joe",
			},
			releaseType: ReleaseType{
				Latest: true,
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
									Name:    "v0.18.4 🌈",
									TagName: "v0.18.4",
									IsDraft: true,
								},
							},
							{
								Cursor: "Y3Vyc29yOnYyOpK5MjAyMi0wMS0xOFQxOTo0Nzo1MyswMTowMM4Da8vn",
								Node: releaseNode{
									Name:     "v0.18.3",
									TagName:  "v0.18.3",
									IsLatest: true,
								},
							},
							{
								Cursor: "Y3Vyc29yOnYyOpK5MjAyMS0xMS0xM1QxMjozNzoxMiswMTowMM4DK66W",
								Node: releaseNode{
									Name:         "v1.0.0",
									TagName:      "v1.0.0",
									IsPrerelease: true,
								},
							},
							{
								Cursor: "Y3Vyc29yOnYyOpK5MjAyMC0wMi0xOVQyMTozMDoxMyswMTowMM4BYOFb",
								Node: releaseNode{
									Name:    "v0.0.1",
									TagName: "v0.0.1",
								},
							},
						},
					},
				},
			},
			wantReleases: []string{"v0.18.3"},
		},
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
			releaseType := tt.releaseType
			releaseType.Init()

			got, err := sut.SearchReleases(releaseType)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			assert.Equal(t, tt.wantReleases, got)
		})
	}
}
