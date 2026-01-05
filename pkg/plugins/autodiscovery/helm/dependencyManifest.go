package helm

const (
	// dependencyManifest is the Go template used to generate
	// the Helm chart manifests specific for Helm dependencies
	dependencyManifest string = `name: '{{ .ManifestName }}'
{{- if .ActionID }}
actions:
  {{ .ActionID }}:
    title: 'deps: update Helm chart dependency to {{ "{{" }} source "helmchart" {{ "}}" }}'
{{ end }}
sources:
  helmchart:
    name: '{{ .SourceName }}'
    kind: 'helmchart'
    spec:
      name: '{{ .DependencyName }}'
      url: '{{ .DependencyRepository }}'
      {{- if .Token }}
      token: '{{ .Token }}'
      {{- end }}
      versionfilter:
        kind: '{{ .SourceVersionFilterKind }}'
        pattern: '{{ .SourceVersionFilterPattern }}'
        {{- if or (eq .SourceVersionFilterKind "regex/semver") (eq .SourceVersionFilterKind "regex/time") }}
        regex: '{{ .SourceVersionFilterRegex }}'
        {{- end }}
conditions:
  {{ .ConditionID }}:
    name: 'Ensure Helm chart dependency "{{ .DependencyName }}" is specified'
    kind: 'yaml'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .File }}'
      key: '{{ .ConditionKey }}'
      value: '{{ .DependencyName}}'
    disablesourceinput: true
targets:
  {{ .TargetID }}:
    name: 'Bump Helm chart dependency "{{ .DependencyName }}" for Helm chart "{{ .ChartName }}"'
    kind: 'helmchart'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .TargetFile }}'
      key: '{{ .TargetKey }}'
      name: '{{ .TargetChartName }}'
      skippackaging: {{ .TargetChartSkipPackaging }}
      versionincrement: '{{ .TargetChartVersionIncrement }}'
    sourceid: 'helmchart'
`
)
