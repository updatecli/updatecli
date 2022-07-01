package github

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/tmp"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name       string
		spec       Spec
		pipelineID string
		want       Github
		wantErr    bool
	}{
		{
			name:       "Nominal case",
			pipelineID: "12345",
			spec: Spec{
				Branch:     "main",
				Repository: "updatecli",
				Owner:      "updatecli",
				Directory:  "/home/updatecli",
				Username:   "joe",
				Token:      "superSecretTOkenOfJoe",
				URL:        "github.com",
			},
			want: Github{
				HeadBranch: "updatecli_12345",
				Spec: Spec{
					Branch:     "main",
					Repository: "updatecli",
					Owner:      "updatecli",
					Directory:  "/home/updatecli",
					Username:   "joe",
					Token:      "superSecretTOkenOfJoe",
					URL:        "github.com",
				},
			},
		},
		{
			name:       "Nominal case with empty directory",
			pipelineID: "12345",
			spec: Spec{
				Branch:     "main",
				Repository: "updatecli",
				Owner:      "updatecli",
				Username:   "joe",
				Token:      "superSecretTOkenOfJoe",
				URL:        "github.com",
			},
			want: Github{
				HeadBranch: "updatecli_12345",
				Spec: Spec{
					Branch:     "main",
					Repository: "updatecli",
					Owner:      "updatecli",
					Username:   "joe",
					Token:      "superSecretTOkenOfJoe",
					URL:        "github.com",
					Directory:  path.Join(tmp.Directory, "github", "updatecli", "updatecli"),
				},
			},
		},
		{
			name:       "Nominal case with empty URL",
			pipelineID: "12345",
			spec: Spec{
				Branch:     "main",
				Repository: "updatecli",
				Owner:      "updatecli",
				Username:   "joe",
				Token:      "superSecretTOkenOfJoe",
				Directory:  "/home/updatecli",
			},
			want: Github{
				HeadBranch: "updatecli_12345",
				Spec: Spec{
					Branch:     "main",
					Repository: "updatecli",
					Owner:      "updatecli",
					Username:   "joe",
					Token:      "superSecretTOkenOfJoe",
					URL:        "github.com",
					Directory:  "/home/updatecli",
				},
			},
		},
		{
			name:       "Validation Error (missing token)",
			pipelineID: "12345",
			spec: Spec{
				Branch:     "main",
				Repository: "updatecli",
				Owner:      "updatecli",
				Directory:  "/tmp/updatecli",
				Username:   "joe",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.spec, tt.pipelineID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want.Spec, got.Spec)
			assert.Equal(t, tt.want.HeadBranch, got.HeadBranch)
			assert.NotNil(t, got.client)
		})
	}
}
