package dockerimage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewImage(t *testing.T) {
	tests := []struct {
		name         string
		imageName    string
		architecture string
		want         Image
		wantErr      bool
	}{
		{
			name:      "User image on Docker Hub",
			imageName: "user/jenkinsci",
			want: Image{
				Namespace:    "user",
				Registry:     "registry-1.docker.io",
				Repository:   "jenkinsci",
				Tag:          "latest",
				Architecture: "amd64",
			},
		},
		{
			name:      "User image on Docker Hub with tag",
			imageName: "user/jenkinsci:alpine",
			want: Image{
				Namespace:    "user",
				Registry:     "registry-1.docker.io",
				Repository:   "jenkinsci",
				Tag:          "alpine",
				Architecture: "amd64",
			},
		},
		{
			name:      "Official image on Docker Hub with tag",
			imageName: "ubuntu:18.04",
			want: Image{
				Namespace:    "library",
				Registry:     "registry-1.docker.io",
				Repository:   "ubuntu",
				Tag:          "18.04",
				Architecture: "amd64",
			},
		},
		{
			name:      "GHCR User image with explicit tag ",
			imageName: "ghcr.io/olblak/updatecli:v0.16.0",
			want: Image{
				Namespace:    "olblak",
				Registry:     "ghcr.io",
				Repository:   "updatecli",
				Tag:          "v0.16.0",
				Architecture: "amd64",
			},
		},
		{
			name:         "Quay.io user image without tag and custom architecture",
			imageName:    "quay.io/ansible/ansible-runner",
			architecture: "s390x",
			want: Image{
				Namespace:    "ansible",
				Registry:     "quay.io",
				Repository:   "ansible-runner",
				Tag:          "latest",
				Architecture: "s390x",
			},
		},
		{
			name:      "Quay.io user image without architecture",
			imageName: "quay.io/ansible/ansible-runner",
			want: Image{
				Namespace:    "ansible",
				Registry:     "quay.io",
				Repository:   "ansible-runner",
				Tag:          "latest",
				Architecture: "amd64",
			},
		},
		{
			name:      "Invalid image name provided",
			imageName: "",
			want:      Image{},
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := New(tt.imageName, tt.architecture)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_tag(t *testing.T) {
	tests := []struct {
		name      string
		imageName string
		want      string
	}{
		{
			name:      "User image with explicit tag from Docker Hub (implicit registry)",
			imageName: "jenkinsci/jenkinsci:alpine",
			want:      "alpine",
		},
		{
			name:      "User image with default tag from Docker Hub (implicit registry)",
			imageName: "jenkinsci/jenkinsci",
			want:      "latest",
		},
		{
			name:      "Official image with explicit tag from Docker Hub (implicit registry)",
			imageName: "nginx:alpine",
			want:      "alpine",
		},
		{
			name:      "Official image with default tag from Docker Hub (implicit registry)",
			imageName: "nginx",
			want:      "latest",
		},
		{
			name:      "Official image with explicit tag from Docker Hub (implicit registry, explicit namespace)",
			imageName: "library/nginx:alpine",
			want:      "alpine",
		},
		{
			name:      "Official image with default tag from Docker Hub (implicit registry, explicit namespace)",
			imageName: "library/nginx",
			want:      "latest",
		},
		{
			name:      "User image with explicit tag (explicit registry)",
			imageName: "quay.io/ansible/ansible-runner:4",
			want:      "4",
		},
		{
			name:      "User image with default tag (explicit registry)",
			imageName: "quay.io/ansible/ansible-runner",
			want:      "latest",
		},
		{
			name:      "Namespaced image in a local registry with explicit tag without port",
			imageName: "192.168.1.0/admins/ubuntu:18.04",
			want:      "18.04",
		},
		{
			name:      "Namespaced image in a local registry with default tag without port",
			imageName: "192.168.1.0/admins/ubuntu",
			want:      "latest",
		},
		{
			name:      "Simple image in a local registry with explicit tag without port",
			imageName: "192.168.1.0/nodejs:3.0",
			want:      "3.0",
		},
		{
			name:      "Simple image in a local registry with default tag without port",
			imageName: "192.168.1.0/nodejs",
			want:      "latest",
		},
		{
			name:      "Namespaced image in a local registry and port",
			imageName: "192.168.1.0:5000/admins/ubuntu:18.04",
			want:      "18.04",
		},
		{
			name:      "Simple image with explicit tag in a local registry and port",
			imageName: "192.168.1.0:5000/ubuntu:18.04",
			want:      "18.04",
		},
		{
			name:      "Simple image with default tag in a local registry and port",
			imageName: "192.168.1.0:5000/ubuntu",
			want:      "latest",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got := tag(tt.imageName)

			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_namespace(t *testing.T) {
	tests := []struct {
		name      string
		imageName string
		want      string
	}{
		{
			name:      "User image from Docker Hub (implicit registry)",
			imageName: "user/jenkinsci:latest",
			want:      "user",
		},
		{
			name:      "Official image (default implicit namespace) from Docker Hub (implicit registry)",
			imageName: "nginx:alpine",
			want:      "library",
		},
		{
			name:      "Official image (default explicit namespace) from Docker Hub (implicit registry)",
			imageName: "library/nginx:alpine",
			want:      "library",
		},
		{
			name:      "User image from Docker Hub (explicit registry)",
			imageName: "hub.docker.com/user/jenkinsci:latest",
			want:      "user",
		},
		{
			name:      "Official image (default implicit namespace) from Docker Hub (explicit registry)",
			imageName: "hub.docker.com/nginx:alpine",
			want:      "library",
		},
		{
			name:      "Official image (default explicit namespace) from Docker Hub (explicit registry)",
			imageName: "hub.docker.com/library/nginx:alpine",
			want:      "library",
		},
		{
			name:      "User image from Quay.io (explicit registry)",
			imageName: "quay.io/ansible/ansible-runner",
			want:      "ansible",
		},
		{
			name:      "Namespaced image in a local registry without port",
			imageName: "192.168.1.0/admins/ubuntu:18.04",
			want:      "admins",
		},
		{
			name:      "Simple image in a local registry without port",
			imageName: "192.168.1.0/nodejs:latest",
			want:      "library",
		},
		{
			name:      "Namespaced image in a local registry and port",
			imageName: "192.168.1.0:5000/admins/ubuntu:18.04",
			want:      "admins",
		},
		{
			name:      "Simple image in a local registry and port",
			imageName: "192.168.1.0:5000/ubuntu:18.04",
			want:      "library",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got := namespace(tt.imageName)

			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_repository(t *testing.T) {
	tests := []struct {
		name      string
		imageName string
		want      string
	}{
		{
			name:      "User image from Docker Hub (implicit registry)",
			imageName: "user/jenkinsci:latest",
			want:      "jenkinsci",
		},
		{
			name:      "Official image (default implicit namespace) from Docker Hub (implicit registry)",
			imageName: "nginx:alpine",
			want:      "nginx",
		},
		{
			name:      "Official image (default explicit namespace) from Docker Hub (implicit registry)",
			imageName: "library/nginx:alpine",
			want:      "nginx",
		},
		{
			name:      "User image from Docker Hub (explicit registry)",
			imageName: "hub.docker.com/user/jenkinsci:latest",
			want:      "jenkinsci",
		},
		{
			name:      "Official image (default implicit namespace) from Docker Hub (explicit registry)",
			imageName: "hub.docker.com/nginx:alpine",
			want:      "nginx",
		},
		{
			name:      "Official image (default explicit namespace) from Docker Hub (explicit registry)",
			imageName: "hub.docker.com/library/nginx:alpine",
			want:      "nginx",
		},
		{
			name:      "User image from Quay.io (explicit registry)",
			imageName: "quay.io/ansible/ansible-runner",
			want:      "ansible-runner",
		},
		{
			name:      "Namespaced image in a local registry without port",
			imageName: "192.168.1.0/admins/ubuntu:18.04",
			want:      "ubuntu",
		},
		{
			name:      "Simple image in a local registry without port",
			imageName: "192.168.1.0/nodejs:latest",
			want:      "nodejs",
		},
		{
			name:      "Namespaced image in a local registry and port",
			imageName: "192.168.1.0:5000/admins/ubuntu:18.04",
			want:      "ubuntu",
		},
		{
			name:      "Simple image in a local registry and port",
			imageName: "192.168.1.0:5000/ubuntu:18.04",
			want:      "ubuntu",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got := repository(tt.imageName)

			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_registry(t *testing.T) {
	tests := []struct {
		name         string
		imageName    string
		wantHostname string
	}{
		{
			name:         "User image from Docker Hub (implicit registry)",
			imageName:    "jenkinsci/jenkinsci:latest",
			wantHostname: "registry-1.docker.io",
		},
		{
			name:         "Official image (short name) from Docker Hub (implicit registry)",
			imageName:    "nginx:alpine",
			wantHostname: "registry-1.docker.io",
		},
		{
			name:         "Official image (full name) from Docker Hub (implicit registry)",
			imageName:    "library/nginx:alpine",
			wantHostname: "registry-1.docker.io",
		},
		{
			name:         "User image from Docker Hub (explicit registry)",
			imageName:    "hub.docker.com/jenkinsci/jenkinsci:latest",
			wantHostname: "hub.docker.com",
		},
		{
			name:         "Official image (short name) from Docker Hub (explicit registry)",
			imageName:    "hub.docker.com/nginx:alpine",
			wantHostname: "hub.docker.com",
		},
		{
			name:         "Official image (full name) from Docker Hub (explicit registry)",
			imageName:    "hub.docker.com/library/nginx:alpine",
			wantHostname: "hub.docker.com",
		},
		{
			name:         "User image from Quay.io (explicit registry)",
			imageName:    "quay.io/ansible/ansible-runner",
			wantHostname: "quay.io",
		},
		{
			name:         "Namespaced image in a local registry without port",
			imageName:    "192.168.1.0/admins/ubuntu:18.04",
			wantHostname: "192.168.1.0",
		},
		{
			name:         "Simple image in a local registry without port",
			imageName:    "192.168.1.0/nodejs:latest",
			wantHostname: "192.168.1.0",
		},
		{
			name:         "Namespaced image in a local registry and port",
			imageName:    "192.168.1.0:5000/admins/ubuntu:18.04",
			wantHostname: "192.168.1.0:5000",
		},
		{
			name:         "Simple image in a local registry and port",
			imageName:    "192.168.1.0:5000/ubuntu:18.04",
			wantHostname: "192.168.1.0:5000",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got := registry(tt.imageName)

			assert.Equal(t, tt.wantHostname, got)
		})
	}
}
