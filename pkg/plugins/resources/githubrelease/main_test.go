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
		spec    Spec
		want    GitHubRelease
		wantErr bool
	}{
		{
			name: "Nominal case",
			spec: Spec{
				Repository: "updatecli",
				Owner:      "updatecli",
				Username:   "joe",
				Token:      "superSecretTOkenOfJoe",
				URL:        "github.com",
			},
			want: GitHubRelease{
				ghHandler: &github.Github{
					Spec: github.Spec{
						Repository: "updatecli",
						Owner:      "updatecli",
						Username:   "joe",
						Token:      "superSecretTOkenOfJoe",
						URL:        "https://github.com",
					},
				},
				versionFilter: version.Filter{
					Kind:    "latest",
					Pattern: "latest",
				},
			},
		},
		{
			name: "Nominal case with empty directory",
			spec: Spec{
				Repository: "updatecli",
				Owner:      "updatecli",
				Username:   "joe",
				Token:      "superSecretTOkenOfJoe",
				URL:        "github.com",
			},
			want: GitHubRelease{
				ghHandler: &github.Github{
					Spec: github.Spec{
						Repository: "updatecli",
						Owner:      "updatecli",
						Directory:  path.Join(tmp.Directory, "updatecli", "updatecli"),
						Username:   "joe",
						Token:      "superSecretTOkenOfJoe",
						URL:        "https://github.com",
					},
				},
				versionFilter: version.Filter{
					Kind:    "latest",
					Pattern: "latest",
				},
			},
		},
		{
			name: "Nominal case with empty URL",
			spec: Spec{
				Repository: "updatecli",
				Owner:      "updatecli",
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
						URL:        "https://github.com",
					},
				},
				versionFilter: version.Filter{
					Kind:    "latest",
					Pattern: "latest",
				},
			},
		},
		{
			name: "Validation Error (missing token)",
			spec: Spec{
				Repository: "updatecli",
				Owner:      "updatecli",
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
			// Check that the specified attributes were passed to the GitHub SCM handler
			assert.Equal(t, tt.want.ghHandler.(*github.Github).Spec.Owner, got.ghHandler.(*github.Github).Spec.Owner)
			assert.Equal(t, tt.want.ghHandler.(*github.Github).Spec.Repository, got.ghHandler.(*github.Github).Spec.Repository)
			assert.Equal(t, tt.want.ghHandler.(*github.Github).Spec.Token, got.ghHandler.(*github.Github).Spec.Token)
			assert.Equal(t, tt.want.ghHandler.(*github.Github).Spec.URL, got.ghHandler.(*github.Github).Spec.URL)
			assert.Equal(t, tt.want.ghHandler.(*github.Github).Spec.Username, got.ghHandler.(*github.Github).Spec.Username)
			assert.Equal(t, tt.want.versionFilter, got.versionFilter)
		})
	}
}
