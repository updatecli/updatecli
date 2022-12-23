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
      image: '{{ .ImageName }}'
      tagFilter: '{{ .TagFilter }}'
      versionFilter:
        kind: '{{ .SourceVersionFilterKind }}'
        pattern: '{{ .SourceVersionFilterPattern }}'
conditions:
  {{ .ConditionID }}:
    disablesourceinput: true
    name: 'Ensure Helm chart repository {{ .DependencyRepository }} is specified'
    kind: 'yaml'
{{- if .ScmID }}
    scmid: {{ .ScmID }}
{{ end }}
    spec:
      file: '{{ .File }}'
      key: '{{ .ConditionKey }}'
      value: '{{ .DependencyName}}'
targets:
  {{ .TargetID }}:
    name: 'Bump Helm chart dependency {{ .DependencyName }} for Helm chart {{ .ChartName }}'
    kind: 'helmchart'
{{- if .ScmID }}
    scmid: {{ .ScmID }}
{{ end }}
    spec:
      file: '{{ .File }}'
      name: '{{ .TargetChartName }}'
      key: '{{ .TargetKey }}'
      VersionIncrement: 'minor'
    sourceid: '{{ .SourceID }}'
`
)
