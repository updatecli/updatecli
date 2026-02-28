package argocd

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscoverManifests(t *testing.T) {
	testdata := []struct {
		name              string
		rootDir           string
		actionID          string
		scmID             string
		expectedPipelines []string
		spec              Spec
	}{
		{
			name:    "ArgoCD manifests discovered with multiple documents",
			rootDir: "testdata/multi-release",
			expectedPipelines: []string{
				`name: 'deps(helm): bump Helm chart "nginx" in ArgoCD manifest "manifest.yaml"'
sources:
  nginx:
    name: 'Get latest "nginx" Helm chart version'
    kind: 'helmchart'
    spec:
      name: 'nginx'
      url: 'oci://registry-1.docker.io/bitnamicharts'
      versionfilter:
        kind: 'semver'
        pattern: '*'
conditions:
  nginx-name:
    name: 'Ensure Helm chart name nginx is specified'
    kind: 'yaml'
    disablesourceinput: true
    spec:
      file: 'manifest.yaml'
      key: '$.spec.source.chart'
      documentindex: 1
      value: 'nginx'
  nginx-repository:
    name: 'Ensure Helm chart repository oci://registry-1.docker.io/bitnamicharts is specified'
    kind: 'yaml'
    disablesourceinput: true
    spec:
      file: 'manifest.yaml'
      key: '$.spec.source.repoURL'
      documentindex: 1
      value: 'oci://registry-1.docker.io/bitnamicharts'
targets:
  nginx:
    name: 'deps(helm): update Helm chart "nginx" to {{ source "nginx" }}'
    kind: 'yaml'
    spec:
      file: 'manifest.yaml'
      key: '$.spec.source.targetRevision'
      documentindex: 1
    sourceid: 'nginx'
`,
				`name: 'deps(helm): bump Helm chart "nginx" in ArgoCD manifest "manifest.yaml"'
sources:
  nginx:
    name: 'Get latest "nginx" Helm chart version'
    kind: 'helmchart'
    spec:
      name: 'nginx'
      url: 'oci://registry-1.docker.io/bitnamicharts'
      versionfilter:
        kind: 'semver'
        pattern: '*'
conditions:
  nginx-name:
    name: 'Ensure Helm chart name nginx is specified'
    kind: 'yaml'
    disablesourceinput: true
    spec:
      file: 'manifest.yaml'
      key: '$.spec.template.spec.source.chart'
      documentindex: 2
      value: 'nginx'
  nginx-repository:
    name: 'Ensure Helm chart repository oci://registry-1.docker.io/bitnamicharts is specified'
    kind: 'yaml'
    disablesourceinput: true
    spec:
      file: 'manifest.yaml'
      key: '$.spec.template.spec.source.repoURL'
      documentindex: 2
      value: 'oci://registry-1.docker.io/bitnamicharts'
targets:
  nginx:
    name: 'deps(helm): update Helm chart "nginx" to {{ source "nginx" }}'
    kind: 'yaml'
    spec:
      file: 'manifest.yaml'
      key: '$.spec.template.spec.source.targetRevision'
      documentindex: 2
    sourceid: 'nginx'
`,
			},
		},
		{
			name:              "ArgoCD manifests discovered no source",
			rootDir:           "testdata/empty",
			expectedPipelines: []string{},
		},
		{
			name:    "ArgoCD manifests discovery with a single source and auths",
			rootDir: "testdata/sealed-secrets",
			spec: Spec{
				Auths: map[string]auth{
					"bitnami-labs.github.io": {
						Token: "token",
					},
				},
			},
			expectedPipelines: []string{`name: 'deps(helm): bump Helm chart "sealed-secrets" in ArgoCD manifest "manifest.yaml"'
sources:
  sealed-secrets:
    name: 'Get latest "sealed-secrets" Helm chart version'
    kind: 'helmchart'
    spec:
      name: 'sealed-secrets'
      url: 'https://bitnami-labs.github.io/sealed-secrets'
      token: 'token'
      versionfilter:
        kind: 'semver'
        pattern: '*'
conditions:
  sealed-secrets-name:
    name: 'Ensure Helm chart name sealed-secrets is specified'
    kind: 'yaml'
    disablesourceinput: true
    spec:
      file: 'manifest.yaml'
      key: '$.spec.source.chart'
      documentindex: 0
      value: 'sealed-secrets'
  sealed-secrets-repository:
    name: 'Ensure Helm chart repository https://bitnami-labs.github.io/sealed-secrets is specified'
    kind: 'yaml'
    disablesourceinput: true
    spec:
      file: 'manifest.yaml'
      key: '$.spec.source.repoURL'
      documentindex: 0
      value: 'https://bitnami-labs.github.io/sealed-secrets'
targets:
  sealed-secrets:
    name: 'deps(helm): update Helm chart "sealed-secrets" to {{ source "sealed-secrets" }}'
    kind: 'yaml'
    spec:
      file: 'manifest.yaml'
      key: '$.spec.source.targetRevision'
      documentindex: 0
    sourceid: 'sealed-secrets'
`},
		},
		{
			name:    "ArgoCD manifests discovery with a single source",
			rootDir: "testdata/sealed-secrets",
			expectedPipelines: []string{`name: 'deps(helm): bump Helm chart "sealed-secrets" in ArgoCD manifest "manifest.yaml"'
sources:
  sealed-secrets:
    name: 'Get latest "sealed-secrets" Helm chart version'
    kind: 'helmchart'
    spec:
      name: 'sealed-secrets'
      url: 'https://bitnami-labs.github.io/sealed-secrets'
      versionfilter:
        kind: 'semver'
        pattern: '*'
conditions:
  sealed-secrets-name:
    name: 'Ensure Helm chart name sealed-secrets is specified'
    kind: 'yaml'
    disablesourceinput: true
    spec:
      file: 'manifest.yaml'
      key: '$.spec.source.chart'
      documentindex: 0
      value: 'sealed-secrets'
  sealed-secrets-repository:
    name: 'Ensure Helm chart repository https://bitnami-labs.github.io/sealed-secrets is specified'
    kind: 'yaml'
    disablesourceinput: true
    spec:
      file: 'manifest.yaml'
      key: '$.spec.source.repoURL'
      documentindex: 0
      value: 'https://bitnami-labs.github.io/sealed-secrets'
targets:
  sealed-secrets:
    name: 'deps(helm): update Helm chart "sealed-secrets" to {{ source "sealed-secrets" }}'
    kind: 'yaml'
    spec:
      file: 'manifest.yaml'
      key: '$.spec.source.targetRevision'
      documentindex: 0
    sourceid: 'sealed-secrets'
`},
		},
		{
			name:    "ArgoCD manifests discovery with several sources",
			rootDir: "testdata/sealed-secrets_sources",
			expectedPipelines: []string{`name: 'deps(helm): bump Helm chart "sealed-secrets" in ArgoCD manifest "manifest.yaml"'
sources:
  sealed-secrets:
    name: 'Get latest "sealed-secrets" Helm chart version'
    kind: 'helmchart'
    spec:
      name: 'sealed-secrets'
      url: 'https://bitnami-labs.github.io/sealed-secrets'
      versionfilter:
        kind: 'semver'
        pattern: '*'
conditions:
  sealed-secrets-name:
    name: 'Ensure Helm chart name sealed-secrets is specified'
    kind: 'yaml'
    disablesourceinput: true
    spec:
      file: 'manifest.yaml'
      key: '$.spec.sources[0].chart'
      documentindex: 0
      value: 'sealed-secrets'
  sealed-secrets-repository:
    name: 'Ensure Helm chart repository https://bitnami-labs.github.io/sealed-secrets is specified'
    kind: 'yaml'
    disablesourceinput: true
    spec:
      file: 'manifest.yaml'
      key: '$.spec.sources[0].repoURL'
      documentindex: 0
      value: 'https://bitnami-labs.github.io/sealed-secrets'
targets:
  sealed-secrets:
    name: 'deps(helm): update Helm chart "sealed-secrets" to {{ source "sealed-secrets" }}'
    kind: 'yaml'
    spec:
      file: 'manifest.yaml'
      key: '$.spec.sources[0].targetRevision'
      documentindex: 0
    sourceid: 'sealed-secrets'
`},
		},
		{
			name:     "ArgoCD manifests discovery with several sources and both action and scm IDs",
			rootDir:  "testdata/sealed-secrets_sources",
			actionID: "argoRepo",
			scmID:    "scm123",
			expectedPipelines: []string{`name: 'deps(helm): bump Helm chart "sealed-secrets" in ArgoCD manifest "manifest.yaml"'
actions:
  argoRepo:
    title: 'deps(argocd): update Helm chart sealed-secrets to {{ source "sealed-secrets" }}'
sources:
  sealed-secrets:
    name: 'Get latest "sealed-secrets" Helm chart version'
    kind: 'helmchart'
    spec:
      name: 'sealed-secrets'
      url: 'https://bitnami-labs.github.io/sealed-secrets'
      versionfilter:
        kind: 'semver'
        pattern: '*'
conditions:
  sealed-secrets-name:
    name: 'Ensure Helm chart name sealed-secrets is specified'
    kind: 'yaml'
    disablesourceinput: true
    scmid: scm123
    spec:
      file: 'manifest.yaml'
      key: '$.spec.sources[0].chart'
      documentindex: 0
      value: 'sealed-secrets'
  sealed-secrets-repository:
    name: 'Ensure Helm chart repository https://bitnami-labs.github.io/sealed-secrets is specified'
    kind: 'yaml'
    disablesourceinput: true
    scmid: scm123
    spec:
      file: 'manifest.yaml'
      key: '$.spec.sources[0].repoURL'
      documentindex: 0
      value: 'https://bitnami-labs.github.io/sealed-secrets'
targets:
  sealed-secrets:
    name: 'deps(helm): update Helm chart "sealed-secrets" to {{ source "sealed-secrets" }}'
    kind: 'yaml'
    scmid: scm123
    spec:
      file: 'manifest.yaml'
      key: '$.spec.sources[0].targetRevision'
      documentindex: 0
    sourceid: 'sealed-secrets'
`},
		},
		{
			name:    "ArgoCD manifests discovery with OCI source",
			rootDir: "testdata/oci-helm-source",
			expectedPipelines: []string{`name: 'deps(helm): bump Helm chart "nginx" in ArgoCD manifest "manifest.yaml"'
sources:
  nginx:
    name: 'Get latest "nginx" Helm chart version'
    kind: 'helmchart'
    spec:
      name: 'nginx'
      url: 'oci://registry-1.docker.io/bitnamicharts'
      versionfilter:
        kind: 'semver'
        pattern: '*'
conditions:
  nginx-name:
    name: 'Ensure Helm chart name nginx is specified'
    kind: 'yaml'
    disablesourceinput: true
    spec:
      file: 'manifest.yaml'
      key: '$.spec.source.chart'
      documentindex: 0
      value: 'nginx'
  nginx-repository:
    name: 'Ensure Helm chart repository registry-1.docker.io/bitnamicharts is specified'
    kind: 'yaml'
    disablesourceinput: true
    spec:
      file: 'manifest.yaml'
      key: '$.spec.source.repoURL'
      documentindex: 0
      value: 'registry-1.docker.io/bitnamicharts'
targets:
  nginx:
    name: 'deps(helm): update Helm chart "nginx" to {{ source "nginx" }}'
    kind: 'yaml'
    spec:
      file: 'manifest.yaml'
      key: '$.spec.source.targetRevision'
      documentindex: 0
    sourceid: 'nginx'
`},
		},
	}

	for _, tt := range testdata {

		t.Run(tt.name, func(t *testing.T) {
			argocd, err := New(
				tt.spec, tt.rootDir, tt.scmID, tt.actionID)

			require.NoError(t, err)

			var pipelines []string
			rawPipelines, err := argocd.DiscoverManifests()
			require.NoError(t, err)

			// Sort pipelines by name to ensure consistent order for assertion
			sort.Slice(rawPipelines, func(i, j int) bool {
				return string(rawPipelines[i]) < string(rawPipelines[j])
			})

			require.Equal(t, len(tt.expectedPipelines), len(rawPipelines), "number of discovered pipelines does not match expected")

			for i := range rawPipelines {
				// We expect manifest generated by the autodiscovery to use the yaml syntax
				pipelines = append(pipelines, string(rawPipelines[i]))
				assert.Equal(t, tt.expectedPipelines[i], pipelines[i])
			}
		})
	}
}
