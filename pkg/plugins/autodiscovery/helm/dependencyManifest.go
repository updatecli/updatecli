package helm

const (
	// dependencyManifest is the Go template used to generate
	// the Helm chart manifests specific for Helm dependencies
	dependencyManifest string = `name: '{{ .ManifestName }}'
sources:
  {{ .SourceID }}:
    name: '{{ .SourceName }}'
    kind: 'helmchart'
    spec:
      name: '{{ .DependencyName }}'
      url: '{{ .DependencyRepository }}'
      versionFilter:
        kind: '{{ .SourceVersionFilterKind }}'
        pattern: '{{ .SourceVersionFilterPattern }}'
conditions:
  {{ .ConditionID }}:
    name: 'Ensure Helm chart repository {{ .DependencyRepository }} is specified'
    kind: 'yaml'
    disablesourceinput: true
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
      file: '{{ .TargetFile }}'
      name: '{{ .TargetChartName }}'
      key: '{{ .TargetKey }}'
      VersionIncrement: 'minor'
    sourceid: '{{ .SourceID }}'
`
)
