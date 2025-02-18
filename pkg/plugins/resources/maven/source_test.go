package maven

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/mavenmetadata"
)

func TestSource(t *testing.T) {
	tests := []struct {
		name                   string
		workingDir             string
		spec                   Spec
		mockedMetadataHandlers []mavenmetadata.Handler
		want                   string
		wantErr                bool
	}{
		{
			name: "Normal case with org.eclipse.mylyn.wikitext.wikitext.core on repo.jenkins-ci.org/releases",
			spec: Spec{
				Repository: "repo.jenkins-ci.org/releases",
				GroupID:    "org.eclipse.mylyn.wikitext",
				ArtifactID: "wikitext.core",
			},
			mockedMetadataHandlers: []mavenmetadata.Handler{
				&mavenmetadata.MockMetadataHandler{
					LatestVersion: "1.7.4.v20130429",
				},
			},
			want: "1.7.4.v20130429",
		},
		{
			name: "Error case with an error returned by the handler",
			spec: Spec{
				Repository: "repo.jenkins-ci.org/releases",
				GroupID:    "org.eclipse.mylyn.wikitext",
				ArtifactID: "wikitext.core",
			},
			mockedMetadataHandlers: []mavenmetadata.Handler{
				&mavenmetadata.MockMetadataHandler{
					Err: fmt.Errorf("TCP I/O general error"),
				},
			},
			wantErr: true,
		},
		{
			name: "Error case with empty latest version returned by the handler",
			spec: Spec{
				Repository: "repo.jenkins-ci.org/releases",
				GroupID:    "org.eclipse.mylyn.wikitext",
				ArtifactID: "wikitext.core",
			},
			mockedMetadataHandlers: []mavenmetadata.Handler{
				&mavenmetadata.MockMetadataHandler{
					LatestVersion: "",
				},
			},
			wantErr: true,
		},
		{
			name: "Returns version from first repo",
			spec: Spec{
				Repositories: []string{
					"some.private.repo/releases",
					"repo.jenkins-ci.org/releases",
				},
				GroupID:    "org.eclipse.mylyn.wikitext",
				ArtifactID: "wikitext.core",
			},
			mockedMetadataHandlers: []mavenmetadata.Handler{
				&mavenmetadata.MockMetadataHandler{
					LatestVersion: "1.2.3",
				},
				&mavenmetadata.MockMetadataHandler{
					Err: fmt.Errorf("Not found"),
				},
			},
			want: "1.2.3",
		},
		{
			name: "Returns version from second repo if the first one returns an error",
			spec: Spec{
				Repositories: []string{
					"some.private.repo/releases",
					"repo.jenkins-ci.org/releases",
				},
				GroupID:    "org.eclipse.mylyn.wikitext",
				ArtifactID: "wikitext.core",
			},
			mockedMetadataHandlers: []mavenmetadata.Handler{
				&mavenmetadata.MockMetadataHandler{
					Err: fmt.Errorf("Not found"),
				},
				&mavenmetadata.MockMetadataHandler{
					LatestVersion: "1.2.3",
				},
			},
			want: "1.2.3",
		},
		{
			name: "Returns version from second repo if the first one returns an empty version",
			spec: Spec{
				Repositories: []string{
					"some.private.repo/releases",
					"repo.jenkins-ci.org/releases",
				},
				GroupID:    "org.eclipse.mylyn.wikitext",
				ArtifactID: "wikitext.core",
			},
			mockedMetadataHandlers: []mavenmetadata.Handler{
				&mavenmetadata.MockMetadataHandler{
					LatestVersion: "",
				},
				&mavenmetadata.MockMetadataHandler{
					LatestVersion: "1.2.3",
				},
			},
			want: "1.2.3",
		},
		{
			name: "Fails if both repos return an error",
			spec: Spec{
				Repositories: []string{
					"some.private.repo/releases",
					"repo.jenkins-ci.org/releases",
				},
				GroupID:    "org.eclipse.mylyn.wikitext",
				ArtifactID: "wikitext.core",
			},
			mockedMetadataHandlers: []mavenmetadata.Handler{
				&mavenmetadata.MockMetadataHandler{
					Err: fmt.Errorf("Not found"),
				},
				&mavenmetadata.MockMetadataHandler{
					Err: fmt.Errorf("Not found"),
				},
			},
			wantErr: true,
		},
		{
			name: "Fails if both repos return an empty version",
			spec: Spec{
				Repositories: []string{
					"some.private.repo/releases",
					"repo.jenkins-ci.org/releases",
				},
				GroupID:    "org.eclipse.mylyn.wikitext",
				ArtifactID: "wikitext.core",
			},
			mockedMetadataHandlers: []mavenmetadata.Handler{
				&mavenmetadata.MockMetadataHandler{
					LatestVersion: "",
				},
				&mavenmetadata.MockMetadataHandler{
					LatestVersion: "",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut := Maven{
				spec:             tt.spec,
				metadataHandlers: tt.mockedMetadataHandlers,
			}

			gotResult := result.Source{}
			gotErr := sut.Source(tt.workingDir, &gotResult)
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.want, gotResult.Information)
		})
	}
}
