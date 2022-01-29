package maven

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/mavenmetadata"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		spec    Spec
		want    *Maven
		wantErr bool
	}{
		{
			name: "Normal Maven case",
			spec: Spec{
				URL:        "repo.jenkins-ci.org",
				Repository: "releases",
				GroupID:    "org.eclipse.mylyn.wikitext",
				ArtifactID: "wikitext.core",
			},
			want: &Maven{
				spec: Spec{
					URL:        "repo.jenkins-ci.org",
					Repository: "releases",
					GroupID:    "org.eclipse.mylyn.wikitext",
					ArtifactID: "wikitext.core",
				},
				metadataHandler: &mavenmetadata.DefaultHandler{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.spec)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want.spec, got.spec)
			assert.IsType(t, tt.want.metadataHandler, tt.want.metadataHandler)
		})
	}
}
