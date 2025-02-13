package helmfile

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscoverManifests(t *testing.T) {
	testdata := []struct {
		name              string
		rootDir           string
		expectedPipelines []string
	}{
		{
			name:    "Scenario 1",
			rootDir: "testdata",
			expectedPipelines: []string{`name: 'Bump "datadog" Helm Chart version for Helmfile "helmfile.d/cik8s.yaml"'
sources:
  datadog:
    name: 'Get latest "datadog" Helm Chart version'
    kind: 'helmchart'
    spec:
      name: 'datadog'
      url: 'https://helm.datadoghq.com'
      versionfilter:
        kind: 'semver'
        pattern: '*'
conditions:
  datadog:
    name: 'Ensure release "datadog" is specified for Helmfile "helmfile.d/cik8s.yaml"'
    kind: 'yaml'
    spec:
      file: 'testdata/helmfile.d/cik8s.yaml'
      key: '$.releases[0].chart'
      value: 'datadog/datadog'
    disablesourceinput: true
targets:
  datadog:
    name: 'deps(helmfile): update "datadog" Helm Chart version to {{ source "datadog"}}'
    kind: 'yaml'
    spec:
      file: 'testdata/helmfile.d/cik8s.yaml'
      key: '$.releases[0].version'
    sourceid: 'datadog'
`, `name: 'Bump "docker-registry-secrets" Helm Chart version for Helmfile "helmfile.d/cik8s.yaml"'
sources:
  docker-registry-secrets:
    name: 'Get latest "docker-registry-secrets" Helm Chart version'
    kind: 'helmchart'
    spec:
      name: 'docker-registry-secrets'
      url: 'https://jenkins-infra.github.io/helm-charts'
      versionfilter:
        kind: 'semver'
        pattern: '*'
conditions:
  docker-registry-secrets:
    name: 'Ensure release "docker-registry-secrets" is specified for Helmfile "helmfile.d/cik8s.yaml"'
    kind: 'yaml'
    spec:
      file: 'testdata/helmfile.d/cik8s.yaml'
      key: '$.releases[1].chart'
      value: 'jenkins-infra/docker-registry-secrets'
    disablesourceinput: true
targets:
  docker-registry-secrets:
    name: 'deps(helmfile): update "docker-registry-secrets" Helm Chart version to {{ source "docker-registry-secrets"}}'
    kind: 'yaml'
    spec:
      file: 'testdata/helmfile.d/cik8s.yaml'
      key: '$.releases[1].version'
    sourceid: 'docker-registry-secrets'
`, `name: 'Bump "myOCIChart" Helm Chart version for Helmfile "helmfile.d/cik8s.yaml"'
sources:
  myOCIChart:
    name: 'Get latest "myOCIChart" Helm Chart version'
    kind: 'helmchart'
    spec:
      name: 'myOCIChart'
      url: 'oci://myregistry.azurecr.io'
      versionfilter:
        kind: 'semver'
        pattern: '*'
conditions:
  myOCIChart:
    name: 'Ensure release "myOCIChart" is specified for Helmfile "helmfile.d/cik8s.yaml"'
    kind: 'yaml'
    spec:
      file: 'testdata/helmfile.d/cik8s.yaml'
      key: '$.releases[3].chart'
      value: 'myOCIRegistry/myOCIChart'
    disablesourceinput: true
targets:
  myOCIChart:
    name: 'deps(helmfile): update "myOCIChart" Helm Chart version to {{ source "myOCIChart"}}'
    kind: 'yaml'
    spec:
      file: 'testdata/helmfile.d/cik8s.yaml'
      key: '$.releases[3].version'
    sourceid: 'myOCIChart'
`,
			},
		},
	}

	for _, tt := range testdata {

		t.Run(tt.name, func(t *testing.T) {
			helmfile, err := New(
				Spec{}, tt.rootDir, "", "")
			require.NoError(t, err)

			pipelines, err := helmfile.DiscoverManifests()
			require.NoError(t, err)

			require.Equal(t, len(tt.expectedPipelines), len(pipelines))

			for i := range tt.expectedPipelines {
				assert.Equal(t, tt.expectedPipelines[i], string(pipelines[i]))
			}

		})
	}

}
