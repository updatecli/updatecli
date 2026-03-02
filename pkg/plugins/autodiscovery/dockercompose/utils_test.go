package dockercompose

import (
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
			rootDir:      "testdata/",
			filePatterns: DefaultFilePattern,
			expectedFiles: []string{
				"testdata/compose.yaml",
				"testdata/docker-compose.yaml",
			},
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			gotFiles, err := searchDockerComposeFiles(tt.rootDir, tt.filePatterns)

			if tt.expectedErr != nil {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedFiles, gotFiles)
		})
	}
}

func TestGetDockerComposeSpec(t *testing.T) {
	testdata := []struct {
		name             string
		filename         string
		expectedServices dockercomposeServicesList
		expectedErr      bool
	}{
		{
			name:     "Case from testdata with sorted services",
			filename: "testdata/docker-compose.yaml",
			expectedServices: dockercomposeServicesList{
				dockerComposeService{
					Name: "jenkins-lts",
					Spec: dockerComposeServiceSpec{
						Image: "jenkinsci/jenkins:2.150.1-alpine@sha256:1c626c48537129ad375832cd961d6ecd4eb8aa3866f24ab23c57c21bd6d81f3f",
					},
				},
				dockerComposeService{
					Name: "jenkins-weekly",
					Spec: dockerComposeServiceSpec{
						Image:    "jenkinsci/jenkins:2.254-alpine",
						Platform: "linux/amd64",
					},
				},
				dockerComposeService{
					Name: "postgres",
					Spec: dockerComposeServiceSpec{
						Image:    "${IMAGE_PREFIX}/safeline-postgres${ARCH_SUFFIX}:15.2",
						Platform: "",
					},
				},
			},
		},
		{
			name:             "Case with no services found (not a Docker Compose Yaml)",
			filename:         "testdata/not-compose.yaml",
			expectedServices: dockercomposeServicesList{},
		},
		{
			name:        "Case with a non-YAML file",
			filename:    "testdata/not-yaml.txt",
			expectedErr: true,
		},
		{
			name:        "Case with a nonexistent file",
			filename:    "does-not-exist.yaml",
			expectedErr: true,
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			gotDockerComposeServices, err := getDockerComposeSpecFromFile(tt.filename)

			if tt.expectedErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedServices, gotDockerComposeServices)
		})
	}
}
