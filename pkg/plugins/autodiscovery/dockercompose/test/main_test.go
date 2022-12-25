package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/dockercompose"
)

func TestDiscoverManifests(t *testing.T) {

	// Disable condition testing with running short test
	if testing.Short() {
		return
	}

	testdata := []struct {
		name              string
		rootDir           string
		expectedPipelines []string
	}{
		{
			name:    "Scenario 1",
			rootDir: "testdata",
			expectedPipelines: []string{`name: 'Bump Docker image tag for jenkinsci/jenkins'
sources:
  jenkinsci/jenkins:
    name: '[jenkinsci/jenkins] Get latest Docker image tag'
    kind: 'dockerimage'
    spec:
      image: 'jenkinsci/jenkins'
      tagFilter: '^\d*(\.\d*){1}-alpine$'
      versionFilter:
        kind: 'semver'
        pattern: '>=2.254-alpine'
targets:
  jenkinsci/jenkins:
    name: '[jenkinsci/jenkins] Bump Docker image tag in "docker-compose.yaml"'
    kind: 'yaml'
    spec:
      file: 'docker-compose.yaml'
      key: 'services.jenkins-weekly.image'
    sourceid: 'jenkinsci/jenkins'
    transformers:
      - addprefix: 'jenkinsci/jenkins:'
`,
			},
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {

			composefile, err := dockercompose.New(
				dockercompose.Spec{
					RootDir: tt.rootDir,
				}, "", "")
			require.NoError(t, err)

			var pipelines []string
			rawPipelines, err := composefile.DiscoverManifests()
			require.NoError(t, err)

			for i := range rawPipelines {
				// We expect manifest generated by the autodiscovery to use the yaml syntax
				pipelines = append(pipelines, string(rawPipelines[i]))
				assert.Equal(t, tt.expectedPipelines[i], pipelines[i])
			}

		})
	}

}
