package fleet

const (
	// manifestTemplate is the Go template used to generate Fleet manifests
	manifestTemplate string = `name: '{{ .ManifestName }}'
{{- if .ActionID }}
actions:
  {{ .ActionID }}:
    title: 'deps: update Helm chart "{{ .ChartName }}" to {{ "{{" }} source "{{ .SourceID }}" {{ "}}" }}'
{{ end }}
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
      key: '$.helm.chart'
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
      key: '$.helm.repo'
      value: '{{ .ChartRepository }}'
targets:
  {{ .TargetID }}:
    name: 'deps(helm): bump chart "{{ .ChartName }}" in Fleet bundle "{{ .FleetBundle }}"'
    kind: 'yaml'
{{- if .ScmID }}
    scmid: {{ .ScmID }}
{{ end }}
    spec:
      file: '{{ .File }}'
      key: '$.helm.version'
    sourceid: '{{ .SourceID }}'
`
)
