package github

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTags(t *testing.T) {
	tests := []struct {
		name        string
		spec        Spec
		mockedQuery *tagsQuery
		mockedError error
		wantTags    []string
		wantErr     bool
	}{
		{
			name: "Nominal case with 4 kubernetes tags",
			spec: Spec{
				Owner:      "kubernetes",
				Repository: "kubernetes",
				Token:      "superSecretTokenForJoe",
				Username:   "joe",
			},
			mockedQuery: &tagsQuery{
				RateLimit: RateLimit{
					Cost:      1,
					Remaining: 5000,
				},
				Repository: struct {
					Refs repositoryRef `graphql:"refs(refPrefix: $refPrefix, last: 100, before: $before,orderBy: $orderBy)"`
				}{
					Refs: repositoryRef{
						TotalCount: 4,
						PageInfo: PageInfo{
							HasNextPage: false,
							EndCursor:   "MTAw",
						},
						Edges: []refEdge{
							{
								Cursor: "MQ",
								Node: refNode{
									Name: "v0.23.2-rc.0",
								},
							},
							{
								Cursor: "MjI",
								Node: refNode{
									Name: "kubernetes-1.23.0",
								},
							},
							{
								Cursor: "Mg",
								Node: refNode{
									Name: "v0.23.1",
								},
							},
							{
								Cursor: "MTAw",
								Node: refNode{
									Name: "kubernetes-1.22.0",
								},
							},
						},
					},
				},
			},
			wantTags: []string{"kubernetes-1.22.0", "v0.23.1", "kubernetes-1.23.0", "v0.23.2-rc.0"},
		},
		{
			name:        "Case with error returned from github query",
			mockedQuery: &tagsQuery{},
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
			got, err := sut.SearchTags(0)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			assert.Equal(t, tt.wantTags, got)
		})
	}
}
