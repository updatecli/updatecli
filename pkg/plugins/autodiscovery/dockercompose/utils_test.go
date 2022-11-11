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

	expectedDockerComposeSpec := dockerComposeSpec{
		Services: map[string]service{
			"agent-2": {
				Image: "ghcr.io/updatecli/updatemonitor@sha256:2b3436753ba2c671c382777d8980ed6ea7350851e71bff0326dc776c0c9e2334",
			},
			"mongodb": {
				Image: "mongo:6.0.2",
			},
			"agent": {
				Image: "ghcr.io/updatecli/updatemonitor:v0.1.0",
			},
			"server": {
				Image: "ghcr.io/updatecli/updatemonitor:v0.1.0",
			},
			"front": {
				Image: "ghcr.io/updatecli/updatemonitor-ui:v0.1.1",
			},
			"traefik": {
				Image: "traefik:v2.9",
			},
		},
	}

	assert.Equal(t, expectedDockerComposeSpec, *gotDockerComposeSpec)
}
