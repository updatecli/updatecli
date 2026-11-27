package docker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegistryMappingsResolve(t *testing.T) {
	testdata := []struct {
		name     string
		mappings RegistryMappings
		image    string
		expected string
	}{
		{
			name:     "no mapping",
			image:    "my-registry.com/docker-io/nginx",
			expected: "my-registry.com/docker-io/nginx",
		},
		{
			name:     "matching prefix with trailing slashes",
			mappings: RegistryMappings{"my-registry.com/docker-io/": "docker.io/"},
			image:    "my-registry.com/docker-io/nginx",
			expected: "docker.io/nginx",
		},
		{
			name:     "matching prefix without trailing slashes",
			mappings: RegistryMappings{"my-registry.com/docker-io": "docker.io"},
			image:    "my-registry.com/docker-io/library/nginx",
			expected: "docker.io/library/nginx",
		},
		{
			name:     "prefix only matches whole path segments",
			mappings: RegistryMappings{"my-registry.com/docker-io": "docker.io"},
			image:    "my-registry.com/docker-iox/nginx",
			expected: "my-registry.com/docker-iox/nginx",
		},
		{
			name:     "prefix equal to the image name",
			mappings: RegistryMappings{"my-registry.com/nginx": "docker.io/nginx"},
			image:    "my-registry.com/nginx",
			expected: "docker.io/nginx",
		},
		{
			name: "longest prefix wins",
			mappings: RegistryMappings{
				"my-registry.com":            "registry.example.com",
				"my-registry.com/docker-io":  "docker.io",
				"my-registry.com/ghcr-io/ab": "ghcr.io/ab",
			},
			image:    "my-registry.com/docker-io/nginx",
			expected: "docker.io/nginx",
		},
		{
			name:     "registry host only",
			mappings: RegistryMappings{"my-registry.com/ghcr-io": "ghcr.io"},
			image:    "my-registry.com/ghcr-io/updatecli/updatecli",
			expected: "ghcr.io/updatecli/updatecli",
		},
		{
			name: "flat mirror, full image name overrides the registry default",
			mappings: RegistryMappings{
				"my-registry.com/":          "docker.io/",
				"my-registry.com/updatecli": "ghcr.io/updatecli/updatecli",
			},
			image:    "my-registry.com/updatecli",
			expected: "ghcr.io/updatecli/updatecli",
		},
		{
			name: "flat mirror, other images use the registry default",
			mappings: RegistryMappings{
				"my-registry.com/":          "docker.io/",
				"my-registry.com/updatecli": "ghcr.io/updatecli/updatecli",
			},
			image:    "my-registry.com/nginx",
			expected: "docker.io/nginx",
		},
		{
			name: "flat mirror, full image name does not match a longer name",
			mappings: RegistryMappings{
				"my-registry.com/":          "docker.io/",
				"my-registry.com/updatecli": "ghcr.io/updatecli/updatecli",
			},
			image:    "my-registry.com/updatecli-action",
			expected: "docker.io/updatecli-action",
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.mappings.Resolve(tt.image))
		})
	}
}

func TestRegistryMappingsValidate(t *testing.T) {
	testdata := []struct {
		name          string
		mappings      RegistryMappings
		expectedError string
	}{
		{
			name:     "nil",
			mappings: nil,
		},
		{
			name:     "valid mapping",
			mappings: RegistryMappings{"my-registry.com/docker-io/": "docker.io/"},
		},
		{
			name:     "upstream registry with a port",
			mappings: RegistryMappings{"my-registry.com/internal/": "registry.example.com:5000/team"},
		},
		{
			name:          "empty mirror prefix",
			mappings:      RegistryMappings{"/": "docker.io"},
			expectedError: "the mirror prefix is empty",
		},
		{
			name:          "empty upstream prefix",
			mappings:      RegistryMappings{"my-registry.com/docker-io": ""},
			expectedError: "the upstream prefix is empty",
		},
		{
			name: "mirror prefixes only differing by a trailing slash",
			mappings: RegistryMappings{
				"my-registry.com/docker-io":  "docker.io",
				"my-registry.com/docker-io/": "quay.io/",
			},
			expectedError: `registry mapping "my-registry.com/docker-io/" => "quay.io/": the mirror prefix is the same as "my-registry.com/docker-io"`,
		},
		{
			name:          "upstream prefix with a tag",
			mappings:      RegistryMappings{"my-registry.com/nginx": "docker.io/library/nginx:1.25"},
			expectedError: "the upstream prefix holds a tag",
		},
		{
			name:          "upstream prefix with a digest",
			mappings:      RegistryMappings{"my-registry.com/nginx": "docker.io/library/nginx@sha256:abc"},
			expectedError: "the upstream prefix holds whitespace or a digest",
		},
		{
			name:          "upstream prefix with whitespace",
			mappings:      RegistryMappings{"my-registry.com/nginx": "docker.io/library/nginx "},
			expectedError: "the upstream prefix holds whitespace or a digest",
		},
		{
			name:          "upstream prefix with a leading slash",
			mappings:      RegistryMappings{"my-registry.com/docker-io": "/docker.io"},
			expectedError: "the upstream prefix has an empty path segment",
		},
		{
			name:          "upstream prefix with a double slash",
			mappings:      RegistryMappings{"my-registry.com/docker-io/": "docker.io//"},
			expectedError: "the upstream prefix has an empty path segment",
		},
		{
			name: "errors are sorted by mirror prefix",
			mappings: RegistryMappings{
				"my-registry.com/b": "",
				"my-registry.com/a": "",
			},
			expectedError: "invalid registry mappings:\n" +
				`registry mapping "my-registry.com/a" => "": the upstream prefix is empty` + "\n" +
				`registry mapping "my-registry.com/b" => "": the upstream prefix is empty`,
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.mappings.Validate()
			if tt.expectedError == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, tt.expectedError)
		})
	}
}
