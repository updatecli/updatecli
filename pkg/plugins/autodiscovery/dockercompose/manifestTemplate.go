package dockercompose

const (
	// manifestTemplate is the Go template used to generate Docker compose manifests
	manifestTemplate string = `name: '{{ .ManifestName }}'
sources:
  {{ .SourceID }}:
    name: '{{ .SourceName }}'
    kind: 'dockerimage'
    spec:
{{- if .ScmID }}
      architecture: '{{ .ImageArchitecture }}'
{{ end }}
      image: '{{ .ImageName }}'
      tagfilter: '{{ .TagFilter }}'
      versionfilter:
        kind: '{{ .VersionFilterKind }}'
        pattern: '{{ .VersionFilterPattern }}'
targets:
  {{ .TargetID }}:
    name: '{{ .TargetName }}'
    kind: 'yaml'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .TargetFile }}'
      key: '{{ .TargetKey }}'
    sourceid: '{{ .SourceID }}'
    transformers:
      - addprefix: '{{ .TargetPrefix }}'
`
)
