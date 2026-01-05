package flux

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker"
)

func TestDiscoverManifests(t *testing.T) {
	testdata := []struct {
		name              string
		rootDir           string
		digest            bool
		auths             map[string]docker.InlineKeyChain
		scmID             string
		actionID          string
		expectedPipelines []string
	}{
		{
			name:    "Scenario - helmrelease Simple with auth",
			rootDir: "testdata/helmrelease/simple",
			auths: map[string]docker.InlineKeyChain{
				"updatecli.github.io": {
					Token: "mytoken",
				},
			},
			expectedPipelines: []string{`name: 'deps(flux): bump Helmrelease "udash"'
sources:
  helmrelease:
    name: 'Get latest "udash" Helm chart version'
    kind: 'helmchart'
    spec:
      name: 'udash'
      url: 'https://updatecli.github.io/charts'
      token: 'mytoken'
      versionfilter:
        kind: 'semver'
        pattern: '*'
conditions:
  helmrelease:
    name: 'Ensure Helm Chart name "udash"'
    kind: 'yaml'
    disablesourceinput: true
    spec:
      file: 'helmrelease.yaml'
      key: '$.spec.chart.spec.chart'
      value: 'udash'
targets:
  helmrelease:
    name: 'deps(flux): bump Helmrelease "udash"'
    kind: 'yaml'
    spec:
      file: 'helmrelease.yaml'
      key: '$.spec.chart.spec.version'
    sourceid: 'helmrelease'
`},
		},
		{
			name:    "Scenario - helmrelease Simple",
			rootDir: "testdata/helmrelease/simple",
			expectedPipelines: []string{`name: 'deps(flux): bump Helmrelease "udash"'
sources:
  helmrelease:
    name: 'Get latest "udash" Helm chart version'
    kind: 'helmchart'
    spec:
      name: 'udash'
      url: 'https://updatecli.github.io/charts'
      versionfilter:
        kind: 'semver'
        pattern: '*'
conditions:
  helmrelease:
    name: 'Ensure Helm Chart name "udash"'
    kind: 'yaml'
    disablesourceinput: true
    spec:
      file: 'helmrelease.yaml'
      key: '$.spec.chart.spec.chart'
      value: 'udash'
targets:
  helmrelease:
    name: 'deps(flux): bump Helmrelease "udash"'
    kind: 'yaml'
    spec:
      file: 'helmrelease.yaml'
      key: '$.spec.chart.spec.version'
    sourceid: 'helmrelease'
`},
		},
		{
			name:     "Scenario - helmrelease Simple with scmid and actionid",
			rootDir:  "testdata/helmrelease/simple",
			scmID:    "defaultscmid",
			actionID: "defaultactionid",
			expectedPipelines: []string{`name: 'deps(flux): bump Helmrelease "udash"'
actions:
  defaultactionid:
    title: 'deps: update Helm chart to {{ source "helmrelease" }}'
sources:
  helmrelease:
    name: 'Get latest "udash" Helm chart version'
    kind: 'helmchart'
    spec:
      name: 'udash'
      url: 'https://updatecli.github.io/charts'
      versionfilter:
        kind: 'semver'
        pattern: '*'
conditions:
  helmrelease:
    name: 'Ensure Helm Chart name "udash"'
    kind: 'yaml'
    disablesourceinput: true
    scmid: defaultscmid
    spec:
      file: 'helmrelease.yaml'
      key: '$.spec.chart.spec.chart'
      value: 'udash'
targets:
  helmrelease:
    name: 'deps(flux): bump Helmrelease "udash"'
    kind: 'yaml'
    scmid: defaultscmid
    spec:
      file: 'helmrelease.yaml'
      key: '$.spec.chart.spec.version'
    sourceid: 'helmrelease'
`},
		},
		{
			name:    "Scenario - helmrelease Simple with both release and repository in same file",
			rootDir: "testdata/helmrelease/simple-combined",
			expectedPipelines: []string{`name: 'deps(flux): bump Helmrelease "udash"'
sources:
  helmrelease:
    name: 'Get latest "udash" Helm chart version'
    kind: 'helmchart'
    spec:
      name: 'udash'
      url: 'https://updatecli.github.io/charts'
      versionfilter:
        kind: 'semver'
        pattern: '*'
conditions:
  helmrelease:
    name: 'Ensure Helm Chart name "udash"'
    kind: 'yaml'
    disablesourceinput: true
    spec:
      file: 'helmrelease-helmrepository.yaml'
      key: '$.spec.chart.spec.chart'
      value: 'udash'
targets:
  helmrelease:
    name: 'deps(flux): bump Helmrelease "udash"'
    kind: 'yaml'
    spec:
      file: 'helmrelease-helmrepository.yaml'
      key: '$.spec.chart.spec.version'
    sourceid: 'helmrelease'
`},
		},
		{
			name:    "Scenario - helmrelease OCI",
			rootDir: "testdata/helmrelease/oci",
			expectedPipelines: []string{`name: 'deps(flux): bump Helmrelease "upgrade-responder"'
sources:
  helmrelease:
    name: 'Get latest "upgrade-responder" Helm chart version'
    kind: 'helmchart'
    spec:
      name: 'upgrade-responder'
      url: 'oci://ghcr.io/olblak/charts/'
      versionfilter:
        kind: 'semver'
        pattern: '*'
conditions:
  helmrelease:
    name: 'Ensure Helm Chart name "upgrade-responder"'
    kind: 'yaml'
    disablesourceinput: true
    spec:
      file: 'helmrelease.yaml'
      key: '$.spec.chart.spec.chart'
      value: 'upgrade-responder'
targets:
  helmrelease:
    name: 'deps(flux): bump Helmrelease "upgrade-responder"'
    kind: 'yaml'
    spec:
      file: 'helmrelease.yaml'
      key: '$.spec.chart.spec.version'
    sourceid: 'helmrelease'
`},
		},
		{
			name:    "Scenario - helmrelease OCI",
			rootDir: "testdata/helmrelease/oci-combined",
			expectedPipelines: []string{`name: 'deps(flux): bump Helmrelease "upgrade-responder"'
sources:
  helmrelease:
    name: 'Get latest "upgrade-responder" Helm chart version'
    kind: 'helmchart'
    spec:
      name: 'upgrade-responder'
      url: 'oci://ghcr.io/olblak/charts/'
      versionfilter:
        kind: 'semver'
        pattern: '*'
conditions:
  helmrelease:
    name: 'Ensure Helm Chart name "upgrade-responder"'
    kind: 'yaml'
    disablesourceinput: true
    spec:
      file: 'helmrelease-helmrepository.yaml'
      key: '$.spec.chart.spec.chart'
      value: 'upgrade-responder'
targets:
  helmrelease:
    name: 'deps(flux): bump Helmrelease "upgrade-responder"'
    kind: 'yaml'
    spec:
      file: 'helmrelease-helmrepository.yaml'
      key: '$.spec.chart.spec.version'
    sourceid: 'helmrelease'
`},
		},
		{
			name:    "Scenario - ocirepository ",
			digest:  false,
			rootDir: "testdata/ociRepository",
			expectedPipelines: []string{`name: 'deps(flux): bump ociRepository "ghcr.io/updatecli/updatecli"'
sources:
  oci:
    name: 'Get latest "ghcr.io/updatecli/updatecli" OCI artifact tag'
    kind: 'dockerimage'
    spec:
      image: 'ghcr.io/updatecli/updatecli'
      tagfilter: '^v\d*(\.\d*){2}$'
      versionfilter:
        kind: 'semver'
        pattern: '>=v0.22.0'
targets:
  oci:
    name: 'deps(flux): bump OCI repository "ghcr.io/updatecli/updatecli"'
    kind: 'yaml'
    spec:
      file: 'example.yaml'
      key: '$.spec.ref.tag'
    sourceid: 'oci'
`},
		},
		{
			name:    "Scenario - ocirepository version with digest",
			digest:  true,
			rootDir: "testdata/ociRepository",
			expectedPipelines: []string{`name: 'deps(flux): bump ociRepository "ghcr.io/updatecli/updatecli"'
sources:
  oci:
    name: 'Get latest "ghcr.io/updatecli/updatecli" OCI artifact tag'
    kind: 'dockerimage'
    spec:
      image: 'ghcr.io/updatecli/updatecli'
      tagfilter: '^v\d*(\.\d*){2}$'
      versionfilter:
        kind: 'semver'
        pattern: '>=v0.22.0'
  oci-digest:
    name: 'Get latest "ghcr.io/updatecli/updatecli" OCI artifact digest'
    kind: 'dockerdigest'
    spec:
      image: 'ghcr.io/updatecli/updatecli'
      tag: '{{ source "oci" }}'
    dependson:
      - 'oci'
targets:
  oci:
    name: 'deps(flux): bump OCI repository "ghcr.io/updatecli/updatecli"'
    kind: 'yaml'
    spec:
      file: 'example.yaml'
      key: '$.spec.ref.tag'
    sourceid: 'oci-digest'
`},
		},
		{
			name:     "Scenario - ocirepository latest with digest",
			digest:   true,
			rootDir:  "testdata/ociRepository-latest",
			scmID:    "defaultscmid",
			actionID: "defaultactionid",
			expectedPipelines: []string{`name: 'deps(flux): bump ociRepository "ghcr.io/updatecli/updatecli"'
actions:
  defaultactionid:
    title: 'deps: update OCI repository digest for ghcr.io/updatecli/updatecli:latest'
sources:
  oci-digest:
    name: 'Get latest "ghcr.io/updatecli/updatecli" OCI artifact digest'
    kind: 'dockerdigest'
    spec:
      image: 'ghcr.io/updatecli/updatecli'
      tag: 'latest'
targets:
  oci:
    name: 'deps(flux): bump OCI repository "ghcr.io/updatecli/updatecli"'
    kind: 'yaml'
    scmid: defaultscmid
    spec:
      file: 'example.yaml'
      key: '$.spec.ref.tag'
    sourceid: 'oci-digest'
`},
		},
	}

	for _, tt := range testdata {

		t.Run(tt.name, func(t *testing.T) {
			digest := tt.digest
			flux, err := New(
				Spec{
					Digest: &digest,
					Auths:  tt.auths,
				}, tt.rootDir, tt.scmID, tt.actionID)

			require.NoError(t, err)

			var pipelines []string
			rawPipelines, err := flux.DiscoverManifests()
			require.NoError(t, err)

			assert.Equal(t, len(tt.expectedPipelines), len(rawPipelines))

			for i := range rawPipelines {
				// We expect manifest generated by the autodiscovery to use the yaml syntax
				pipelines = append(pipelines, string(rawPipelines[i]))
				assert.Equal(t, tt.expectedPipelines[i], pipelines[i])
			}
		})
	}

}
