package helm

const (
	// containerManifest is the Go template used to generate
	// Docker compose manifests
	containerManifest string = `name: '{{ .ManifestName }}'
sources:
  {{ .SourceID }}:
    name: '{{ .SourceName }}'
    kind: 'dockerimage'
    spec:
      image: '{{ .SourceImageName }}'
      tagfilter: '{{ .SourceTagFilter }}'
      versionFilter:
        kind: '{{ .SourceVersionFilterKind }}'
        pattern: '{{ .SourceVersionFilterPattern }}'
conditions:
  {{ .ConditionID }}:
    disablesourceinput: true
    name: '{{ .ConditionName }}'
    kind: 'yaml'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .File }}'
      key: '{{ .ConditionKey }}'
      value: '{{ .ConditionValue }}'
targets:
  {{ .TargetID }}:
    name: '{{ .TargetName }}'
    kind: 'helmchart'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .TargetFile }}'
      name: '{{ .TargetChartName }}'
      key: '{{ .TargetKey }}'
      VersionIncrement: 'minor'
    sourceid: '{{ .SourceID }}'
`
)
