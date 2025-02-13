package updatecli

const (
	// manifestTemplate is the Go template used to generate Fleet manifests
	manifestTemplate string = `name: '{{ .ManifestName }}'
{{- if .ActionID }}
actions:
  {{ .ActionID }}:
    title: '{{ .TargetName }}'
{{ end }}
sources:
  version:
    name: '{{ .SourceVersionName }}'
    kind: 'dockerimage'
    spec:
      image: '{{ .PolicyName }}'
      versionfilter:
        kind: '{{ .SourceVersionFilterKind }}'
        pattern: '{{ .SourceVersionFilterPattern }}'
  digest:
    name: '{{ .SourceDigestName }}'
    kind: 'dockerdigest'
    dependson:
     - version
    spec:
      image: '{{ .PolicyName }}'
      tag: '{{ .SourceDigestTag }}'
targets:
  compose:
    name: '{{ .TargetName }}'
    kind: 'yaml'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .File }}'
      key: '{{ .TargetKey }}'
    transformers:
      - addprefix: '{{ .PolicyName }}:'
    sourceid: 'digest'
`
)
