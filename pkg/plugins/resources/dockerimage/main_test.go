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
		name    string
		spec    Spec
		want    *DockerImage
		wantErr bool
	}{
		{
			name: "Normal case",
			spec: Spec{
				Architecture: "arm64",
				Image:        "ghcr.io/updatecli/updatecli",
				Tag:          "0.15.0",
				InlineKeyChain: docker.InlineKeyChain{
					Username: "eva",
					Password: "SuperSecretPass",
				},
			},
			want: &DockerImage{
				spec: Spec{
					Architecture: "arm64",
					Image:        "ghcr.io/updatecli/updatecli",
					Tag:          "0.15.0",
					InlineKeyChain: docker.InlineKeyChain{
						Username: "eva",
						Password: "SuperSecretPass",
					},
				},
				versionFilter: version.Filter{
					Kind:    "latest",
					Pattern: "latest",
				},
			},
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
			want:    nil,
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
			assert.Equal(t, tt.want.spec, got.spec)
			assert.Equal(t, tt.want.versionFilter, got.versionFilter)
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
