package maven

import (
	"bytes"
	"text/template"

	"github.com/updatecli/updatecli/pkg/core/log"
)

var (
	// manifestTemplate is the Go template used to generate Maven manifests
	manifestTemplate string = `name: '{{ .ManifestName }}'
{{- if .ActionID }}
actions:
  {{ .ActionID }}:
    title: '{{ .TargetName }}'
{{ end }}
sources:
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
conditions:
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
`
)

// routineManifestTmpl generates the manifests bumping a dependency to its latest version.
// The manifest template is a constant, so it is parsed once rather than for every dependency.
var routineManifestTmpl = template.Must(template.New("manifest").Parse(manifestTemplate))

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
	TargetID                   string
	TargetName                 string
	TargetXMLPath              string
	File                       string
	ScmID                      string
}

// renderManifest generates the manifest updating an artifact from its current version.
func (m Maven) renderManifest(params manifestTemplateParams, currentVersion string) ([]byte, error) {
	params.SourceVersionFilterKind = m.versionFilter.Kind
	params.SourceVersionFilterPattern = m.versionFilter.Pattern
	params.SourceVersionFilterRegex = m.versionFilter.Regex

	// A user defined version filter only accepts versions greater than the current one
	if !m.spec.VersionFilter.IsZero() {
		pattern, err := m.versionFilter.GreaterThanPattern(currentVersion)
		if err != nil {
			log.Debugf("building version filter pattern: %s", err)
			pattern = "*"
		}
		params.SourceVersionFilterPattern = pattern
	}

	manifest := bytes.Buffer{}
	if err := routineManifestTmpl.Execute(&manifest, params); err != nil {
		return nil, err
	}

	return manifest.Bytes(), nil
}
