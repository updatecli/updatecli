package dockercompose

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/config"
	discoveryConfig "github.com/updatecli/updatecli/pkg/core/pipeline/autodiscovery/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/source"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
	"github.com/updatecli/updatecli/pkg/core/transformer"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerimage"
	"github.com/updatecli/updatecli/pkg/plugins/resources/yaml"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func TestDiscoverManifests(t *testing.T) {
	testdata := []struct {
		name              string
		rootDir           string
		expectedPipelines []config.Spec
	}{
		{
			name:    "Scenario 1",
			rootDir: "testdata",
			expectedPipelines: []config.Spec{
				{
					Name: "Bump Docker Image Tag for \"jenkinsci/jenkins\"",
					Sources: map[string]source.Config{
						"jenkins-lts": {
							ResourceConfig: resource.ResourceConfig{
								Name: "[jenkinsci/jenkins] Get latest Docker Image Tag",
								Kind: "dockerimage",
								Spec: dockerimage.Spec{
									Image:     "jenkinsci/jenkins",
									TagFilter: `^\d*(\.\d*){2}-alpine$`,
									VersionFilter: version.Filter{
										Kind:    "semver",
										Pattern: ">=2.150.1-alpine",
									},
								},
							},
						},
					},
					Targets: map[string]target.Config{
						"jenkins-lts": {
							SourceID: "jenkins-lts",
							ResourceConfig: resource.ResourceConfig{
								Name: "[jenkinsci/jenkins] Bump Docker Image tag in \"docker-compose.yaml\"",
								Kind: "yaml",
								Spec: yaml.Spec{
									File: "docker-compose.yaml",
									Key:  "services.jenkins-lts.image",
								},
								Transformers: transformer.Transformers{
									transformer.Transformer{
										AddPrefix: "jenkinsci/jenkins:",
									},
								},
							},
						},
					},
				},
				{
					Name: "Bump Docker Image Tag for \"jenkinsci/jenkins\"",
					Sources: map[string]source.Config{
						"jenkins-weekly": {
							ResourceConfig: resource.ResourceConfig{
								Name: "[jenkinsci/jenkins] Get latest Docker Image Tag",
								Kind: "dockerimage",
								Spec: dockerimage.Spec{
									Architecture: "amd64",
									Image:        "jenkinsci/jenkins",
									TagFilter:    `^\d*(\.\d*){1}-alpine$`,
									VersionFilter: version.Filter{
										Kind:    "semver",
										Pattern: ">=2.254-alpine",
									},
								},
							},
						},
					},
					Targets: map[string]target.Config{
						"jenkins-weekly": {
							SourceID: "jenkins-weekly",
							ResourceConfig: resource.ResourceConfig{
								Name: "[jenkinsci/jenkins] Bump Docker Image tag in \"docker-compose.yaml\"",
								Kind: "yaml",
								Spec: yaml.Spec{
									File: "docker-compose.yaml",
									Key:  "services.jenkins-weekly.image",
								},
								Transformers: transformer.Transformers{
									transformer.Transformer{
										AddPrefix: "jenkinsci/jenkins:",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range testdata {

		t.Run(tt.name, func(t *testing.T) {
			helmfile, err := New(
				Spec{
					RootDir: tt.rootDir,
				}, "")

			require.NoError(t, err)

			pipelines, err := helmfile.DiscoverManifests(discoveryConfig.Input{})

			require.NoError(t, err)
			// !! Order matter between expected result and docker-compose file
			assert.Equal(t, tt.expectedPipelines, pipelines)

		})
	}

}
