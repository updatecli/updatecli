package maven

var (
	// manifestTemplate is the Go template used to generate Fleet manifests
	manifestTemplate string = `name: '{{ .ManifestName }}'
sources:
  {{ .SourceID }}:
    name: '{{ .SourceName }}'
    kind: '{{ .SourceKind }}'
    spec:
      groupid: '{{ .SourceGroupID }}'
      artifactid: '{{ .SourceArtifactID }}'
  {{- if .SourceRepositories }}
      repositories:
  {{- range $repo := .SourceRepositories }}
        - '{{ $repo }}'
  {{ end }}
  {{ end }}
  
conditions:
  {{ .ConditionGroupID }}:
    name: '{{ .ConditionGroupIDName }}'
    kind: 'xml'
    disablesourceinput: true
{{- if .ScmID }}
    scmid: {{ .ScmID }}
{{ end }}
    spec:
      file: '{{ .File }}'
      path: '{{ .ConditionGroupIDPath }}'
      value: '{{ .ConditionGroupIDValue }}'
  {{ .ConditionArtifactID }}:
    name: '{{ .ConditionArtifactIDName }}'
    kind: 'xml'
    disablesourceinput: true
{{- if .ScmID }}
    scmid: {{ .ScmID }}
{{ end }}
    spec:
      file: '{{ .File }}'
      path: '{{ .ConditionArtifactIDPath }}'
      value: '{{ .ConditionArtifactIDValue }}'
targets:
  {{ .TargetID }}:
    name: '{{ .TargetName }}'
    kind: 'xml'
{{- if .ScmID }}
    scmid: {{ .ScmID }}
{{ end }}
    spec:
      file: '{{ .File }}'
      Path: '{{ .TargetXMLPath }}'
    sourceid: '{{ .SourceID }}'
`
)
