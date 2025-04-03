package kubernetes

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
		flavor            string
		expectedPipelines []string
	}{
		{
			name:    "Scenario 1",
			rootDir: "testdata/success",
			flavor:  FlavorKubernetes,
			expectedPipelines: []string{`name: 'deps: bump container image "updatecli"'
sources:
  updatecli:
    name: 'get latest container image tag for "ghcr.io/updatecli/updatecli"'
    kind: 'dockerimage'
    spec:
      image: 'ghcr.io/updatecli/updatecli'
      tagfilter: '^v\d*(\.\d*){2}$'
      versionfilter:
        kind: 'semver'
        pattern: '>=v0.67.0'
targets:
  updatecli:
    name: 'deps: bump container image "ghcr.io/updatecli/updatecli" to {{ source "updatecli" }}'
    kind: 'yaml'
    spec:
      file: 'pod.yaml'
      key: "$.spec.containers[0].image"
    sourceid: 'updatecli'
    transformers:
      - addprefix: 'ghcr.io/updatecli/updatecli:'
`},
		},
		{
			name:    "Scenario 2 - Kustomize",
			rootDir: "testdata/kustomize",
			flavor:  FlavorKubernetes,
			expectedPipelines: []string{`name: 'deps: bump container image "nginx"'
sources:
  nginx:
    name: 'get latest container image tag for "nginx"'
    kind: 'dockerimage'
    spec:
      image: 'nginx'
      tagfilter: '^\d*(\.\d*){2}$'
      versionfilter:
        kind: 'semver'
        pattern: '>=1.20.0'
targets:
  nginx:
    name: 'deps: bump container image "nginx" to {{ source "nginx" }}'
    kind: 'yaml'
    spec:
      file: 'deployment.yaml'
      key: "$.spec.template.spec.containers[0].image"
    sourceid: 'nginx'
    transformers:
      - addprefix: 'nginx:'
`},
		},
		{
			name:    "Scenario - latest and digest",
			rootDir: "testdata/success",
			digest:  true,
			flavor:  FlavorKubernetes,
			expectedPipelines: []string{`name: 'deps: bump container image "updatecli"'
sources:
  updatecli:
    name: 'get latest container image tag for "ghcr.io/updatecli/updatecli"'
    kind: 'dockerimage'
    spec:
      image: 'ghcr.io/updatecli/updatecli'
      tagfilter: '^v\d*(\.\d*){2}$'
      versionfilter:
        kind: 'semver'
        pattern: '>=v0.67.0'
  updatecli-digest:
    name: 'get latest container image digest for "ghcr.io/updatecli/updatecli:v0.67.0"'
    kind: 'dockerdigest'
    spec:
      image: 'ghcr.io/updatecli/updatecli'
      tag: '{{ source "updatecli" }}'
    dependson:
      - 'updatecli'
targets:
  updatecli:
    name: 'deps: bump container image digest for "ghcr.io/updatecli/updatecli:{{ source "updatecli" }}"'
    kind: 'yaml'
    spec:
      file: 'pod.yaml'
      key: "$.spec.containers[0].image"
    sourceid: 'updatecli-digest'
    transformers:
      - addprefix: 'ghcr.io/updatecli/updatecli:'
`},
		},
		{
			name:    "Scenario - prow",
			rootDir: "testdata/prow",
			digest:  true,
			flavor:  FlavorProw,
			expectedPipelines: []string{`name: 'deps: bump container image "ghcr.io/updatecli/updatecli" for repo "*" and presubmit test "pull-updatecli-diff"'
sources:
  ghcr.io/updatecli/updatecli:
    name: 'get latest container image tag for "ghcr.io/updatecli/updatecli"'
    kind: 'dockerimage'
    spec:
      image: 'ghcr.io/updatecli/updatecli'
      tagfilter: '^v\d*(\.\d*){2}$'
      versionfilter:
        kind: 'semver'
        pattern: '>=v0.82.2'
  ghcr.io/updatecli/updatecli-digest:
    name: 'get latest container image digest for "ghcr.io/updatecli/updatecli:v0.82.2"'
    kind: 'dockerdigest'
    spec:
      image: 'ghcr.io/updatecli/updatecli'
      tag: '{{ source "ghcr.io/updatecli/updatecli" }}'
    dependson:
      - 'ghcr.io/updatecli/updatecli'
targets:
  ghcr.io/updatecli/updatecli:
    name: 'deps: bump container image digest for "ghcr.io/updatecli/updatecli:{{ source "ghcr.io/updatecli/updatecli" }}"'
    kind: 'yaml'
    spec:
      file: 'prow.yaml'
      key: "$.presubmits.'*'[0].spec.containers[0].image"
    sourceid: 'ghcr.io/updatecli/updatecli-digest'
    transformers:
      - addprefix: 'ghcr.io/updatecli/updatecli:'
`, `name: 'deps: bump container image "updatecli" for repo "updatecli/updatecli" and postsubmit test "pull-updatecli-apply"'
sources:
  updatecli:
    name: 'get latest container image tag for "ghcr.io/updatecli/updatecli"'
    kind: 'dockerimage'
    spec:
      image: 'ghcr.io/updatecli/updatecli'
      tagfilter: '^v\d*(\.\d*){2}$'
      versionfilter:
        kind: 'semver'
        pattern: '>=v0.82.2'
  updatecli-digest:
    name: 'get latest container image digest for "ghcr.io/updatecli/updatecli:v0.82.2"'
    kind: 'dockerdigest'
    spec:
      image: 'ghcr.io/updatecli/updatecli'
      tag: '{{ source "updatecli" }}'
    dependson:
      - 'updatecli'
targets:
  updatecli:
    name: 'deps: bump container image digest for "ghcr.io/updatecli/updatecli:{{ source "updatecli" }}"'
    kind: 'yaml'
    spec:
      file: 'prow.yaml'
      key: "$.postsubmits.'updatecli/updatecli'[0].spec.containers[0].image"
    sourceid: 'updatecli-digest'
    transformers:
      - addprefix: 'ghcr.io/updatecli/updatecli:'
`, `name: 'deps: bump container image "updatecli" for periodic test "pull-updatecli-apply-cron"'
sources:
  updatecli:
    name: 'get latest container image tag for "ghcr.io/updatecli/updatecli"'
    kind: 'dockerimage'
    spec:
      image: 'ghcr.io/updatecli/updatecli'
      tagfilter: '^v\d*(\.\d*){2}$'
      versionfilter:
        kind: 'semver'
        pattern: '>=v0.82.2'
  updatecli-digest:
    name: 'get latest container image digest for "ghcr.io/updatecli/updatecli:v0.82.2"'
    kind: 'dockerdigest'
    spec:
      image: 'ghcr.io/updatecli/updatecli'
      tag: '{{ source "updatecli" }}'
    dependson:
      - 'updatecli'
targets:
  updatecli:
    name: 'deps: bump container image digest for "ghcr.io/updatecli/updatecli:{{ source "updatecli" }}"'
    kind: 'yaml'
    spec:
      file: 'prow.yaml'
      key: "$.periodics[0].spec.containers[0].image"
    sourceid: 'updatecli-digest'
    transformers:
      - addprefix: 'ghcr.io/updatecli/updatecli:'
`},
		},
		{
			name:    "cronjob",
			rootDir: "testdata/cronjob",
			digest:  false,
			flavor:  FlavorKubernetes,
			expectedPipelines: []string{`name: 'deps: bump container image "updatecli"'
sources:
  updatecli:
    name: 'get latest container image tag for "ghcr.io/updatecli/updatecli"'
    kind: 'dockerimage'
    spec:
      image: 'ghcr.io/updatecli/updatecli'
      tagfilter: '^v\d*(\.\d*){2}$'
      versionfilter:
        kind: 'semver'
        pattern: '>=v0.67.0'
targets:
  updatecli:
    name: 'deps: bump container image "ghcr.io/updatecli/updatecli" to {{ source "updatecli" }}'
    kind: 'yaml'
    spec:
      file: 'cronjob.yaml'
      key: "$.spec.jobTemplate.spec.template.spec.containers[0].image"
    sourceid: 'updatecli'
    transformers:
      - addprefix: 'ghcr.io/updatecli/updatecli:'
`},
		},
		{
			name:    "initContainers",
			rootDir: "testdata/initContainers",
			digest:  false,
			flavor:  FlavorKubernetes,
			expectedPipelines: []string{`name: 'deps: bump container image "updatecli"'
sources:
  updatecli:
    name: 'get latest container image tag for "ghcr.io/updatecli/updatecli"'
    kind: 'dockerimage'
    spec:
      image: 'ghcr.io/updatecli/updatecli'
      tagfilter: '^v\d*(\.\d*){2}$'
      versionfilter:
        kind: 'semver'
        pattern: '>=v0.67.0'
targets:
  updatecli:
    name: 'deps: bump container image "ghcr.io/updatecli/updatecli" to {{ source "updatecli" }}'
    kind: 'yaml'
    spec:
      file: 'initContainers.yaml'
      key: "$.spec.template.spec.initContainers[0].image"
    sourceid: 'updatecli'
    transformers:
      - addprefix: 'ghcr.io/updatecli/updatecli:'
`},
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			digest := tt.digest
			pod, err := New(
				Spec{
					Digest: &digest,
				}, tt.rootDir, "", "", tt.flavor)

			require.NoError(t, err)

			var pipelines []string
			rawPipelines, err := pod.DiscoverManifests()
			require.NoError(t, err)

			if len(rawPipelines) == 0 {
				t.Errorf("No pipelines found for %s", tt.name)
			}

			assert.Equal(t, len(tt.expectedPipelines), len(rawPipelines))

			for i := range rawPipelines {
				// We expect manifest generated by the autodiscovery to use the yaml syntax
				pipelines = append(pipelines, string(rawPipelines[i]))
				assert.Equal(t, tt.expectedPipelines[i], pipelines[i])
			}
		})
	}
}
