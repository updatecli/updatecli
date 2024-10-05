package precommit

var (
	// manifestTemplate is the Go template used to generate Fleet manifests
	manifestTemplate string = `name: '{{ .ManifestName }}'
scms:
  '{{ .SourceScmId }}':
    kind: git
    spec:
      url: {{ .SourceScmUrl }}

sources:
  '{{ .SourceID }}':
    name: '{{ .SourceName }}'
    kind: '{{ .SourceKind }}'
    scmid: '{{ .SourceScmId }}'
    spec:
      versionfilter:
        kind: '{{ .SourceVersionFilterKind }}'
        pattern: '{{ .SourceVersionFilterPattern }}'

targets:
  '{{ .TargetID }}':
    name: '{{ .TargetName }}'
    kind: yaml
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    sourceid: '{{ .SourceID }}'
    spec:
      file: '{{ .File }}'
      key: "{{ .TargetKey }}"
      engine: '{{ .TargetEngine }}'
`
)
