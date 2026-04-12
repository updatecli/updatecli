package pyproject

// manifestTemplate is the Go text/template used to generate updatecli manifests
// for Python dependency updates discovered via pyproject.toml.
var manifestTemplate = `name: '{{ .ManifestName }}'
{{- if .ActionID }}
actions:
  {{ .ActionID }}:
    title: '{{ .TargetName }}'
{{ end }}
sources:
  {{ .SourceID }}:
    name: '{{ .SourceName }}'
    kind: 'pypi'
    spec:
      name: '{{ .DependencyName }}'
{{- if .IndexURL }}
      url: '{{ .IndexURL }}'
{{- end }}
      versionfilter:
        kind: '{{ .SourceVersionFilterKind }}'
        pattern: '{{ .SourceVersionFilterPattern }}'
{{- if or (eq .SourceVersionFilterKind "regex/semver") (eq .SourceVersionFilterKind "regex/time") }}
        regex: '{{ .SourceVersionFilterRegex }}'
{{- end }}
{{- if .UvEnabled }}
targets:
  {{ .TargetID }}:
    name: '{{ .TargetName }}'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{- end }}
    kind: 'shell'
    spec:
      command: 'uv lock --upgrade-package {{ .DependencyName }}=={{ "{{" }} source "{{ .SourceID }}" {{ "}}" }}'
      changedif:
        kind: file/checksum
        spec:
          files:
            - "{{ .LockFile }}"
      environments:
        - name: PATH
      workdir: '{{ .Workdir }}'
    disablesourceinput: true
{{- end }}
`

// manifestTemplateParams holds the values injected into manifestTemplate.
type manifestTemplateParams struct {
	ManifestName               string
	ActionID                   string
	SourceID                   string
	SourceName                 string
	SourceVersionFilterKind    string
	SourceVersionFilterPattern string
	SourceVersionFilterRegex   string
	DependencyName             string
	IndexURL                   string
	TargetID                   string
	TargetName                 string
	ScmID                      string
	UvEnabled                  bool
	LockFile                   string
	Workdir                    string
}
