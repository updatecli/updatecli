package dockerfile

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/config"
	discoveryConfig "github.com/updatecli/updatecli/pkg/core/pipeline/autodiscovery/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/source"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerfile"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerimage"
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
			rootDir: "testdata/updatecli-action",
			expectedPipelines: []config.Spec{
				{
					Name: "Bump Docker Image Tag for \"updatecli/updatecli\"",
					Sources: map[string]source.Config{
						"updatecli/updatecli": {
							ResourceConfig: resource.ResourceConfig{
								Name: "[updatecli/updatecli] Get latest Docker Image Tag",
								Kind: "dockerimage",
								Spec: dockerimage.Spec{
									Image:     "updatecli/updatecli",
									TagFilter: `^v\d*(\.\d*){2}$`,
									VersionFilter: version.Filter{
										Kind:    "semver",
										Pattern: ">=v0.25.0",
									},
								},
							},
						},
					},
					Targets: map[string]target.Config{
						"updatecli/updatecli": {
							SourceID: "updatecli/updatecli",
							ResourceConfig: resource.ResourceConfig{
								Name: "[updatecli/updatecli] Bump Docker Image tag in \"Dockerfile\"",
								Kind: "dockerfile",
								Spec: dockerfile.Spec{
									File: "Dockerfile",
									Instruction: map[string]string{
										"keyword": "FROM",
										"matcher": "updatecli/updatecli",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:    "Scenario 2: arg with suffix",
			rootDir: "testdata/jenkins",
			expectedPipelines: []config.Spec{
				{
					Name: "Bump Docker Image Tag for \"jenkins/jenkins\"",
					Sources: map[string]source.Config{
						"jenkins/jenkins": {
							ResourceConfig: resource.ResourceConfig{
								Name: "[jenkins/jenkins] Get latest Docker Image Tag",
								Kind: "dockerimage",
								Spec: dockerimage.Spec{
									Image:     "jenkins/jenkins",
									TagFilter: `^\d*(\.\d*){2}-lts$`,
									VersionFilter: version.Filter{
										Kind:    "semver",
										Pattern: ">=2.235.1-lts",
									},
								},
							},
						},
					},
					Targets: map[string]target.Config{
						"jenkins/jenkins": {
							SourceID: "jenkins/jenkins",
							ResourceConfig: resource.ResourceConfig{
								Name: "[jenkins/jenkins] Bump Docker Image tag in \"Dockerfile\"",
								Kind: "dockerfile",
								Spec: dockerfile.Spec{
									File: "Dockerfile",
									Instruction: map[string]string{
										"keyword": "ARG",
										"matcher": "jenkins_version",
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
