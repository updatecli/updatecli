package dockercompose

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchFiles(t *testing.T) {

	gotFiles, err := searchDockerComposeFiles(
		"testdata/", DefaultFilePattern[:])
	if err != nil {
		t.Errorf("%s\n", err)
	}
	expectedFiles := []string{
		"testdata/docker-compose.2.yaml",
		"testdata/docker-compose.yaml",
	}

	assert.Equal(t, expectedFiles, gotFiles)
}

func TestGetDockerComposeSpec(t *testing.T) {

	gotDockerComposeSpec, err := getDockerComposeData(
		"testdata/docker-compose.yaml")
	if err != nil {
		t.Errorf("%s\n", err)
	}

	expectedDockerComposeSpec := DockerComposeSpec{
		Services: map[string]Service{
			"agent": {
				Image: "ghcr.io/updatecli/updatemonitor:v0.1.0",
			},
			"front": {
				Image: "ghcr.io/updatecli/updatemonitor-ui:v0.1.1",
			},
			"mongodb": {
				Image: "mongo:6.0.2",
			},
			"server": {
				Image: "ghcr.io/updatecli/updatemonitor:v0.1.0",
			},
			"traefik": {
				Image: "traefik:v2.9",
			},
		},
	}

	assert.Equal(t, expectedDockerComposeSpec, *gotDockerComposeSpec)
}
