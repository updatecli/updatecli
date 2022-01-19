package github

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/tmp"
	"github.com/updatecli/updatecli/pkg/plugins/version"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		spec    Spec
		want    Github
		wantErr bool
	}{
		{
			name: "Nominal case",
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
				Spec: Spec{
					Branch:     "main",
					Repository: "updatecli",
					Owner:      "updatecli",
					Directory:  "/home/updatecli",
					Username:   "joe",
					Token:      "superSecretTOkenOfJoe",
					URL:        "github.com",
					VersionFilter: version.Filter{
						Kind:    "latest",
						Pattern: "latest",
					},
				},
			},
		},
		{
			name: "Nominal case with empty directory",
			spec: Spec{
				Branch:     "main",
				Repository: "updatecli",
				Owner:      "updatecli",
				Username:   "joe",
				Token:      "superSecretTOkenOfJoe",
				URL:        "github.com",
			},
			want: Github{
				Spec: Spec{
					Branch:     "main",
					Repository: "updatecli",
					Owner:      "updatecli",
					Username:   "joe",
					Token:      "superSecretTOkenOfJoe",
					URL:        "github.com",
					Directory:  path.Join(tmp.Directory, "updatecli", "updatecli"),
					VersionFilter: version.Filter{
						Kind:    "latest",
						Pattern: "latest",
					},
				},
			},
		},
		{
			name: "Nominal case with empty URL",
			spec: Spec{
				Branch:     "main",
				Repository: "updatecli",
				Owner:      "updatecli",
				Username:   "joe",
				Token:      "superSecretTOkenOfJoe",
				Directory:  "/home/updatecli",
			},
			want: Github{
				Spec: Spec{
					Branch:     "main",
					Repository: "updatecli",
					Owner:      "updatecli",
					Username:   "joe",
					Token:      "superSecretTOkenOfJoe",
					URL:        "github.com",
					Directory:  "/home/updatecli",
					VersionFilter: version.Filter{
						Kind:    "latest",
						Pattern: "latest",
					},
				},
			},
		},
		{
			name: "Validation Error (missing token)",
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
			got, err := New(tt.spec)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want.Spec, got.Spec)
			assert.NotNil(t, got.client)
		})
	}
}
