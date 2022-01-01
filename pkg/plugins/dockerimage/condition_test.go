package dockerimage

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/plugins/docker/dockerimage"
	"github.com/updatecli/updatecli/pkg/plugins/docker/dockermocks"
)

func TestDockerImage_Condition(t *testing.T) {
	tests := []struct {
		name                     string
		inputSourceValue         string
		mockImg                  dockerimage.Image
		mockRegistryReturnDigest string
		mockRegistryReturnError  error
		wantResult               bool
		wantErr                  bool
		wantMockImageName        string
	}{
		{
			name:             "Normal case with tag from input source value",
			inputSourceValue: "v1.0.0",
			mockImg: dockerimage.Image{
				Registry:     "hub.docker.com",
				Namespace:    "library",
				Repository:   "nginx",
				Tag:          "latest",
				Architecture: "amd64",
			},
			mockRegistryReturnDigest: "123456789",
			wantResult:               true,
			wantMockImageName:        "hub.docker.com/library/nginx:v1.0.0",
		},
		{
			name: "Normal case with empty input source value (disablesourceinput set to true)",
			mockImg: dockerimage.Image{
				Registry:     "hub.docker.com",
				Namespace:    "library",
				Repository:   "nginx",
				Tag:          "latest",
				Architecture: "amd64",
			},
			mockRegistryReturnDigest: "123456789",
			wantResult:               true,
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
			wantResult:               false,
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
			wantResult:               false,
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
			d := &DockerImage{
				image:    tt.mockImg,
				registry: &mockRegistry,
			}
			gotSource, gotErr := d.Condition(tt.inputSourceValue)
			if tt.wantErr {
				assert.Error(t, gotErr)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.wantResult, gotSource)
			assert.Equal(t, tt.wantMockImageName, mockRegistry.InputImageName)
		})
	}
}

func TestDockerImage_ConditionFromSCM(t *testing.T) {
	tests := []struct {
		name                     string
		inputSourceValue         string
		mockImg                  dockerimage.Image
		mockRegistryReturnDigest string
		mockRegistryReturnError  error
		wantResult               bool
		wantErr                  bool
		wantMockImageName        string
		scm                      scm.ScmHandler
	}{
		{
			name:             "Normal case with tag from input source value",
			inputSourceValue: "v1.0.0",
			mockImg: dockerimage.Image{
				Registry:     "hub.docker.com",
				Namespace:    "library",
				Repository:   "nginx",
				Tag:          "latest",
				Architecture: "amd64",
			},
			mockRegistryReturnDigest: "123456789",
			wantResult:               true,
			wantMockImageName:        "hub.docker.com/library/nginx:v1.0.0",
			scm: &scm.MockScm{
				WorkingDir: "/tmp",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRegistry := dockermocks.MockRegistry{
				ReturnedDigest: tt.mockRegistryReturnDigest,
				ReturnedError:  tt.mockRegistryReturnError,
			}
			d := &DockerImage{
				image:    tt.mockImg,
				registry: &mockRegistry,
			}
			gotSource, gotErr := d.ConditionFromSCM(tt.inputSourceValue, tt.scm)
			if tt.wantErr {
				assert.Error(t, gotErr)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.wantResult, gotSource)
			assert.Equal(t, tt.wantMockImageName, mockRegistry.InputImageName)
		})
	}
}
