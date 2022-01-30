package jenkins

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/utils/mavenmetadata"
)

func TestJenkins_Condition(t *testing.T) {
	tests := []struct {
		name                  string
		source                string
		spec                  Spec
		mockedMetadataHandler mavenmetadata.Handler
		want                  bool
		wantErr               bool
	}{
		{
			name:   "Passing case by checking input source with latest stable baseline",
			source: "2.319.2",
			spec: Spec{
				Release: STABLE,
			},
			mockedMetadataHandler: &mavenmetadata.MockMetadataHandler{
				Versions: []string{"2.319", "2.319.1", "2.319.2", "2.320", "2.331"},
			},
			want: true,
		},
		{
			name:   "Passing case by checking specified version with weekly baseline",
			source: "2.320",
			spec: Spec{
				Release: WEEKLY,
			},
			mockedMetadataHandler: &mavenmetadata.MockMetadataHandler{
				Versions: []string{"2.319", "2.319.1", "2.319.2", "2.320", "2.331"},
			},
			want: true,
		},
		{
			name:   "Error case by checking a version-invalid input source",
			source: "LATEST-2.319.2",
			spec: Spec{
				Release: STABLE,
			},
			mockedMetadataHandler: &mavenmetadata.MockMetadataHandler{
				Versions: []string{"2.319", "2.319.1", "2.319.2", "2.320", "2.331"},
			},
			wantErr: true,
		},
		{
			name:   "Error case by checking input source with a different release baseline",
			source: "2.320", // this is a weekly release, not a stable
			spec: Spec{
				Release: STABLE,
			},
			mockedMetadataHandler: &mavenmetadata.MockMetadataHandler{
				Versions: []string{"2.319", "2.319.1", "2.319.2", "2.320", "2.331"},
			},
			wantErr: true,
		},
		{
			name:   "Error case with an error returned by maven meta",
			source: "2.319.2",
			spec: Spec{
				Release: STABLE,
			},
			mockedMetadataHandler: &mavenmetadata.MockMetadataHandler{
				LatestVersion: "2.331",
				Versions:      []string{"2.319", "2.319.1", "2.319.2", "2.320", "2.331"},
				Err:           fmt.Errorf("Network general error"),
			},
			wantErr: true,
		},
		{
			name:   "Error case by checking a version that does not exist in the weekly baseline",
			source: "2.200",
			spec: Spec{
				Release: WEEKLY,
			},
			mockedMetadataHandler: &mavenmetadata.MockMetadataHandler{
				Versions: []string{"2.319", "2.319.1", "2.319.2", "2.320", "2.331"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut := Jenkins{
				spec:             tt.spec,
				mavenMetaHandler: tt.mockedMetadataHandler,
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
