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
					Name: "Bump \"mongo\" Docker compose service image version for \"docker-compose.yaml\"",
					Sources: map[string]source.Config{
						"mongodb": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get latest \"mongo\" Docker Image Tag",
								Kind: "dockerimage",
								Spec: dockerimage.Spec{
									Image:        "mongo",
									Architecture: "amd64",
									TagFilter:    `^\d*(\.\d*){2}$`,
									VersionFilter: version.Filter{
										Kind:    "semver",
										Pattern: ">=6.0.2",
									},
								},
							},
						},
					},
					Targets: map[string]target.Config{
						"mongodb": {
							SourceID: "mongodb",
							ResourceConfig: resource.ResourceConfig{
								Name: "Bump \"mongo\" Docker Image tag for docker compose file \"docker-compose.yaml\"",
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

					Name: "Bump \"ghcr.io/updatecli/updatemonitor\" Docker compose service image version for \"docker-compose.yaml\"",
					Sources: map[string]source.Config{
						"agent": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get latest \"ghcr.io/updatecli/updatemonitor\" Docker Image Tag",
								Kind: "dockerimage",
								Spec: dockerimage.Spec{
									Image:     "ghcr.io/updatecli/updatemonitor",
									TagFilter: `^v\d*(\.\d*){2}$`,
									VersionFilter: version.Filter{
										Kind:    "semver",
										Pattern: ">=v0.1.0",
									},
								},
							},
						},
					},
					Targets: map[string]target.Config{
						"agent": {
							SourceID: "agent",
							ResourceConfig: resource.ResourceConfig{
								Name: "Bump \"ghcr.io/updatecli/updatemonitor\" Docker Image tag for docker compose file \"docker-compose.yaml\"",
								Kind: "yaml",
								Spec: yaml.Spec{
									File: "testdata/docker-compose.yaml",
									Key:  "services.agent.image",
								},
								Transformers: transformer.Transformers{
									transformer.Transformer{
										AddPrefix: "ghcr.io/updatecli/updatemonitor:",
									},
								},
							},
						},
					},
				},
				{
					Name: "Bump \"ghcr.io/updatecli/updatemonitor\" Docker compose service image version for \"docker-compose.yaml\"",
					Sources: map[string]source.Config{
						"server": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get latest \"ghcr.io/updatecli/updatemonitor\" Docker Image Tag",
								Kind: "dockerimage",
								Spec: dockerimage.Spec{
									Image:     "ghcr.io/updatecli/updatemonitor",
									TagFilter: `^v\d*(\.\d*){2}$`,
									VersionFilter: version.Filter{
										Kind:    "semver",
										Pattern: ">=v0.1.0",
									},
								},
							},
						},
					},
					Targets: map[string]target.Config{
						"server": {
							SourceID: "server",
							ResourceConfig: resource.ResourceConfig{
								Name: "Bump \"ghcr.io/updatecli/updatemonitor\" Docker Image tag for docker compose file \"docker-compose.yaml\"",
								Kind: "yaml",
								Spec: yaml.Spec{
									File: "testdata/docker-compose.yaml",
									Key:  "services.server.image",
								},
								Transformers: transformer.Transformers{
									transformer.Transformer{
										AddPrefix: "ghcr.io/updatecli/updatemonitor:",
									},
								},
							},
						},
					},
				},
				{
					Name: "Bump \"ghcr.io/updatecli/updatemonitor-ui\" Docker compose service image version for \"docker-compose.yaml\"",
					Sources: map[string]source.Config{
						"front": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get latest \"ghcr.io/updatecli/updatemonitor-ui\" Docker Image Tag",
								Kind: "dockerimage",
								Spec: dockerimage.Spec{
									Image:     "ghcr.io/updatecli/updatemonitor-ui",
									TagFilter: `^v\d*(\.\d*){2}$`,
									VersionFilter: version.Filter{
										Kind:    "semver",
										Pattern: ">=v0.1.1",
									},
								},
							},
						},
					},
					Targets: map[string]target.Config{
						"front": {
							SourceID: "front",
							ResourceConfig: resource.ResourceConfig{
								Name: "Bump \"ghcr.io/updatecli/updatemonitor-ui\" Docker Image tag for docker compose file \"docker-compose.yaml\"",
								Kind: "yaml",
								Spec: yaml.Spec{
									File: "testdata/docker-compose.yaml",
									Key:  "services.front.image",
								},
								Transformers: transformer.Transformers{
									transformer.Transformer{
										AddPrefix: "ghcr.io/updatecli/updatemonitor-ui:",
									},
								},
							},
						},
					},
				},
				{
					Name: "Bump \"traefik\" Docker compose service image version for \"docker-compose.yaml\"",
					Sources: map[string]source.Config{
						"traefik": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get latest \"traefik\" Docker Image Tag",
								Kind: "dockerimage",
								Spec: dockerimage.Spec{
									Image:     "traefik",
									TagFilter: `^v?\d*(\.\d*){1}$`,
									VersionFilter: version.Filter{
										Kind:    "semver",
										Pattern: ">=v2.9",
									},
								},
							},
						},
					},
					Targets: map[string]target.Config{
						"traefik": {
							SourceID: "traefik",
							ResourceConfig: resource.ResourceConfig{
								Name: "Bump \"traefik\" Docker Image tag for docker compose file \"docker-compose.yaml\"",
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
			// !! Order matter between expected result and docker-compose file
			assert.Equal(t, tt.expectedPipelines, pipelines)

		})
	}

}
