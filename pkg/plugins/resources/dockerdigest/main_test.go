package dockerdigest

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker/dockerimage"
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker/dockerregistry"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		spec    Spec
		want    *DockerDigest
		wantErr bool
	}{
		{
			name: "Normal case",
			spec: Spec{
				Architecture: "arm64",
				Image:        "ghcr.io/updatecli/updatecli",
				Tag:          "0.15.0",
				Username:     "eva",
				Password:     "SuperSecretPass",
			},
			want: &DockerDigest{
				spec: Spec{
					Architecture: "arm64",
					Image:        "ghcr.io/updatecli/updatecli",
					Tag:          "0.15.0",
					Username:     "eva",
					Password:     "SuperSecretPass",
				},
				image: dockerimage.Image{
					Registry:     "ghcr.io",
					Namespace:    "updatecli",
					Repository:   "updatecli",
					Tag:          "0.15.0",
					Architecture: "arm64",
				},
				registry: dockerregistry.DockerGenericRegistry{
					Auth: dockerregistry.RegistryAuth{
						Username: "eva",
						Password: "SuperSecretPass",
					},
					WebClient: http.DefaultClient,
				},
			},
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
			assert.Equal(t, tt.want, got)
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
				Username:     "eva",
				Password:     "SuperSecretPass",
			},
			wantErr: false,
		},
		{
			name: "Error when token with username",
			spec: Spec{
				Architecture: "amd64",
				Image:        "ghcr.io/updatecli/updatecli",
				Tag:          "0.15.0",
				Token:        "UberBearerToken",
				Username:     "eva",
				Password:     "SuperSecretPass",
			},
			wantErr: true,
		},
		{
			name: "Error when only password",
			spec: Spec{
				Architecture: "amd64",
				Image:        "ghcr.io/updatecli/updatecli",
				Tag:          "0.15.0",
				Password:     "SuperSecretPass",
			},
			wantErr: true,
		},
		{
			name: "Error when only username",
			spec: Spec{
				Architecture: "amd64",
				Image:        "ghcr.io/updatecli/updatecli",
				Tag:          "0.15.0",
				Username:     "eva",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut := DockerDigest{
				spec: tt.spec,
			}
			gotErr := sut.Validate()
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
		})
	}
}
