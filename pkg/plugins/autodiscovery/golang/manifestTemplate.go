package golang

var (
	// goManifestTemplate is the Go template used to generate Golang manifest update
	goManifestTemplate string = `name: 'deps(golang): bump Go version'
{{- if .ActionID }}
actions:
  {{ .ActionID }}:
    title: 'Update Go version to {{ "{{" }} source "go" {{ "}}" }}'
{{ end }}
sources:
  go:
    name: 'Get latest Go version'
    kind: 'golang'
    spec:
      versionfilter:
        kind: '{{ .VersionFilterKind }}'
        pattern: '{{ .VersionFilterPattern }}'
targets:
  go:
    name: 'deps(golang): bump Go version to {{ "{{" }} source "go" {{ "}}" }}'
    kind: 'golang/gomod'
    sourceid: 'go'
    spec:
      file: '{{ .GoModFile }}'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
`

	// goModuleManifestTemplate is the Go template used to generate Golang manifest update
	goModuleManifestTemplate string = `name: 'deps(go): bump module {{ .Module }}'
{{- if .ActionID }}
actions:
  {{ .ActionID }}:
    title: 'deps(go): bump module {{ .Module }} to {{ "{{" }} source "module" {{ "}}" }}'
{{ end }}
sources:
  module:
    name: 'Get latest golang module {{ .Module }} version'
    kind: 'golang/module'
    spec:
      module: '{{ .Module }}'
      versionfilter:
        kind: '{{ .VersionFilterKind }}'
        pattern: '{{ .VersionFilterPattern }}'
targets:
  module:
    name: 'deps(go): bump module {{ .Module }} to {{ "{{" }} source "module" {{ "}}" }}'
    kind: 'golang/gomod'
    sourceid: 'module'
    spec:
      file: '{{ .GoModFile }}'
      module: '{{ .Module }}'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
{{- if .GoModTidyEnabled }}
  tidy:
    name: 'clean: go mod tidy'
    disablesourceinput: true
    dependsonchange: true
    dependson:
      - 'module'
    kind: 'shell'
    spec:
      command: 'go mod tidy'
      environments:
        - name: HOME
        - name: PATH
      workdir: {{ .WorkDir }}
      changedif:
        kind: 'file/checksum'
        spec:
          files:
           - 'go.mod'
           - 'go.sum'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
{{- end }}
`

	// goReplaceModuleManifestTemplate is the Go template used to generate Golang manifest update
	goReplaceModuleManifestTemplate string = `name: 'deps(go): bump replaced module {{ .NewPathModule }}'
{{- if .ActionID }}
actions:
  {{ .ActionID }}:
    title: 'deps(go): bump replaced module {{ .NewPathModule }} to {{ "{{" }} source "module" {{ "}}" }}'
{{ end }}
sources:
  module:
    name: 'Get latest golang module {{ .NewPathModule }} version'
    kind: 'golang/module'
    spec:
      module: '{{ .NewPathModule }}'
      versionfilter:
        kind: '{{ .VersionFilterKind }}'
        pattern: '{{ .VersionFilterPattern }}'
targets:
  module:
    name: 'deps(go): bump module {{ .NewPathModule }} to {{ "{{" }} source "module" {{ "}}" }}'
    kind: 'golang/gomod'
    sourceid: 'module'
    spec:
      file: '{{ .GoModFile }}'
      module: '{{ .OldPathModule }}'
      replace: true
      replaceVersion: '{{ .OldVersionModule }}'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
{{- if .GoModTidyEnabled }}
  tidy:
    name: 'clean: go mod tidy'
    disablesourceinput: true
    dependsonchange: true
    dependson:
      - 'module'
    kind: 'shell'
    spec:
      command: 'go mod tidy'
      environments:
        - name: HOME
        - name: PATH
      workdir: {{ .WorkDir }}
      changedif:
        kind: 'file/checksum'
        spec:
          files:
           - 'go.mod'
           - 'go.sum'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
{{- end }}
`
)
