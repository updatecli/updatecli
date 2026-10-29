package maven

import (
	"bytes"
	"fmt"
	"sync"
	"text/template"

	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/vulnerability"
)

// headerTemplate defines the "header" Go template, the manifest name and the action, shared by
// the routine and security manifests.
const headerTemplate = `{{ define "header" -}}
name: '{{ .ManifestName }}'
{{- if .ActionID }}
actions:
  {{ .ActionID }}:
    title: '{{ .TargetName }}'
{{ end }}
{{ end }}`

// xmlTemplate defines the "xmlconditions" Go template, the conditions checking the pom.xml still
// declares the artifact, and the "xmltarget" Go template, the target updating its version.
const xmlTemplate = `{{ define "xmlconditions" }}
  {{ .ConditionArtifactID }}:
    name: '{{ .ConditionArtifactIDName }}'
    kind: 'xml'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .File }}'
      path: '{{ .ConditionArtifactIDPath }}'
      value: '{{ .ConditionArtifactIDValue }}'
    disablesourceinput: true
  {{ .ConditionGroupID }}:
    name: '{{ .ConditionGroupIDName }}'
    kind: 'xml'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .File }}'
      path: '{{ .ConditionGroupIDPath }}'
      value: '{{ .ConditionGroupIDValue }}'
    disablesourceinput: true
{{ end }}
{{- define "xmltarget" -}}
targets:
  {{ .TargetID }}:
    name: '{{ .TargetName }}'
    kind: 'xml'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .File }}'
      path: '{{ .TargetXMLPath }}'
    sourceid: '{{ .SourceID }}'
{{ end }}`

// manifestTemplate is the Go template used to generate Maven manifests
const manifestTemplate = `{{ template "header" . }}sources:
  {{ .SourceID }}:
    name: '{{ .SourceName }}'
    kind: '{{ .SourceKind }}'
    spec:
      artifactid: '{{ .SourceArtifactID }}'
      groupid: '{{ .SourceGroupID }}'
  {{- if .SourceRepositories }}
      repositories:
  {{- range $repo := .SourceRepositories }}
        - '{{ $repo }}'
  {{- end }}
  {{- end }}
      versionfilter:
        kind: '{{ .SourceVersionFilterKind }}'
        pattern: '{{ .SourceVersionFilterPattern }}'
{{- if or (eq .SourceVersionFilterKind "regex/semver") (eq .SourceVersionFilterKind "regex/time") }}
        regex: '{{ .SourceVersionFilterRegex }}'
{{- end }}
{{- template "age" .SourceAge }}
conditions:{{ template "xmlconditions" . }}{{ template "xmltarget" . }}`

// securityManifestTemplate is the Go template used to generate Maven manifests updating an
// artifact with known vulnerabilities, according to the OSV database.
// The "vulnerable" condition stops the pipeline when the current version has none.
// The pipeline keeps the routine name, which tells the dependency kinds apart, as the
// pullrequest title is the action one.
// The OSV values are rendered as double-quoted YAML strings, so a pom.xml value can't break the manifest.
const securityManifestTemplate = `{{ template "header" . }}sources:
  {{ .SourceID }}:
    name: '{{ .SourceName }}'
    kind: 'vulnerability/osv'
    spec:
      ecosystem: 'Maven'
      name: {{ printf "%s:%s" .SourceGroupID .SourceArtifactID | printf "%q" }}
      version: {{ printf "%q" .CurrentVersion }}
{{- template "vulnerability" .Vulnerability }}
conditions:
  vulnerable:
    name: {{ printf "Test if Maven artifact %q version %s has known vulnerabilities" .SourceID .CurrentVersion | printf "%q" }}
    kind: 'vulnerability/osv'
    disablesourceinput: true
    failwhen: true
    spec:
      ecosystem: 'Maven'
      name: {{ printf "%s:%s" .SourceGroupID .SourceArtifactID | printf "%q" }}
      version: {{ printf "%q" .CurrentVersion }}
{{- template "vulnerability" .Vulnerability }}{{ template "xmlconditions" . }}{{ template "xmltarget" . }}`

