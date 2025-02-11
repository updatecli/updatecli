package helm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscoverManifests(t *testing.T) {
	// Disable condition testing with running short test
	if testing.Short() {
		return
	}

	testdata := []struct {
		name              string
		rootDir           string
		digest            bool
		expectedPipelines []string
	}{
		{
			name:    "Scenario 1",
			rootDir: "testdata-1/chart",
			expectedPipelines: []string{`name: 'Bump dependency "minio" for Helm chart "epinio"'
sources:
  helmchart:
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
      skippackaging: false
      versionincrement: ''
    sourceid: 'helmchart'
`,
				`name: 'Bump dependency "kubed" for Helm chart "epinio"'
sources:
  helmchart:
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
      skippackaging: false
      versionincrement: ''
    sourceid: 'helmchart'
`,
				`name: 'Bump dependency "epinio-ui" for Helm chart "epinio"'
sources:
  helmchart:
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
      skippackaging: false
      versionincrement: ''
    sourceid: 'helmchart'
`,
				`name: 'deps(helm): bump image "splatform/epinio-server" tag for chart "epinio"'
sources:
  image:
    name: 'get latest image tag for "splatform/epinio-server"'
    kind: 'dockerimage'
    spec:
      image: 'splatform/epinio-server'
      tagfilter: '^\d*(\.\d*){2}$'
      versionfilter:
        kind: 'semver'
        pattern: '>=2.0.0'
conditions:
  splatform_epinio-server-repository:
    disablesourceinput: true
    name: 'Ensure container repository "splatform/epinio-server" is specified'
    kind: 'yaml'
    spec:
      file: 'epinio/values.yaml'
      key: '$.image.repository'
      value: 'splatform/epinio-server'
targets:
  splatform_epinio-server:
    name: 'deps(helm): bump image "splatform/epinio-server" tag'
    kind: 'helmchart'
    spec:
      file: 'values.yaml'
      name: 'epinio'
      key: '$.image.tag'
      skippackaging: false
      versionincrement: ''
    sourceid: 'image'
`,
				`name: 'deps(helm): bump image "epinioteam/epinio-ui-qa" tag for chart "epinio"'
sources:
  image:
    name: 'get latest image tag for "epinioteam/epinio-ui-qa"'
    kind: 'dockerimage'
    spec:
      image: 'epinioteam/epinio-ui-qa'
      tagfilter: '^\d*(\.\d*){2}$'
      versionfilter:
        kind: 'semver'
        pattern: '>=1.0.0'
conditions:
  epinioteam_epinio-ui-qa-repository:
    disablesourceinput: true
    name: 'Ensure container repository "epinioteam/epinio-ui-qa" is specified'
    kind: 'yaml'
    spec:
      file: 'epinio/values.yaml'
      key: '$.images.ui.repository'
      value: 'epinioteam/epinio-ui-qa'
targets:
  epinioteam_epinio-ui-qa:
    name: 'deps(helm): bump image "epinioteam/epinio-ui-qa" tag'
    kind: 'helmchart'
    spec:
      file: 'values.yaml'
      name: 'epinio'
      key: '$.images.ui.tag'
      skippackaging: false
      versionincrement: ''
    sourceid: 'image'
`},
		},
		{
			name:    "Test the tag update for images referenced in the chart with digest",
			rootDir: "testdata-2/chart",
			digest:  true,
			expectedPipelines: []string{`name: 'deps(helm): bump image "epinio/epinio-server" digest for chart "sample"'
sources:
  image:
    name: 'get latest "epinio/epinio-server" container tag'
    kind: 'dockerimage'
    spec:
      image: 'epinio/epinio-server'
      tagfilter: '^v\d*(\.\d*){2}$'
      versionfilter:
        kind: 'semver'
        pattern: '>=v1.8.0'
  image-digest:
    name: 'get latest image "epinio/epinio-server" digest'
    kind: 'dockerdigest'
    spec:
      image: 'epinio/epinio-server'
      tag: '{{ source "image" }}'
    dependson:
      - 'image'
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
    name: 'deps(helm): bump image "epinio/epinio-server" digest'
    kind: 'helmchart'
    spec:
      file: 'values.yaml'
      name: 'sample'
      key: '$.image.tag'
      skippackaging: false
      versionincrement: ''
    sourceid: 'image-digest'
`,
				`name: 'deps(helm): bump image "epinio/epinio-ui" digest for chart "sample"'
sources:
  image:
    name: 'get latest "epinio/epinio-ui" container tag'
    kind: 'dockerimage'
    spec:
      image: 'epinio/epinio-ui'
      tagfilter: '^v\d*(\.\d*){2}$'
      versionfilter:
        kind: 'semver'
        pattern: '>=v1.8.0'
  image-digest:
    name: 'get latest image "epinio/epinio-ui" digest'
    kind: 'dockerdigest'
    spec:
      image: 'epinio/epinio-ui'
      tag: '{{ source "image" }}'
    dependson:
      - 'image'
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
    name: 'deps(helm): bump image "epinio/epinio-ui" digest'
    kind: 'helmchart'
    spec:
      file: 'values.yaml'
      name: 'sample'
      key: '$.images.ui.tag'
      skippackaging: false
      versionincrement: ''
    sourceid: 'image-digest'
`},
		},
		{
			name:    "Test the tag update for images referenced in the cart including the registry",
			rootDir: "testdata-3/chart",
			expectedPipelines: []string{`name: 'deps(helm): bump image "ghcr.io/epinio/epinio-server" tag for chart "sample"'
sources:
  image:
    name: 'get latest image tag for "ghcr.io/epinio/epinio-server"'
    kind: 'dockerimage'
    spec:
      image: 'ghcr.io/epinio/epinio-server'
      tagfilter: '^v\d*(\.\d*){2}$'
      versionfilter:
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
    name: 'deps(helm): bump image "ghcr.io/epinio/epinio-server" tag'
    kind: 'helmchart'
    spec:
      file: 'values.yaml'
      name: 'sample'
      key: '$.image.tag'
      skippackaging: false
      versionincrement: ''
    sourceid: 'image'
`,
				`name: 'deps(helm): bump image "ghcr.io/epinio/epinio-ui" tag for chart "sample"'
sources:
  image:
    name: 'get latest image tag for "ghcr.io/epinio/epinio-ui"'
    kind: 'dockerimage'
    spec:
      image: 'ghcr.io/epinio/epinio-ui'
      tagfilter: '^v\d*(\.\d*){2}$'
      versionfilter:
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
    name: 'deps(helm): bump image "ghcr.io/epinio/epinio-ui" tag'
    kind: 'helmchart'
    spec:
      file: 'values.yaml'
      name: 'sample'
      key: '$.images.ui.tag'
      skippackaging: false
      versionincrement: ''
    sourceid: 'image'
`},
		},
		{
			name:              "Scenario 4 - missing required information",
			rootDir:           "testdata-4/chart",
			expectedPipelines: []string{},
		},
		{
			name:              "Scenario 5 - nil sourceSpec",
			rootDir:           "testdata-5/chart",
			expectedPipelines: []string{},
		},
		{
			name:    "Test latest tag when empty string",
			rootDir: "testdata-6/chart",
			digest:  true,
			expectedPipelines: []string{`name: 'deps(helm): bump image "nginx" digest for chart "test-tag-01"'
sources:
  image-digest:
    name: 'get latest image "nginx" digest'
    kind: 'dockerdigest'
    spec:
      image: 'nginx'
      tag: 'latest'
conditions:
  nginx-repository:
    disablesourceinput: true
    name: 'Ensure container repository "nginx" is specified'
    kind: 'yaml'
    spec:
      file: 'test-tag-01/values.yaml'
      key: '$.image.repository'
      value: 'nginx'
targets:
  nginx:
    name: 'deps(helm): bump image "nginx" digest'
    kind: 'helmchart'
    spec:
      file: 'values.yaml'
      name: 'test-tag-01'
      key: '$.image.tag'
      skippackaging: false
      versionincrement: ''
    sourceid: 'image-digest'
`},
		},
		{
			name:    "Test latest tag",
			rootDir: "testdata-6/chart",
			digest:  true,
			expectedPipelines: []string{`name: 'deps(helm): bump image "nginx" digest for chart "test-tag-01"'
sources:
  image-digest:
    name: 'get latest image "nginx" digest'
    kind: 'dockerdigest'
    spec:
      image: 'nginx'
      tag: 'latest'
conditions:
  nginx-repository:
    disablesourceinput: true
    name: 'Ensure container repository "nginx" is specified'
    kind: 'yaml'
    spec:
      file: 'test-tag-01/values.yaml'
      key: '$.image.repository'
      value: 'nginx'
targets:
  nginx:
    name: 'deps(helm): bump image "nginx" digest'
    kind: 'helmchart'
    spec:
      file: 'values.yaml'
      name: 'test-tag-01'
      key: '$.image.tag'
      skippackaging: false
      versionincrement: ''
    sourceid: 'image-digest'
`},
		},
		{
			name:    "Test latest tag when tag not defined",
			rootDir: "testdata-6/chart",
			digest:  true,
			expectedPipelines: []string{`name: 'deps(helm): bump image "nginx" digest for chart "test-tag-01"'
sources:
  image-digest:
    name: 'get latest image "nginx" digest'
    kind: 'dockerdigest'
    spec:
      image: 'nginx'
      tag: 'latest'
conditions:
  nginx-repository:
    disablesourceinput: true
    name: 'Ensure container repository "nginx" is specified'
    kind: 'yaml'
    spec:
      file: 'test-tag-01/values.yaml'
      key: '$.image.repository'
      value: 'nginx'
targets:
  nginx:
    name: 'deps(helm): bump image "nginx" digest'
    kind: 'helmchart'
    spec:
      file: 'values.yaml'
      name: 'test-tag-01'
      key: '$.image.tag'
      skippackaging: false
      versionincrement: ''
    sourceid: 'image-digest'
`},
		},
	}

	for _, tt := range testdata {

		digest := tt.digest
		t.Run(tt.name, func(t *testing.T) {
			helm, err := New(
				Spec{
					Digest: &digest,
				}, tt.rootDir, "", "")

			require.NoError(t, err)

			var pipelines []string
			bytesPipelines, err := helm.DiscoverManifests()
			require.NoError(t, err)
			assert.Equal(t, len(tt.expectedPipelines), len(bytesPipelines))

			stringPipelines := []string{}
			for i := range bytesPipelines {
				stringPipelines = append(stringPipelines, string(bytesPipelines[i]))
			}
			//sort.Strings(stringPipelines)

			for i := range stringPipelines {
				pipelines = append(pipelines, stringPipelines...)
				assert.Equal(t, tt.expectedPipelines[i], pipelines[i])
			}
		})
	}
}
