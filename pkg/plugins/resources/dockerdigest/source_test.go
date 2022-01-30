package dockerdigest

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker/dockerimage"
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker/dockermocks"
)

func TestDockerDigest_Source(t *testing.T) {
	tests := []struct {
		name                     string
		workingDir               string
		mockImg                  dockerimage.Image
		mockRegistryReturnDigest string
		mockRegistryReturnError  error
		wantSource               string
		wantErr                  bool
		wantMockImageName        string
	}{
		{
			name: "Normal case",
			mockImg: dockerimage.Image{
				Registry:     "hub.docker.com",
				Namespace:    "library",
				Repository:   "nginx",
				Tag:          "latest",
				Architecture: "amd64",
			},
			mockRegistryReturnDigest: "123456789",
			wantSource:               "123456789",
			wantMockImageName:        "hub.docker.com/library/nginx:latest",
		},
		{
			name: "Empty digest returned (e.g. image manifest not found)",
			mockImg: dockerimage.Image{
				Registry:     "hub.docker.com",
				Namespace:    "library",
				Repository:   "nginx",
				Tag:          "latest",
				Architecture: "amd64",
			},
			mockRegistryReturnDigest: "",
			wantErr:                  true,
			wantMockImageName:        "hub.docker.com/library/nginx:latest",
		},
		{
			name: "Registry's Digest() method returns an error",
			mockImg: dockerimage.Image{
				Registry:     "hub.docker.com",
				Namespace:    "library",
				Repository:   "nginx",
				Tag:          "latest",
				Architecture: "amd64",
			},
			mockRegistryReturnDigest: "123456789",
			mockRegistryReturnError:  fmt.Errorf("HTTP/403 Unauthorized"),
			wantErr:                  true,
			wantMockImageName:        "hub.docker.com/library/nginx:latest",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRegistry := dockermocks.MockRegistry{
				ReturnedDigest: tt.mockRegistryReturnDigest,
				ReturnedError:  tt.mockRegistryReturnError,
			}
			ds := &DockerDigest{
				image:    tt.mockImg,
				registry: &mockRegistry,
			}
			gotSource, gotErr := ds.Source(tt.workingDir)
			if tt.wantErr {
				assert.Error(t, gotErr)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.wantSource, gotSource)
			assert.Equal(t, tt.wantMockImageName, mockRegistry.InputImageName)
		})
	}
}
