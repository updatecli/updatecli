package dockercompose

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchFiles(t *testing.T) {

	gotFiles, err := searchDockerComposeFiles(
		"testdata/", DefaultFileMatch[:])
	if err != nil {
		t.Errorf("%s\n", err)
	}
	expectedFiles := []string{
		"testdata/docker-compose.yaml",
	}

	assert.Equal(t, expectedFiles, gotFiles)
}

func TestGetDockerComposeSpec(t *testing.T) {

	gotDockerComposeSpec, err := getDockerComposeSpecFromFile(
		"testdata/docker-compose.yaml")
	if err != nil {
		t.Errorf("%s\n", err)
	}

	expectedDockerComposeSpec := dockerComposeSpec{
		Services: map[string]service{
			"jenkins-lts": {
				Image: "jenkinsci/jenkins:2.150.1-alpine",
			},
			"jenkins-weekly": {
				Image:    "jenkinsci/jenkins:2.254-alpine",
				Platform: "linux/amd64",
			},
		},
	}

	assert.Equal(t, expectedDockerComposeSpec, *gotDockerComposeSpec)
}
