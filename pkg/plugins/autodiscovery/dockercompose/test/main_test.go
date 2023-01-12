package test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/source"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
	"github.com/updatecli/updatecli/pkg/core/transformer"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/dockercompose"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerimage"
	"github.com/updatecli/updatecli/pkg/plugins/resources/yaml"
	"github.com/updatecli/updatecli/pkg/plugins/utils/test"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func TestDiscoverManifests(t *testing.T) {

	// Disable condition testing with running short test
	if testing.Short() {
		return
	}

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
					Name: "Bump Docker image tag for \"jenkinsci/jenkins\"",
					Sources: map[string]source.Config{
						"jenkins-lts": {
							ResourceConfig: resource.ResourceConfig{
								Name: "[jenkinsci/jenkins] Get latest Docker image tag",
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
								Name: "[jenkinsci/jenkins] Bump Docker image tag in \"docker-compose.yaml\"",
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
					Name: "Bump Docker image tag for \"jenkinsci/jenkins\"",
					Sources: map[string]source.Config{
						"jenkins-weekly": {
							ResourceConfig: resource.ResourceConfig{
								Name: "[jenkinsci/jenkins] Get latest Docker image tag",
								Kind: "dockerimage",
								Spec: dockerimage.Spec{
									Image:     "jenkinsci/jenkins",
									TagFilter: `^\d*(\.\d*){1}-alpine$`,
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
								Name: "[jenkinsci/jenkins] Bump Docker image tag in \"docker-compose.yaml\"",
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
			composefile, err := dockercompose.New(
				dockercompose.Spec{
					RootDir: tt.rootDir,
				}, "", "")
			require.NoError(t, err)

			pipelines, err := composefile.DiscoverManifests()
			require.NoError(t, err)

			require.NoError(t, err)
			for i := range tt.expectedPipelines {
				test.AssertConfigSpecEqualByteArray(t, &tt.expectedPipelines[i], string(pipelines[i]))
			}

		})
	}

}
