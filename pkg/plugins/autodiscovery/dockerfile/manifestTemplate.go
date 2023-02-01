package dockerfile

var (
	// manifestTemplate is the Go template used to generate
	// Updatecli manifests
	manifestTemplate string = `name: '{{ .ManifestName }}'
sources:
  {{ .SourceID }}:
    name: '[{{ .ImageName }}] Get latest Docker image tag'
    kind: 'dockerimage'
    spec:
      image: '{{ .ImageName }}'
      tagfilter: '{{ .TagFilter }}'
      versionfilter:
        kind: '{{ .VersionFilterKind }}'
        pattern: '{{ .VersionFilterPattern }}'
targets:
  {{ .TargetID }}:
    name: '{{ .TargetName }}'
    kind: 'dockerfile'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .TargetFile }}'
      instruction:
        keyword: '{{ .TargetKeyword }}'
        matcher: '{{ .TargetMatcher }}'
    sourceid: '{{ .SourceID }}'
`
)
