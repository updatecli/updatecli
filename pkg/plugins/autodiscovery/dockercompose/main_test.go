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

					Name: "Bump \"mongo\" Docker compose service image version for \"testdata/docker-compose.yaml\"",
					Sources: map[string]source.Config{
						"mongodb": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get latest \"mongo\" Docker Image Tag",
								Kind: "dockerimage",
								Spec: dockerimage.Spec{
									Image: "mongo",
									VersionFilter: version.Filter{
										Kind: "semver",
									},
								},
							},
						},
					},
					Targets: map[string]target.Config{
						"mongodb": {
							SourceID: "mongodb",
							ResourceConfig: resource.ResourceConfig{
								Name: "Bump \"mongo\" Docker Image tag for docker compose file \"testdata/docker-compose.yaml\"",
								Kind: "yaml",
								Spec: yaml.Spec{
									File: "testdata/docker-compose.yaml",
									Key:  "services.mongodb.image",
								},
								Transformers: transformer.Transformers{
									transformer.Transformer{
										AddPrefix: "mongo:",
									},
								},
							},
						},
					},
				},
				{

					Name: "Bump \"traefik\" Docker compose service image version for \"testdata/docker-compose.yaml\"",
					Sources: map[string]source.Config{
						"traefik": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get latest \"traefik\" Docker Image Tag",
								Kind: "dockerimage",
								Spec: dockerimage.Spec{
									Image: "traefik",
									VersionFilter: version.Filter{
										Kind: "semver",
									},
								},
							},
						},
					},
					Targets: map[string]target.Config{
						"traefik": {
							SourceID: "traefik",
							ResourceConfig: resource.ResourceConfig{
								Name: "Bump \"traefik\" Docker Image tag for docker compose file \"testdata/docker-compose.yaml\"",
								Kind: "yaml",
								Spec: yaml.Spec{
									File: "testdata/docker-compose.yaml",
									Key:  "services.traefik.image",
								},
								Transformers: transformer.Transformers{
									transformer.Transformer{
										AddPrefix: "traefik:",
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
			assert.Equal(t, tt.expectedPipelines, pipelines)

		})
	}

}
