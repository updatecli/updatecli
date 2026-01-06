package flux

const (
	// helmreleaseManifestTemplate is the Go template used to generate Flux manifests
	helmreleaseManifestTemplate string = `name: 'deps(flux): bump Helmrelease "{{ .ChartName }}"'
{{- if .ActionID }}
actions:
  {{ .ActionID }}:
    title: 'deps: update Helm chart to {{ "{{" }} source "helmrelease" {{ "}}" }}'
{{- end }}
sources:
  helmrelease:
    name: 'Get latest "{{ .ChartName }}" Helm chart version'
    kind: 'helmchart'
    spec:
      name: '{{ .ChartName }}'
      url: '{{ .ChartRepository }}'
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
  helmrelease:
    name: 'Ensure Helm Chart name "{{ .ChartName }}"'
    kind: 'yaml'
    disablesourceinput: true
{{- if .ScmID }}
    scmid: {{ .ScmID }}
{{- end }}
    spec:
      file: '{{ .File }}'
      key: '$.spec.chart.spec.chart'
      value: '{{ .ChartName }}'
targets:
  helmrelease:
    name: 'deps(flux): bump Helmrelease "{{ .ChartName }}"'
    kind: 'yaml'
{{- if .ScmID }}
    scmid: {{ .ScmID }}
{{- end }}
    spec:
      file: '{{ .File }}'
      key: '$.spec.chart.spec.version'
    sourceid: 'helmrelease'
`
)
