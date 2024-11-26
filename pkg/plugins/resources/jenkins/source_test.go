package jenkins

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/mavenmetadata"
)

func TestJenkins_Source(t *testing.T) {
	tests := []struct {
		name                  string
		workingDir            string
		spec                  Spec
		mockedMetadataHandler mavenmetadata.Handler
		want                  []result.SourceInformation
		wantErr               bool
	}{
		{
			name: "Normal case with latest stable version",
			spec: Spec{
				Release: STABLE,
			},
			mockedMetadataHandler: &mavenmetadata.MockMetadataHandler{
				LatestVersion: "2.331",
				Versions:      []string{"2.319", "2.319.1", "2.319.2", "2.320", "2.331"},
			},
			want: []result.SourceInformation{{
				Value: "2.319.2",
			}},
		},
		{
			name: "Normal case with latest weekly version",
			spec: Spec{
				Release: WEEKLY,
			},
			mockedMetadataHandler: &mavenmetadata.MockMetadataHandler{
				LatestVersion: "2.331",
				Versions:      []string{"2.319", "2.319.1", "2.319.2", "2.320", "2.331"},
			},
			want: []result.SourceInformation{{
				Value: "2.331",
			}},
		},
		{
			name: "Error case with an error returned by maven meta",
			spec: Spec{
				Release: WEEKLY,
			},
			mockedMetadataHandler: &mavenmetadata.MockMetadataHandler{
				LatestVersion: "2.331",
				Versions:      []string{"2.319", "2.319.1", "2.319.2", "2.320", "2.331"},
				Err:           fmt.Errorf("Network general error"),
			},
			wantErr: true,
		},
		{
			name: "Normal case with latest stable version",
			spec: Spec{
				Release: "FOO",
			},
			mockedMetadataHandler: &mavenmetadata.MockMetadataHandler{
				LatestVersion: "2.331",
				Versions:      []string{"2.319", "2.319.1", "2.319.2", "2.320", "2.331"},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut := Jenkins{
				spec:             tt.spec,
				mavenMetaHandler: tt.mockedMetadataHandler,
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
