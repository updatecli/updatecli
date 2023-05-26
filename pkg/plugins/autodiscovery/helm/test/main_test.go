package helm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/fleet"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/helm"
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
			rootDir: "testdata-1/chart",
			expectedPipelines: []string{`name: 'Bump dependency "minio" for Helm chart "epinio"'
sources:
  minio:
    name: 'Get latest "minio" Helm chart version'
    kind: 'helmchart'
    spec:
      name: 'minio'
      url: 'https://charts.min.io/'
      versionfilter:
        kind: 'semver'
        pattern: '*'
conditions:
  minio:
    name: 'Ensure Helm chart dependency "minio" is specified'
    kind: 'yaml'
    spec:
      file: 'epinio/Chart.yaml'
      key: '$.dependencies[0].name'
      value: 'minio'
    disablesourceinput: true
targets:
  minio:
    name: 'Bump Helm chart dependency "minio" for Helm chart "epinio"'
    kind: 'helmchart'
    spec:
      file: 'Chart.yaml'
      key: '$.dependencies[0].version'
      name: 'epinio'
      versionincrement: 'minor'
    sourceid: 'minio'
`, `name: 'Bump dependency "kubed" for Helm chart "epinio"'
sources:
  kubed:
    name: 'Get latest "kubed" Helm chart version'
    kind: 'helmchart'
    spec:
      name: 'kubed'
      url: 'https://charts.appscode.com/stable'
      versionfilter:
        kind: 'semver'
        pattern: '*'
conditions:
  kubed:
    name: 'Ensure Helm chart dependency "kubed" is specified'
    kind: 'yaml'
    spec:
      file: 'epinio/Chart.yaml'
      key: '$.dependencies[1].name'
      value: 'kubed'
    disablesourceinput: true
targets:
  kubed:
    name: 'Bump Helm chart dependency "kubed" for Helm chart "epinio"'
    kind: 'helmchart'
    spec:
      file: 'Chart.yaml'
      key: '$.dependencies[1].version'
      name: 'epinio'
      versionincrement: 'minor'
    sourceid: 'kubed'
`, `name: 'Bump dependency "epinio-ui" for Helm chart "epinio"'
sources:
  epinio-ui:
    name: 'Get latest "epinio-ui" Helm chart version'
    kind: 'helmchart'
    spec:
      name: 'epinio-ui'
      url: 'https://epinio.github.io/helm-charts'
      versionfilter:
        kind: 'semver'
        pattern: '*'
conditions:
  epinio-ui:
    name: 'Ensure Helm chart dependency "epinio-ui" is specified'
    kind: 'yaml'
    spec:
      file: 'epinio/Chart.yaml'
      key: '$.dependencies[2].name'
      value: 'epinio-ui'
    disablesourceinput: true
targets:
  epinio-ui:
    name: 'Bump Helm chart dependency "epinio-ui" for Helm chart "epinio"'
    kind: 'helmchart'
    spec:
      file: 'Chart.yaml'
      key: '$.dependencies[2].version'
      name: 'epinio'
      versionincrement: 'minor'
    sourceid: 'epinio-ui'
`, `name: 'Bump Docker Image "epinioteam/epinio-ui-qa" for Helm chart "epinio"'
sources:
  epinioteam_epinio-ui-qa:
    name: 'Get latest "epinioteam/epinio-ui-qa" Container tag'
    kind: 'dockerimage'
    spec:
      image: 'epinioteam/epinio-ui-qa'
      versionfilter:
        kind: 'semver'
        pattern: '*'
conditions:
  epinioteam_epinio-ui-qa:
    name: 'Ensure container repository "epinioteam/epinio-ui-qa" is specified'
    kind: 'yaml'
    spec:
      file: 'epinio/values.yaml'
      key: '$.images.ui.repository'
      value: 'epinioteam/epinio-ui-qa'
    disablesourceinput: true
targets:
  epinioteam_epinio-ui-qa:
    name: 'Bump container image tag for image "epinioteam/epinio-ui-qa" in chart "epinio"'
    kind: 'helmchart'
    spec:
      file: 'values.yaml'
      key: '$.images.ui.tag'
      name: 'epinio'
      versionincrement: 'minor'
    sourceid: 'epinioteam/epinio-ui-qa'
`, `name: 'Bump Docker Image "splatform/epinio-server" for Helm chart "epinio"'
sources:
  splatform_epinio-server:
    name: 'Get latest "splatform/epinio-server" Container tag'
    kind: 'dockerimage'
    spec:
      image: 'splatform/epinio-server'
      versionfilter:
        kind: 'semver'
        pattern: '*'
conditions:
  splatform_epinio-server:
    name: 'Ensure container repository "splatform/epinio-server" is specified'
    kind: 'yaml'
    spec:
      file: 'epinio/values.yaml'
      key: '$.image.repository'
      value: 'splatform/epinio-server'
    disablesourceinput: true
targets:
  splatform_epinio-server:
    name: 'Bump container image tag for image "splatform/epinio-server" in chart "epinio"'
    kind: 'helmchart'
    spec:
      file: 'values.yaml'
      key: '$.image.tag'
      name: 'epinio'
      versionincrement: 'minor'
    sourceid: 'splatform/epinio-server'
`},
		},
		{
			name:    "Test the tag update for images referenced in the cart",
			rootDir: "testdata-2/chart",
			expectedPipelines: []string{`name: 'Bump Docker image "epinio/epinio-server" for Helm chart "sample"'
sources:
  epinio_epinio-server:
    name: 'Get latest "epinio/epinio-server" container tag'
    kind: 'dockerimage'
    spec:
      image: 'epinio/epinio-server'
      tagfilter: '^v\d*(\.\d*){2}$'
      versionFilter:
        kind: 'semver'
        pattern: '>=v1.8.0'
conditions:
  epinio_epinio-server-repository:
    disablesourceinput: true
    name: 'Ensure container repository "epinio/epinio-server" is specified'
    kind: 'yaml'
    spec:
      file: 'sample/values.yaml'
      key: '$.image.repository'
      value: 'epinio/epinio-server'
targets:
  epinio_epinio-server:
    name: 'Bump container image tag for image "epinio/epinio-server" in chart "sample"'
    kind: 'helmchart'
    spec:
      file: 'values.yaml'
      name: 'sample'
      key: '$.image.tag'
      VersionIncrement: 'minor'
    sourceid: 'epinio_epinio-server'
`,
				`name: 'Bump Docker image "epinio/epinio-ui" for Helm chart "sample"'
sources:
  epinio_epinio-ui:
    name: 'Get latest "epinio/epinio-ui" container tag'
    kind: 'dockerimage'
    spec:
      image: 'epinio/epinio-ui'
      tagfilter: '^v\d*(\.\d*){2}$'
      versionFilter:
        kind: 'semver'
        pattern: '>=v1.8.0'
conditions:
  epinio_epinio-ui-repository:
    disablesourceinput: true
    name: 'Ensure container repository "epinio/epinio-ui" is specified'
    kind: 'yaml'
    spec:
      file: 'sample/values.yaml'
      key: '$.images.ui.repository'
      value: 'epinio/epinio-ui'
targets:
  epinio_epinio-ui:
    name: 'Bump container image tag for image "epinio/epinio-ui" in chart "sample"'
    kind: 'helmchart'
    spec:
      file: 'values.yaml'
      name: 'sample'
      key: '$.images.ui.tag'
      VersionIncrement: 'minor'
    sourceid: 'epinio_epinio-ui'
`},
		},
		{
			name:    "Test the tag update for images referenced in the cart including the registry",
			rootDir: "testdata-3/chart",
			expectedPipelines: []string{`name: 'Bump Docker image "ghcr.io/epinio/epinio-server" for Helm chart "sample"'
sources:
  ghcr.io_epinio_epinio-server:
    name: 'Get latest "ghcr.io/epinio/epinio-server" container tag'
    kind: 'dockerimage'
    spec:
      image: 'ghcr.io/epinio/epinio-server'
      tagfilter: '^v\d*(\.\d*){2}$'
      versionFilter:
        kind: 'semver'
        pattern: '>=v1.8.0'
conditions:
  ghcr.io_epinio_epinio-server-registry:
    disablesourceinput: true
    name: 'Ensure container registry "ghcr.io" is specified'
    kind: 'yaml'
    spec:
      file: 'sample/values.yaml'
      key: '$.image.registry'
      value: 'ghcr.io'
  ghcr.io_epinio_epinio-server-repository:
    disablesourceinput: true
    name: 'Ensure container repository "epinio/epinio-server" is specified'
    kind: 'yaml'
    spec:
      file: 'sample/values.yaml'
      key: '$.image.repository'
      value: 'epinio/epinio-server'
targets:
  ghcr.io_epinio_epinio-server:
    name: 'Bump container image tag for image "ghcr.io/epinio/epinio-server" in chart "sample"'
    kind: 'helmchart'
    spec:
      file: 'values.yaml'
      name: 'sample'
      key: '$.image.tag'
      VersionIncrement: 'minor'
    sourceid: 'ghcr.io_epinio_epinio-server'
`,
				`name: 'Bump Docker image "ghcr.io/epinio/epinio-ui" for Helm chart "sample"'
sources:
  ghcr.io_epinio_epinio-ui:
    name: 'Get latest "ghcr.io/epinio/epinio-ui" container tag'
    kind: 'dockerimage'
    spec:
      image: 'ghcr.io/epinio/epinio-ui'
      tagfilter: '^v\d*(\.\d*){2}$'
      versionFilter:
        kind: 'semver'
        pattern: '>=v1.8.0'
conditions:
  ghcr.io_epinio_epinio-ui-registry:
    disablesourceinput: true
    name: 'Ensure container registry "ghcr.io" is specified'
    kind: 'yaml'
    spec:
      file: 'sample/values.yaml'
      key: '$.images.ui.registry'
      value: 'ghcr.io'
  ghcr.io_epinio_epinio-ui-repository:
    disablesourceinput: true
    name: 'Ensure container repository "epinio/epinio-ui" is specified'
    kind: 'yaml'
    spec:
      file: 'sample/values.yaml'
      key: '$.images.ui.repository'
      value: 'epinio/epinio-ui'
targets:
  ghcr.io_epinio_epinio-ui:
    name: 'Bump container image tag for image "ghcr.io/epinio/epinio-ui" in chart "sample"'
    kind: 'helmchart'
    spec:
      file: 'values.yaml'
      name: 'sample'
      key: '$.images.ui.tag'
      VersionIncrement: 'minor'
    sourceid: 'ghcr.io_epinio_epinio-ui'
`},
		},
	}

	for _, tt := range testdata {

		t.Run(tt.name, func(t *testing.T) {
			helm, err := helm.New(
				fleet.Spec{
					RootDir: tt.rootDir,
				}, "", "")

			require.NoError(t, err)

			var pipelines []string
			rawPipelines, err := helm.DiscoverManifests()
			require.NoError(t, err)

			for i := range rawPipelines {
				pipelines = append(pipelines, string(rawPipelines[i]))
				assert.Equal(t, tt.expectedPipelines[i], pipelines[i])
			}
		})
	}
}
