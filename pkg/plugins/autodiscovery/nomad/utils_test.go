package nomad

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchFiles(t *testing.T) {
	testdata := []struct {
		name          string
		rootDir       string
		filePatterns  []string
		expectedFiles []string
		expectedErr   error
	}{
		{
			name:         "Nominal case with test data and default file pattern set",
			rootDir:      "testdata",
			filePatterns: DefaultFilePattern,
			expectedFiles: []string{
				"testdata/containerd/redis.nomad",
				"testdata/podman/cache.nomad",
				"testdata/registryMappings/nomad.hcl",
				"testdata/simple/nomad.hcl",
				"testdata/variable/grafana.nomad",
			},
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			gotFiles, err := searchNomadFiles(tt.rootDir, tt.filePatterns)

			if tt.expectedErr != nil {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedFiles, gotFiles)
		})
	}
}

func TestGetNomadSpecFromFile(t *testing.T) {
	testdata := []struct {
		name               string
		filename           string
		expectedNomadSpecs []nomadDockerSpec
		expectedErr        bool
	}{
		{
			name:     "Simple case with multiple docker images",
			filename: "testdata/simple/nomad.hcl",
			expectedNomadSpecs: []nomadDockerSpec{
				{
					File:      "testdata/simple/nomad.hcl",
					Value:     "nginx:latest",
					GroupName: "web-group",
					TaskName:  "frontend",
					JobName:   "multi-docker-example",
					Path:      "job.multi-docker-example.group.web-group.task.frontend.config.image",
				},
				{
					File:      "testdata/simple/nomad.hcl",
					Value:     "hashicorp/http-echo:latest",
					GroupName: "web-group",
					TaskName:  "backend",
					JobName:   "multi-docker-example",
					Path:      "job.multi-docker-example.group.web-group.task.backend.config.image",
				},
			},
		},
		{
			name:     "Case with variable name",
			filename: "testdata/variable/grafana.nomad",
			expectedNomadSpecs: []nomadDockerSpec{
				{
					File:          "testdata/variable/grafana.nomad",
					Value:         "grafana/grafana:latest",
					GroupName:     "grafana",
					TaskName:      "grafana",
					JobName:       "grafana",
					Path:          "variable.image_tag.default",
					VariableValue: "latest",
				},
			},
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			gotDockerComposeServices, err := getNomadDockerSpecFromFile(tt.filename)

			if tt.expectedErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedNomadSpecs, gotDockerComposeServices)
		})
	}
}

func TestGetVariableName(t *testing.T) {
	testdata := []struct {
		name               string
		input              string
		expected           string
		expectedErrMessage error
		expectedErr        bool
	}{
		{
			name:     "Case with variable name",
			input:    "${var.image}",
			expected: "image",
		},
		{
			name:               "Case with multiple variable name",
			input:              "${var.image}${ var.image}",
			expected:           "image",
			expectedErr:        true,
			expectedErrMessage: fmt.Errorf("multiple variable detected in image value %q", "${var.image}${ var.image}"),
		},
		{
			name:     "Case with variable name",
			input:    "latest-${var.image}",
			expected: "image",
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := getVariableName(tt.input)

			if tt.expectedErr {
				require.Error(t, gotErr)
				assert.Equal(t, tt.expectedErrMessage, gotErr)
				return
			}

			require.NoError(t, gotErr)

			assert.Equal(t, tt.expected, got)
		})
	}
}

// TestGetNomadSpecFromFileVariableBeforeTag checks a variable followed by more text in the
// image attribute, such as "${var.image}:2", has no value recorded: the target can't update
// the tag through it, even when the variable value happens to end like the tag.
func TestGetNomadSpecFromFileVariableBeforeTag(t *testing.T) {
	file := filepath.Join(t.TempDir(), "redis.nomad")
	require.NoError(t, os.WriteFile(file, []byte(`variable "image" {
  type    = string
  default = "redis-2"
}

job "redis" {
  group "redis" {
    task "redis" {
      driver = "docker"
      config {
        image = "${var.image}:2"
      }
    }
  }
}
`), 0o600))

	got, err := getNomadDockerSpecFromFile(file)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "redis-2:2", got[0].Value)
	assert.Empty(t, got[0].VariableValue)

	_, ok := nomadTargetPrefix(got[0], "redis-2")
	assert.False(t, ok)
}
