package dockercompose

const (
	// manifestTemplate is the Go template used to generate
	// Docker compose manifests
	manifestTemplate string = `name: 'Bump Docker Image Tag for {{ .ImageName }}'
sources:
  {{ .SourceID }}:
    name: '[{{ .ImageName }}] Get latest Docker Image Tag'
    kind: 'dockerimage'
    spec:
      image: '{{ .ImageName }}'
      tagFilter: '{{ .TagFilter }}'
      versionFilter:
        kind: '{{ .VersionFilterKind }}'
        pattern: '{{ .VersionFilterPattern }}'
targets:
  {{ .TargetID }}:
    name: '{{ .TargetName }}'
    kind: 'yaml'
{{- if .ScmID }}
    scmid: {{ .ScmID }}
{{ end }}
    spec:
      file: '{{ .TargetFile }}'
      key: '{{ .TargetKey }}'
    sourceid: '{{ .SourceID }}'
    transformers:
      - addprefix: '{{ .TargetPrefix }}'
`
)
