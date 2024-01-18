package golang

var (
	// goManifestTemplate is the Go template used to generate Golang manifest update
	goManifestTemplate string = `name: 'Update Golang version'
sources:
  golangVersion:
    name: 'Get latest golang version'
    kind: 'golang'
    spec:
      versionfilter:
        kind: '{{ .VersionFilterKind }}'
        pattern: '{{ .VersionFilterPattern }}'
targets:
  golangVersion:
    name: {{ .TargetName }}
    kind: golang/gomod
    sourceid: golangVersion
    spec:
      file: '{{ .GoModFile }}'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
`

	// goModuleManifestTemplate is the Go template used to generate Golang manifest update
	goModuleManifestTemplate string = `name: 'Update Golang module {{ .Module }}'
sources:
  golangModuleVersion:
    name: 'Get latest golang module {{ .Module }} version'
    kind: 'golang/module'
    spec:
      module: '{{ .Module }}'
      versionfilter:
        kind: '{{ .VersionFilterKind }}'
        pattern: '{{ .VersionFilterPattern }}'
targets:
  golangModuleVersion:
    name: {{ .TargetName }}
    kind: golang/gomod
    sourceid: golangModuleVersion
    spec:
      file: '{{ .GoModFile }}'
      module: '{{ .Module }}'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
{{- if .GoModTidyEnabled }}
  goModTidy:
    name: Run Go mod tidy
    disablesourceinput: true
    dependson:
      - golangModuleVersion
    kind: shell
    spec:
      command: go mod tidy
      environments:
        - name: HOME
        - name: PATH
      changedif:
        kind: file/checksum
        spec:
          files:
           - go.mod
           - go.sum
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
{{ end }}

`
)
