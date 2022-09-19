package helmfile

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/config"
	discoveryConfig "github.com/updatecli/updatecli/pkg/core/pipeline/autodiscovery/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/condition"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/source"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
	"github.com/updatecli/updatecli/pkg/plugins/resources/helm"
	"github.com/updatecli/updatecli/pkg/plugins/resources/yaml"
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

					Name: "Bump \"datadog\" Helm Chart version for Helmfile \"testdata/helmfile.d/cik8s.yaml\"",
					Sources: map[string]source.Config{
						"datadog": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get latest \"datadog\" Helm Chart Version",
								Kind: "helmchart",
								Spec: helm.Spec{
									Name: "datadog",
									URL:  "https://helm.datadoghq.com",
								},
							},
						},
					},
					Conditions: map[string]condition.Config{
						"datadog": {
							DisableSourceInput: true,
							ResourceConfig: resource.ResourceConfig{
								Name: "Ensure release \"datadog\" is specified",
								Kind: "yaml",
								Spec: yaml.Spec{
									File:  "testdata/helmfile.d/cik8s.yaml",
									Key:   "releases[0].chart",
									Value: "datadog/datadog",
								},
							},
						},
					},
					Targets: map[string]target.Config{

						"datadog": {
							SourceID: "datadog",
							ResourceConfig: resource.ResourceConfig{
								Name: "Bump \"datadog\" Helm Chart Version for Helmfile \"testdata/helmfile.d/cik8s.yaml\"",
								Kind: "yaml",
								Spec: yaml.Spec{
									File: "testdata/helmfile.d/cik8s.yaml",
									Key:  "releases[0].version",
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
