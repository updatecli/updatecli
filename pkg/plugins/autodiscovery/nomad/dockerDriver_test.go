package nomad

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNomadTargetPrefix(t *testing.T) {
	testdata := []struct {
		name           string
		spec           nomadDockerSpec
		imageName      string
		expectedPrefix string
		expectedOK     bool
	}{
		{
			name:           "image attribute",
			spec:           nomadDockerSpec{Path: "job.web.group.web.task.nginx.config.image", Value: "nginx:1.25"},
			imageName:      "nginx",
			expectedPrefix: "nginx:",
			expectedOK:     true,
		},
		{
			name:           "variable holding the tag",
			spec:           nomadDockerSpec{Path: "variable.tag.default", Value: "nginx:1.25", VariableValue: "1.25"},
			imageName:      "nginx",
			expectedPrefix: "",
			expectedOK:     true,
		},
		{
			name:           "variable holding the whole image",
			spec:           nomadDockerSpec{Path: "variable.image.default", Value: "nginx:1.25", VariableValue: "nginx:1.25"},
			imageName:      "nginx",
			expectedPrefix: "nginx:",
			expectedOK:     true,
		},
		{
			name:           "variable holding the whole image with a digest",
			spec:           nomadDockerSpec{Path: "variable.image.default", Value: "nginx:1.25@sha256:abc", VariableValue: "nginx:1.25@sha256:abc"},
			imageName:      "nginx",
			expectedPrefix: "nginx:",
			expectedOK:     true,
		},
		{
			name:       "variable holding the image name only",
			spec:       nomadDockerSpec{Path: "variable.image.default", Value: "nginx:1.25", VariableValue: "nginx"},
			imageName:  "nginx",
			expectedOK: false,
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			gotPrefix, gotOK := nomadTargetPrefix(tt.spec, tt.imageName)
			assert.Equal(t, tt.expectedOK, gotOK)
			assert.Equal(t, tt.expectedPrefix, gotPrefix)
		})
	}
}
