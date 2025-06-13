package argocd

import (
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
	}{
		{
			name:              "ArgoCD manifests discovered no source",
			rootDir:           "testdata/empty",
			expectedPipelines: []string{},
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
      value: 'sealed-secrets'
  sealed-secrets-repository:
    name: 'Ensure Helm chart repository https://bitnami-labs.github.io/sealed-secrets is specified'
    kind: 'yaml'
    disablesourceinput: true
    spec:
      file: 'manifest.yaml'
      key: '$.spec.source.repoURL'
      value: 'https://bitnami-labs.github.io/sealed-secrets'
targets:
  sealed-secrets:
    name: 'deps(helm): update Helm chart "sealed-secrets" to {{ source "sealed-secrets" }}'
    kind: 'yaml'
    spec:
      file: 'manifest.yaml'
      key: '$.spec.source.targetRevision'
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
      value: 'sealed-secrets'
  sealed-secrets-repository:
    name: 'Ensure Helm chart repository https://bitnami-labs.github.io/sealed-secrets is specified'
    kind: 'yaml'
    disablesourceinput: true
    spec:
      file: 'manifest.yaml'
      key: '$.spec.sources[0].repoURL'
      value: 'https://bitnami-labs.github.io/sealed-secrets'
targets:
  sealed-secrets:
    name: 'deps(helm): update Helm chart "sealed-secrets" to {{ source "sealed-secrets" }}'
    kind: 'yaml'
    spec:
      file: 'manifest.yaml'
      key: '$.spec.sources[0].targetRevision'
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
      value: 'sealed-secrets'
  sealed-secrets-repository:
    name: 'Ensure Helm chart repository https://bitnami-labs.github.io/sealed-secrets is specified'
    kind: 'yaml'
    disablesourceinput: true
    scmid: scm123
    spec:
      file: 'manifest.yaml'
      key: '$.spec.sources[0].repoURL'
      value: 'https://bitnami-labs.github.io/sealed-secrets'
targets:
  sealed-secrets:
    name: 'deps(helm): update Helm chart "sealed-secrets" to {{ source "sealed-secrets" }}'
    kind: 'yaml'
    scmid: scm123
    spec:
      file: 'manifest.yaml'
      key: '$.spec.sources[0].targetRevision'
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
      value: 'nginx'
  nginx-repository:
    name: 'Ensure Helm chart repository registry-1.docker.io/bitnamicharts is specified'
    kind: 'yaml'
    disablesourceinput: true
    spec:
      file: 'manifest.yaml'
      key: '$.spec.source.repoURL'
      value: 'registry-1.docker.io/bitnamicharts'
targets:
  nginx:
    name: 'deps(helm): update Helm chart "nginx" to {{ source "nginx" }}'
    kind: 'yaml'
    spec:
      file: 'manifest.yaml'
      key: '$.spec.source.targetRevision'
    sourceid: 'nginx'
`},
		},
	}

	for _, tt := range testdata {

		t.Run(tt.name, func(t *testing.T) {
			argocd, err := New(
				Spec{}, tt.rootDir, tt.scmID, tt.actionID)

			require.NoError(t, err)

			var pipelines []string
			rawPipelines, err := argocd.DiscoverManifests()
			require.NoError(t, err)

			for i := range rawPipelines {
				// We expect manifest generated by the autodiscovery to use the yaml syntax
				pipelines = append(pipelines, string(rawPipelines[i]))
				assert.Equal(t, tt.expectedPipelines[i], pipelines[i])
			}
		})
	}

}
