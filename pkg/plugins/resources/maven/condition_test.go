package maven

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/utils/mavenmetadata"
)

func TestCondition(t *testing.T) {
	tests := []struct {
		name                  string
		source                string
		spec                  Spec
		mockedMetadataHandler mavenmetadata.Handler
		want                  bool
		wantErr               bool
	}{
		{
			name:   "Passing case with input source for org.eclipse.mylyn.wikitext.wikitext.core on repo.jenkins-ci.org/releases",
			source: "1.7.4.v20130429",
			spec: Spec{
				URL:        "repo.jenkins-ci.org",
				Repository: "releases",
				GroupID:    "org.eclipse.mylyn.wikitext",
				ArtifactID: "wikitext.core",
			},
			mockedMetadataHandler: &mavenmetadata.MockMetadataHandler{
				Versions: []string{"1.7.4.v20130429", "1.7.3", "1.7.2"},
			},
			want: true,
		},
		{
			name: "Passing case with specified version for org.eclipse.mylyn.wikitext.wikitext.core on repo.jenkins-ci.org/releases",
			spec: Spec{
				URL:        "repo.jenkins-ci.org",
				Repository: "releases",
				GroupID:    "org.eclipse.mylyn.wikitext",
				ArtifactID: "wikitext.core",
				Version:    "1.7.3",
			},
			mockedMetadataHandler: &mavenmetadata.MockMetadataHandler{
				Versions: []string{"1.7.4.v20130429", "1.7.3", "1.7.2"},
			},
			want: true,
		},
		{
			name:   "Failing case with input source for org.eclipse.mylyn.wikitext.wikitext.core on repo.jenkins-ci.org/releases",
			source: "1.6.0",
			spec: Spec{
				URL:        "repo.jenkins-ci.org",
				Repository: "releases",
				GroupID:    "org.eclipse.mylyn.wikitext",
				ArtifactID: "wikitext.core",
			},
			mockedMetadataHandler: &mavenmetadata.MockMetadataHandler{
				Versions: []string{"1.7.4.v20130429", "1.7.3", "1.7.2"},
			},
			want: false,
		},
		{
			name:   "Error case with an error returned by the handler",
			source: "1.7.4.v20130429",
			spec: Spec{
				URL:        "repo.jenkins-ci.org",
				Repository: "releases",
				GroupID:    "org.eclipse.mylyn.wikitext",
				ArtifactID: "wikitext.core",
			},
			mockedMetadataHandler: &mavenmetadata.MockMetadataHandler{
				Versions: []string{"1.7.4.v20130429", "1.7.3", "1.7.2"},
				Err:      fmt.Errorf("TCP I/O connection timeout"),
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

			got, gotErr := sut.Condition(tt.source)
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.want, got)
		})
	}
}
