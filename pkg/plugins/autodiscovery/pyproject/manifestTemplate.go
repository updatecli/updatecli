package pyproject

import (
	"text/template"

	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/vulnerability"
)

// uvTargetTemplate defines the "uvtarget" Go template, the shell target re-locking uv.lock
// to the version returned by the source. It emits nothing when uv is not used.
const uvTargetTemplate = `{{- define "uvtarget" }}
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
{{- end }}`

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
{{- template "age" .SourceAge }}
      versionfilter:
        kind: '{{ .SourceVersionFilterKind }}'
        pattern: '{{ .SourceVersionFilterPattern }}'
{{- if or (eq .SourceVersionFilterKind "regex/semver") (eq .SourceVersionFilterKind "regex/time") }}
        regex: '{{ .SourceVersionFilterRegex }}'
{{- end }}
{{- template "uvtarget" . }}
`

// securityManifestTemplate is the Go text/template used to generate updatecli manifests
// updating a Python dependency with known vulnerabilities, according to the OSV database.
// The "vulnerable" condition stops the pipeline when the current version has none.
// The pipeline name contains the fixed version, so it can be used as pullrequest title.
var securityManifestTemplate = `name: '{{ .ManifestName }}'
{{- if .ActionID }}
actions:
  {{ .ActionID }}:
    title: '{{ .TargetName }}'
{{ end }}
sources:
  {{ .SourceID }}:
    name: '{{ .SourceName }}'
    kind: 'vulnerability/osv'
    spec:
      ecosystem: 'PyPI'
      name: '{{ .DependencyName }}'
      version: '{{ .CurrentVersion }}'
{{- template "vulnerability" .Vulnerability }}
conditions:
  vulnerable:
    name: 'Test if "{{ .DependencyName }}" package version {{ .CurrentVersion }} has known vulnerabilities'
    kind: 'vulnerability/osv'
    disablesourceinput: true
    failwhen: true
    spec:
      ecosystem: 'PyPI'
      name: '{{ .DependencyName }}'
      version: '{{ .CurrentVersion }}'
{{- template "vulnerability" .Vulnerability }}
{{- template "uvtarget" . }}
`

// The manifest templates are constants, so they are parsed once rather than for every dependency group.
var (
	// routineManifestTmpl generates the manifests bumping a dependency to its latest version.
	routineManifestTmpl = template.Must(template.New("manifest").Parse(age.ManifestTemplate + uvTargetTemplate + manifestTemplate))
	// securityManifestTmpl generates the manifests updating a dependency with known vulnerabilities.
	securityManifestTmpl = template.Must(template.New("manifest").Parse(vulnerability.ManifestTemplate + uvTargetTemplate + securityManifestTemplate))
)

// manifestTemplateParams holds the values injected into manifestTemplate.
type manifestTemplateParams struct {
	ManifestName               string
	ActionID                   string
	SourceID                   string
	SourceName                 string
	SourceVersionFilterKind    string
	SourceVersionFilterPattern string
	SourceVersionFilterRegex   string
	SourceAge                  age.Spec
	DependencyName             string
	IndexURL                   string
	TargetID                   string
	TargetName                 string
	ScmID                      string
	UvEnabled                  bool
	LockFile                   string
	Workdir                    string
	// Vulnerability switches the manifest to a security update of the package, when set.
	Vulnerability *vulnerability.Spec
	// CurrentVersion is the package version checked against the OSV database by security updates.
	CurrentVersion string
}
