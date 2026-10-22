package maven

import (
	"bytes"
	"fmt"
	"sync"
	"text/template"

	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
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
{{- template "age" .SourceAge }}
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
// A parsing error is returned to the caller instead of panicking.
var routineManifestTmpl = sync.OnceValues(func() (*template.Template, error) {
	return template.New("manifest").Parse(age.ManifestTemplate + manifestTemplate)
})

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
}

// renderManifest generates the manifest updating an artifact from its current version.
func (m Maven) renderManifest(params manifestTemplateParams, currentVersion string) ([]byte, error) {
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

	tmpl, err := routineManifestTmpl()
	if err != nil {
		return nil, fmt.Errorf("parsing Maven manifest template: %w", err)
	}

	manifest := bytes.Buffer{}
	if err := tmpl.Execute(&manifest, params); err != nil {
		return nil, err
	}

	return manifest.Bytes(), nil
}
