package helmfile

const (
	// manifestTemplate is the Go template used to generate Fleet manifests
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
  {{ .ConditionID }}:
    name: '{{ .ConditionName }}'
    kind: 'yaml'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .File }}'
      key: '{{ .ConditionKey }}'
      value: '{{ .ConditionValue }}'
    disablesourceinput: true
targets:
  {{ .TargetID }}:
    name: '{{ .TargetName }}'
    kind: 'yaml'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .File }}'
      key: '{{ .TargetKey }}'
    sourceid: '{{ .SourceID }}'
`
)
