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
{{- if .HasRegistry }}
  {{ .ConditionRegistryID }}:
    disablesourceinput: true
    name: '{{ .ConditionRegistryName }}'
    kind: 'yaml'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .File }}'
      key: '{{ .ConditionRegistryKey }}'
      value: '{{ .ConditionRegistryValue }}'
{{- end }}
  {{ .ConditionRepositoryID }}:
    disablesourceinput: true
    name: '{{ .ConditionRepositoryName }}'
    kind: 'yaml'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .File }}'
      key: '{{ .ConditionRepositoryKey }}'
      value: '{{ .ConditionRepositoryValue }}'
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
      versionincrement: '{{ .TargetChartVersionIncrement }}'
    sourceid: '{{ .SourceID }}'
`
)