// The manifest templates are constants, so they are parsed once rather than for every dependency.
// A parsing error is returned to the caller instead of panicking.
var (
	// routineManifestTmpl generates the manifests bumping a dependency to its latest version.
	routineManifestTmpl = sync.OnceValues(func() (*template.Template, error) {
		return template.New("manifest").Parse(age.ManifestTemplate + headerTemplate + xmlTemplate + manifestTemplate)
	})
	// securityManifestTmpl generates the manifests updating a dependency with known vulnerabilities.
	securityManifestTmpl = sync.OnceValues(func() (*template.Template, error) {
		return template.New("manifest").Parse(vulnerability.ManifestTemplate + headerTemplate + xmlTemplate + securityManifestTemplate)
	})
)

// manifestTemplateParams holds the values injected into manifestTemplate.
type manifestTemplateParams struct {
	ActionID                   string
	ManifestName               string
	ConditionID                string
	ConditionGroupID           string
	ConditionGroupIDName       string
	ConditionGroupIDPath       string
	ConditionGroupIDValue      string
	ConditionArtifactID        string
	ConditionArtifactIDName    string
	ConditionArtifactIDPath    string
	ConditionArtifactIDValue   string
	SourceID                   string
	SourceName                 string
	SourceKind                 string
	SourceGroupID              string
	SourceArtifactID           string
	SourceRepositories         []string
	SourceVersionFilterKind    string
	SourceVersionFilterPattern string
	SourceVersionFilterRegex   string
	SourceAge                  age.Spec
	TargetID                   string
	TargetName                 string
	TargetXMLPath              string
	File                       string
	ScmID                      string
	// Vulnerability switches the manifest to a security update of the artifact, when set.
	Vulnerability *vulnerability.Spec
	// CurrentVersion is the artifact version checked against the OSV database by security updates.
	CurrentVersion string
}

// withRoutineUpdate sets the source of a manifest bumping an artifact to its latest version.
func (m Maven) withRoutineUpdate(params *manifestTemplateParams, currentVersion string) {
	params.SourceVersionFilterKind = m.versionFilter.Kind
	params.SourceVersionFilterPattern = m.versionFilter.Pattern
	params.SourceVersionFilterRegex = m.versionFilter.Regex
	params.SourceAge = m.spec.Age

	// A user defined version filter only accepts versions greater than the current one
	if !m.spec.VersionFilter.IsZero() {
		pattern, err := m.versionFilter.GreaterThanPattern(currentVersion)
		if err != nil {
			log.Debugf("building version filter pattern: %s", err)
			pattern = "*"
		}
		params.SourceVersionFilterPattern = pattern
	}
}

// withSecurityUpdate sets the source of a manifest updating an artifact with known vulnerabilities.
func (m Maven) withSecurityUpdate(params *manifestTemplateParams, currentVersion string) {
	params.SourceName = fmt.Sprintf("Get lowest Maven artifact %q version without known vulnerabilities", params.SourceID)
	params.Vulnerability = m.spec.Vulnerability
	params.CurrentVersion = currentVersion
}

// renderManifest generates the manifest updating an artifact from its current version: a security
// update in vulnerability mode, otherwise a bump to its latest version.
func (m Maven) renderManifest(params manifestTemplateParams, currentVersion string) ([]byte, error) {
	tmpl := routineManifestTmpl

	if m.spec.Vulnerability != nil {
		tmpl = securityManifestTmpl
		m.withSecurityUpdate(&params, currentVersion)
	} else {
		m.withRoutineUpdate(&params, currentVersion)
	}

	parsedTmpl, err := tmpl()
	if err != nil {
		return nil, fmt.Errorf("parsing Maven manifest template: %w", err)
	}

	manifest := bytes.Buffer{}
	if err := parsedTmpl.Execute(&manifest, params); err != nil {
		return nil, err
	}

	return manifest.Bytes(), nil
}
