package githubrelease

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/tmp"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		spec    github.Spec
		want    GitHubRelease
		wantErr bool
	}{
		{
			name: "Nominal case",
			spec: github.Spec{
				Branch:     "main",
				Repository: "updatecli",
				Owner:      "updatecli",
				Directory:  "/home/updatecli",
				Username:   "joe",
				Token:      "superSecretTOkenOfJoe",
				URL:        "github.com",
			},
			want: GitHubRelease{
				ghHandler: &github.Github{
					Spec: github.Spec{
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
		},
		{
			name: "Nominal case with empty directory",
			spec: github.Spec{
				Branch:     "main",
				Repository: "updatecli",
				Owner:      "updatecli",
				Username:   "joe",
				Token:      "superSecretTOkenOfJoe",
				URL:        "github.com",
			},
			want: GitHubRelease{
				ghHandler: &github.Github{
					Spec: github.Spec{
						Branch:     "main",
						Repository: "updatecli",
						Owner:      "updatecli",
						Directory:  path.Join(tmp.Directory, "updatecli", "updatecli"),
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
		},
		{
			name: "Nominal case with empty URL",
			spec: github.Spec{
				Branch:     "main",
				Repository: "updatecli",
				Owner:      "updatecli",
				Directory:  "/home/updatecli",
				Username:   "joe",
				Token:      "superSecretTOkenOfJoe",
			},
			want: GitHubRelease{
				ghHandler: &github.Github{
					Spec: github.Spec{
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
		},
		{
			name: "Validation Error (missing token)",
			spec: github.Spec{
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
			assert.Equal(t,
				tt.want.ghHandler.(*github.Github).Spec,
				got.ghHandler.(*github.Github).Spec,
			)
		})
	}
}
