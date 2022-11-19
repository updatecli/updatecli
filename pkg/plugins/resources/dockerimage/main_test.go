package dockerimage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name                  string
		spec                  Spec
		wantSpec              Spec
		wantVersionFilter     version.Filter
		wantRemoteOptionsSize int
		wantErr               bool
	}{
		{
			name: "Normal case with a single architecture",
			spec: Spec{
				Architecture: "arm64",
				Image:        "ghcr.io/updatecli/updatecli",
				Tag:          "0.15.0",
				InlineKeyChain: docker.InlineKeyChain{
					Username: "eva",
					Password: "SuperSecretPass",
				},
			},
			wantSpec: Spec{
				Architectures: []string{"arm64"},
				Image:         "ghcr.io/updatecli/updatecli",
				Tag:           "0.15.0",
				InlineKeyChain: docker.InlineKeyChain{
					Username: "eva",
					Password: "SuperSecretPass",
				},
			},
			wantVersionFilter: version.Filter{
				Kind:    "latest",
				Pattern: "latest",
			},
			// Only 1 architecture: there is an option for the platform
			wantRemoteOptionsSize: 2,
		},
		{
			name: "Normal case with default (implicit) architecture",
			spec: Spec{
				Image: "ghcr.io/updatecli/updatecli",
				Tag:   "0.15.0",
				InlineKeyChain: docker.InlineKeyChain{
					Username: "eva",
					Password: "SuperSecretPass",
				},
			},
			wantSpec: Spec{
				Architectures: []string{"amd64"},
				Image:         "ghcr.io/updatecli/updatecli",
				Tag:           "0.15.0",
				InlineKeyChain: docker.InlineKeyChain{
					Username: "eva",
					Password: "SuperSecretPass",
				},
			},
			wantVersionFilter: version.Filter{
				Kind:    "latest",
				Pattern: "latest",
			},
			// Only 1 architecture: there is an option for the platform
			wantRemoteOptionsSize: 2,
		},
		{
			name: "Normal case with multiple architectures",
			spec: Spec{
				Architectures: []string{"amd64", "arm/v7", "arm64"},
				Image:         "ghcr.io/updatecli/updatecli",
				Tag:           "0.15.0",
				InlineKeyChain: docker.InlineKeyChain{
					Username: "eva",
					Password: "SuperSecretPass",
				},
			},
			wantSpec: Spec{
				Architectures: []string{"amd64", "arm/v7", "arm64"},
				Image:         "ghcr.io/updatecli/updatecli",
				Tag:           "0.15.0",
				InlineKeyChain: docker.InlineKeyChain{
					Username: "eva",
					Password: "SuperSecretPass",
				},
			},
			wantVersionFilter: version.Filter{
				Kind:    "latest",
				Pattern: "latest",
			},
			// Multiple architectures: there is NO option for the platform
			wantRemoteOptionsSize: 1,
		},
		{
			name: "Normal case with multiple architectures but only one in the list",
			spec: Spec{
				Architectures: []string{"amd64"},
				Image:         "ghcr.io/updatecli/updatecli",
				Tag:           "0.15.0",
				InlineKeyChain: docker.InlineKeyChain{
					Username: "eva",
					Password: "SuperSecretPass",
				},
			},
			wantSpec: Spec{
				Architectures: []string{"amd64"},
				Image:         "ghcr.io/updatecli/updatecli",
				Tag:           "0.15.0",
				InlineKeyChain: docker.InlineKeyChain{
					Username: "eva",
					Password: "SuperSecretPass",
				},
			},
			wantVersionFilter: version.Filter{
				Kind:    "latest",
				Pattern: "latest",
			},
			// Only 1 architecture: there is an option for the platform
			wantRemoteOptionsSize: 2,
		},
		{
			name: "Invalid Spec provided (no password but username)",
			spec: Spec{
				Architecture: "arm64",
				Image:        "ghcr.io/updatecli/updatecli",
				Tag:          "0.15.0",
				InlineKeyChain: docker.InlineKeyChain{
					Username: "eva",
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid Spec provided (both single and  multiple architectures)",
			spec: Spec{
				Architecture:  "arm64",
				Architectures: []string{"arm64", "s390x"},
				Image:         "ghcr.io/updatecli/updatecli",
				Tag:           "0.15.0",
				InlineKeyChain: docker.InlineKeyChain{
					Username: "eva",
					Password: "SuperSecretPass",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := New(tt.spec)
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.wantSpec, got.spec)
			assert.Equal(t, tt.wantVersionFilter, got.versionFilter)
			assert.Equal(t, tt.wantRemoteOptionsSize, len(got.options))
		})
	}
}

func Test_Validate(t *testing.T) {
	tests := []struct {
		name    string
		spec    Spec
		wantErr bool
	}{
		{
			name: "Normal case",
			spec: Spec{
				Architecture: "amd64",
				Image:        "ghcr.io/updatecli/updatecli",
				Tag:          "0.15.0",
				InlineKeyChain: docker.InlineKeyChain{
					Username: "eva",
					Password: "SuperSecretPass",
				},
			},
			wantErr: false,
		},
		{
			name: "Error when token with username",
			spec: Spec{
				Architecture: "amd64",
				Image:        "ghcr.io/updatecli/updatecli",
				Tag:          "0.15.0",
				InlineKeyChain: docker.InlineKeyChain{
					Username: "eva",
					Password: "SuperSecretPass",
					Token:    "UberBearerToken",
				},
			},
			wantErr: true,
		},
		{
			name: "Error when only password",
			spec: Spec{
				Architecture: "amd64",
				Image:        "ghcr.io/updatecli/updatecli",
				Tag:          "0.15.0",
				InlineKeyChain: docker.InlineKeyChain{
					Password: "SuperSecretPass",
				},
			},
			wantErr: true,
		},
		{
			name: "Error when only username",
			spec: Spec{
				Architecture: "amd64",
				Image:        "ghcr.io/updatecli/updatecli",
				Tag:          "0.15.0",
				InlineKeyChain: docker.InlineKeyChain{
					Username: "eva",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut := DockerImage{
				spec: tt.spec,
			}
			gotErr := sut.spec.InlineKeyChain.Validate()
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
		})
	}
}
