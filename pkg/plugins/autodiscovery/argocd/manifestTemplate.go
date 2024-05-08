package argocd

const (
	// manifestTemplate is the Go template used to generate ArgoCD application manifests
	manifestTemplate string = `name: '{{ .ManifestName }}'
sources:
  {{ .SourceID }}:
    name: '{{ .SourceName }}'
    kind: '{{ .SourceKind }}'
    spec:
      name: '{{ .ChartName }}'
      url: '{{ .ChartRepository }}'
      versionfilter:
        kind: '{{ .SourceVersionFilterKind }}'
        pattern: '{{ .SourceVersionFilterPattern }}'
conditions:
  {{ .ConditionID }}-name:
    name: 'Ensure Helm chart name {{ .ChartName }} is specified'
    kind: 'yaml'
    disablesourceinput: true
{{- if .ScmID }}
    scmid: {{ .ScmID }}
{{ end }}
    spec:
      file: '{{ .File }}'
      key: '$.spec.source.chart'
      value: '{{ .ChartName }}'
  {{ .ConditionID }}-repository:
    name: 'Ensure Helm chart repository {{ .ChartRepository }} is specified'
    kind: 'yaml'
    disablesourceinput: true
{{- if .ScmID }}
    scmid: {{ .ScmID }}
{{ end }}
    spec:
      file: '{{ .File }}'
      key: '$.spec.source.repoURL'
      value: '{{ .ChartRepository }}'
targets:
  {{ .TargetID }}:
    name: '{{ .ManifestName }}'
    kind: 'yaml'
{{- if .ScmID }}
    scmid: {{ .ScmID }}
{{ end }}
    spec:
      file: '{{ .File }}'
      key: '$.spec.source.targetRevision'
    sourceid: '{{ .SourceID }}'
`
)
