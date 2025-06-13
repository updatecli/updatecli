package argocd

const (
	// manifestTemplate is the Go template used to generate ArgoCD application manifests
	manifestTemplate string = `name: '{{ .ManifestName }}'
{{- if .ActionID }}
actions:
  {{ .ActionID }}:
    title: 'deps(argocd): update Helm chart {{ .ChartName }} to {{ "{{" }} source "{{ .SourceID }}" {{ "}}" }}'
{{- end }}
sources:
  {{ .SourceID }}:
    name: '{{ .SourceName }}'
    kind: '{{ .SourceKind }}'
    spec:
      name: '{{ .ChartName }}'
      url: '{{ .SourceChartRepository }}'
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
{{- end }}
    spec:
      file: '{{ .File }}'
      key: '{{ .TargetKey }}.chart'
      value: '{{ .ChartName }}'
  {{ .ConditionID }}-repository:
    name: 'Ensure Helm chart repository {{ .ChartRepository }} is specified'
    kind: 'yaml'
    disablesourceinput: true
{{- if .ScmID }}
    scmid: {{ .ScmID }}
{{- end }}
    spec:
      file: '{{ .File }}'
      key: '{{ .TargetKey }}.repoURL'
      value: '{{ .ChartRepository }}'
targets:
  {{ .TargetID }}:
    name: 'deps(helm): update Helm chart "{{ .ChartName }}" to {{ "{{" }} source "{{ .SourceID }}" {{ "}}" }}'
    kind: 'yaml'
{{- if .ScmID }}
    scmid: {{ .ScmID }}
{{- end }}
    spec:
      file: '{{ .File }}'
      key: '{{ .TargetKey }}.targetRevision'
    sourceid: '{{ .SourceID }}'
`
)
