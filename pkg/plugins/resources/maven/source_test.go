package maven

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/utils/mavenmetadata"
)

func TestSource(t *testing.T) {
	tests := []struct {
		name                  string
		workingDir            string
		spec                  Spec
		mockedMetadataHandler mavenmetadata.Handler
		want                  string
		wantErr               bool
	}{
		{
			name: "Normal case with org.eclipse.mylyn.wikitext.wikitext.core on repo.jenkins-ci.org/releases",
			spec: Spec{
				URL:        "repo.jenkins-ci.org",
				Repository: "releases",
				GroupID:    "org.eclipse.mylyn.wikitext",
				ArtifactID: "wikitext.core",
			},
			mockedMetadataHandler: &mavenmetadata.MockMetadataHandler{
				LatestVersion: "1.7.4.v20130429",
			},
			want: "1.7.4.v20130429",
		},
		{
			name: "Error case with an error returned by the handler",
			spec: Spec{
				URL:        "repo.jenkins-ci.org",
				Repository: "releases",
				GroupID:    "org.eclipse.mylyn.wikitext",
				ArtifactID: "wikitext.core",
			},
			mockedMetadataHandler: &mavenmetadata.MockMetadataHandler{
				Err: fmt.Errorf("TCP I/O general error"),
			},
			wantErr: true,
		},
		{
			name: "Error case with empty latest version returned by the handler",
			spec: Spec{
				URL:        "repo.jenkins-ci.org",
				Repository: "releases",
				GroupID:    "org.eclipse.mylyn.wikitext",
				ArtifactID: "wikitext.core",
			},
			mockedMetadataHandler: &mavenmetadata.MockMetadataHandler{
				LatestVersion: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut := Maven{
				spec: tt.spec,
				metadataHandlers: []mavenmetadata.Handler{
					tt.mockedMetadataHandler,
				},
			}

			got, gotErr := sut.Source(tt.workingDir)
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.want, got)
		})
	}
}
