package test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/source"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
	autodiscoveryDockerfile "github.com/updatecli/updatecli/pkg/plugins/autodiscovery/dockerfile"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerfile"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerimage"
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
			rootDir: "testdata/updatecli-action",
			expectedPipelines: []config.Spec{
				{
					Name: "Bump Docker image tag for \"updatecli/updatecli\"",
					Sources: map[string]source.Config{
						"updatecli/updatecli": {
							ResourceConfig: resource.ResourceConfig{
								Name: "[updatecli/updatecli] Get latest Docker image tag",
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
								Name: "[updatecli/updatecli] Bump Docker image tag in \"Dockerfile\"",
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
					Name: "Bump Docker image tag for \"jenkins/jenkins\"",
					Sources: map[string]source.Config{
						"jenkins/jenkins": {
							ResourceConfig: resource.ResourceConfig{
								Name: "[jenkins/jenkins] Get latest Docker image tag",
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
								Name: "[jenkins/jenkins] Bump Docker image tag in \"Dockerfile\"",
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
			dockerfile, err := autodiscoveryDockerfile.New(
				autodiscoveryDockerfile.Spec{
					RootDir: tt.rootDir,
				}, "", "")
			require.NoError(t, err)

			pipelines, err := dockerfile.DiscoverManifests()
			require.NoError(t, err)

			for i := range tt.expectedPipelines {
				test.AssertConfigSpecEqualByteArray(t, &tt.expectedPipelines[i], string(pipelines[i]))
			}
		})
	}

}
