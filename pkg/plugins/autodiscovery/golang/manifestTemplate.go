package golang

var (
	// goManifestTemplate is the Go template used to generate Golang manifest update
	goManifestTemplate string = `name: 'deps(golang): bump Go version'
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
)
