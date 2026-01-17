package bazel

// manifestTemplate is the Go template used to generate Bazel module manifests
var manifestTemplate string = `name: 'Update Bazel module {{ .ModuleName }}'
{{- if .ActionID }}
actions:
  {{ .ActionID }}:
    title: '{{ .TargetName }}'
{{ end }}
sources:
  {{ .SourceID }}:
    name: 'Get latest version of Bazel module {{ .ModuleName }}'
    kind: bazelregistry
    spec:
      module: {{ .ModuleName }}
{{- if .VersionFilterKind }}
      versionfilter:
        kind: '{{ .VersionFilterKind }}'
        pattern: '{{ .VersionFilterPattern }}'
{{- if or (eq .VersionFilterKind "regex/semver") (eq .VersionFilterKind "regex/time") }}
        regex: '{{ .VersionFilterRegex }}'
{{- end }}
{{- end }}
conditions:
  {{ .ConditionID }}:
    name: 'Check if Bazel module {{ .ModuleName }} is up to date'
    kind: bazelmod
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .ModuleFile }}'
      module: {{ .ModuleName }}
    disablesourceinput: true
targets:
  {{ .TargetID }}:
    name: '{{ .TargetName }}'
    kind: bazelmod
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .ModuleFile }}'
      module: {{ .ModuleName }}
    sourceid: '{{ .SourceID }}'
`
