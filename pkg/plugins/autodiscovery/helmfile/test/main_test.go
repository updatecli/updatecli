package helmfile

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/condition"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/source"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/helmfile"
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

					Name: "Bump \"datadog\" Helm Chart version for Helmfile \"helmfile.d/cik8s.yaml\"",
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
								Name: "Ensure release \"datadog\" is specified for Helmfile \"helmfile.d/cik8s.yaml\"",
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
								Name: "Bump \"datadog\" Helm Chart Version for Helmfile \"helmfile.d/cik8s.yaml\"",
								Kind: "yaml",
								Spec: yaml.Spec{
									File: "testdata/helmfile.d/cik8s.yaml",
									Key:  "releases[0].version",
								},
							},
						},
					},
				},
				{

					Name: "Bump \"docker-registry-secrets\" Helm Chart version for Helmfile \"helmfile.d/cik8s.yaml\"",
					Sources: map[string]source.Config{
						"docker-registry-secrets": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get latest \"docker-registry-secrets\" Helm Chart Version",
								Kind: "helmchart",
								Spec: helm.Spec{
									Name: "docker-registry-secrets",
									URL:  "https://jenkins-infra.github.io/helm-charts",
								},
							},
						},
					},
					Conditions: map[string]condition.Config{
						"docker-registry-secrets": {
							DisableSourceInput: true,
							ResourceConfig: resource.ResourceConfig{
								Name: "Ensure release \"docker-registry-secrets\" is specified for Helmfile \"helmfile.d/cik8s.yaml\"",
								Kind: "yaml",
								Spec: yaml.Spec{
									File:  "testdata/helmfile.d/cik8s.yaml",
									Key:   "releases[1].chart",
									Value: "jenkins-infra/docker-registry-secrets",
								},
							},
						},
					},
					Targets: map[string]target.Config{
						"docker-registry-secrets": {
							SourceID: "docker-registry-secrets",
							ResourceConfig: resource.ResourceConfig{
								Name: "Bump \"docker-registry-secrets\" Helm Chart Version for Helmfile \"helmfile.d/cik8s.yaml\"",
								Kind: "yaml",
								Spec: yaml.Spec{
									File: "testdata/helmfile.d/cik8s.yaml",
									Key:  "releases[1].version",
								},
							},
						},
					},
				},
				{

					Name: "Bump \"myOCIChart\" Helm Chart version for Helmfile \"helmfile.d/cik8s.yaml\"",
					Sources: map[string]source.Config{
						"myOCIChart": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get latest \"myOCIChart\" Helm Chart Version",
								Kind: "helmchart",
								Spec: helm.Spec{
									Name: "myOCIChart",
									URL:  "oci://myregistry.azurecr.io",
								},
							},
						},
					},
					Conditions: map[string]condition.Config{
						"myOCIChart": {
							DisableSourceInput: true,
							ResourceConfig: resource.ResourceConfig{
								Name: "Ensure release \"myOCIChart\" is specified for Helmfile \"helmfile.d/cik8s.yaml\"",
								Kind: "yaml",
								Spec: yaml.Spec{
									File:  "testdata/helmfile.d/cik8s.yaml",
									Key:   "releases[3].chart",
									Value: "myOCIRegistry/myOCIChart",
								},
							},
						},
					},
					Targets: map[string]target.Config{

						"myOCIChart": {
							SourceID: "myOCIChart",
							ResourceConfig: resource.ResourceConfig{
								Name: "Bump \"myOCIChart\" Helm Chart Version for Helmfile \"helmfile.d/cik8s.yaml\"",
								Kind: "yaml",
								Spec: yaml.Spec{
									File: "testdata/helmfile.d/cik8s.yaml",
									Key:  "releases[3].version",
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
			helmfile, err := helmfile.New(
				helmfile.Spec{
					RootDir: tt.rootDir,
				}, "", "")

			require.NoError(t, err)

			pipelines, err := helmfile.DiscoverManifests()

			require.NoError(t, err)
			assert.Equal(t, tt.expectedPipelines, pipelines)

		})
	}

}
