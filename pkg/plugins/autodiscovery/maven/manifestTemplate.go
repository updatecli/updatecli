package maven

var (
	// manifestTemplate is the Go template used to generate Fleet manifests
	manifestTemplate string = `name: {{ .ManifestName }}
sources:
  {{ .SourceID }}:
    name: {{ .SourceName }}
    kind: {{ .SourceKind }}
    spec:
  {{- if .SourceRepositories }}
      repositories:
  {{- range $repo := .SourceRepositories }}
        - {{ $repo }}
  {{- end }}
  {{- end }}
      groupid: {{ .SourceGroupID }}
      artifactid: {{ .SourceArtifactID }}
conditions:
  {{ .ConditionGroupID }}:
    name: {{ .ConditionGroupIDName }}
    kind: xml
{{- if .ScmID }}
    scmid: {{ .ScmID }}
{{ end }}
    spec:
      file: {{ .File }}
      path: {{ .ConditionGroupIDPath }}
      value: {{ .ConditionGroupIDValue }}
    disablesourceinput: true
  {{ .ConditionArtifactID }}:
    name: {{ .ConditionArtifactIDName }}
    kind: xml
{{- if .ScmID }}
    scmid: {{ .ScmID }}
{{ end }}
    spec:
      file: {{ .File }}
      path: {{ .ConditionArtifactIDPath }}
      value: {{ .ConditionArtifactIDValue }}
    disablesourceinput: true
targets:
  {{ .TargetID }}:
    name: {{ .TargetName }}
    kind: xml
{{- if .ScmID }}
    scmid: {{ .ScmID }}
{{ end }}
    spec:
      file: {{ .File }}
      path: {{ .TargetXMLPath }}
    sourceid: {{ .SourceID }}
`
)
