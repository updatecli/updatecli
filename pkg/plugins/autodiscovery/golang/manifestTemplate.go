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
      {{- if or .Age.Minimum .Age.Maximum }}
      age:
        {{- if .Age.Minimum }}
        minimum: '{{ .Age.Minimum }}'
        {{- end }}
        {{- if .Age.Maximum }}
        maximum: '{{ .Age.Maximum }}'
        {{- end }}
      {{- end }}
      versionfilter:
        kind: '{{ .VersionFilterKind }}'
        pattern: '{{ .VersionFilterPattern }}'
{{- if or (eq .VersionFilterKind "regex/semver") (eq .VersionFilterKind "regex/time") }}
        regex: '{{ .VersionFilterRegex }}'
{{- end }}
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

	// goTidyTemplate is the Go template of the "go mod tidy" target shared by module manifests
	goTidyTemplate string = `{{- define "tidy" }}
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
{{- end }}`

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
      {{- if or .Age.Minimum .Age.Maximum }}
      age:
        {{- if .Age.Minimum }}
        minimum: '{{ .Age.Minimum }}'
        {{- end }}
        {{- if .Age.Maximum }}
        maximum: '{{ .Age.Maximum }}'
        {{- end }}
      {{- end }}
      module: '{{ .Module }}'
      versionfilter:
        kind: '{{ .VersionFilterKind }}'
        pattern: '{{ .VersionFilterPattern }}'
{{- if or (eq .VersionFilterKind "regex/semver") (eq .VersionFilterKind "regex/time") }}
        regex: '{{ .VersionFilterRegex }}'
{{- end }}
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
{{- template "tidy" . }}
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
      {{- if or .Age.Minimum .Age.Maximum }}
      age:
        {{- if .Age.Minimum }}
        minimum: '{{ .Age.Minimum }}'
        {{- end }}
        {{- if .Age.Maximum }}
        maximum: '{{ .Age.Maximum }}'
        {{- end }}
      {{- end }}
      module: '{{ .NewPathModule }}'
      versionfilter:
        kind: '{{ .VersionFilterKind }}'
        pattern: '{{ .VersionFilterPattern }}'
{{- if or (eq .VersionFilterKind "regex/semver") (eq .VersionFilterKind "regex/time") }}
        regex: '{{ .VersionFilterRegex }}'
{{- end }}
targets:
  module:
    name: 'deps(go): bump module {{ .NewPathModule }} to {{ "{{" }} source "module" {{ "}}" }}'
    kind: 'golang/gomod'
    sourceid: 'module'
    spec:
      file: '{{ .GoModFile }}'
      module: '{{ .OldPathModule }}'
      replace: true
{{- if .OldVersionModule }}
      replaceVersion: '{{ .OldVersionModule }}'
{{ end }}
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
{{- template "tidy" . }}
`

	// goModuleSecurityManifestTemplate is the Go template used to generate Golang module security update manifest.
	// The pipeline name contains the fixed version, so it can be used as pullrequest title.
	goModuleSecurityManifestTemplate string = `name: 'deps(go): bump {{ if .Replace }}replaced {{ end }}module {{ .Module }} to {{ "{{" }} source "fixed" {{ "}}" }}'
{{- if .ActionID }}
actions:
  {{ .ActionID }}:
    title: 'deps(go): bump {{ if .Replace }}replaced {{ end }}module {{ .Module }} to {{ "{{" }} source "fixed" {{ "}}" }}'
{{ end }}
sources:
  fixed:
    name: 'Get lowest golang module {{ .Module }} version without known vulnerabilities'
    kind: 'vulnerability/osv'
    spec:
      ecosystem: 'Go'
      name: '{{ .Module }}'
      version: '{{ .Version }}'
{{- template "vulnerability" .Vulnerability }}
targets:
  module:
    name: 'deps(go): bump module {{ .Module }} to {{ "{{" }} source "fixed" {{ "}}" }}'
    kind: 'golang/gomod'
    sourceid: 'fixed'
    spec:
      file: '{{ .GoModFile }}'
      module: '{{ .TargetModule }}'
      {{- if .Replace }}
      replace: true
      {{- if .ReplaceVersion }}
      replaceVersion: '{{ .ReplaceVersion }}'
      {{- end }}
      {{- end }}
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
{{- template "tidy" . }}
`
)
