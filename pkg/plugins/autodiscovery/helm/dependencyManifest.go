package helm

const (
	// dependencyManifest is the Go template used to generate
	// the Helm chart manifests specific for Helm dependencies
	dependencyManifest string = `name: {{ .ManifestName }}
sources:
  {{ .SourceID }}:
    name: {{ .SourceName }}
    kind: helmchart
    spec:
      name: {{ .DependencyName }}
      url: {{ .DependencyRepository }}
      versionfilter:
        kind: {{ .SourceVersionFilterKind }}
        pattern: "{{ .SourceVersionFilterPattern }}"
conditions:
  {{ .ConditionID }}:
    name: Ensure Helm chart dependency "{{ .DependencyName }}" is specified
    kind: yaml
{{- if .ScmID }}
    scmid: {{ .ScmID }}
{{ end }}
    spec:
      file: {{ .File }}
      key: {{ .ConditionKey }}
      value: {{ .DependencyName}}
    disablesourceinput: true
targets:
  {{ .TargetID }}:
    name: Bump Helm chart dependency "{{ .DependencyName }}" for Helm chart "{{ .ChartName }}"
    kind: helmchart
{{- if .ScmID }}
    scmid: {{ .ScmID }}
{{ end }}
    spec:
      file: {{ .TargetFile }}
      key: {{ .TargetKey }}
      name: {{ .TargetChartName }}
      versionincrement: minor
    sourceid: {{ .SourceID }}
`
)
