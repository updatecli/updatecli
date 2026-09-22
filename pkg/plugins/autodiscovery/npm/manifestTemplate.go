package npm

import (
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/vulnerability"
)

var (
	// manifestTemplate is the Go template used to generate Fleet manifests
	manifestTemplate string = `name: '{{ .ManifestName }}'
sources:
  {{ .SourceID }}:
    name: '{{ .SourceName }}'
    kind: '{{ .SourceKind }}'
    spec:
{{- if .Vulnerability }}
      ecosystem: 'npm'
      name: '{{ .SourceNPMName }}'
      version: '{{ .CurrentVersion }}'
{{- template "vulnerability" .Vulnerability }}
conditions:
  vulnerable:
    name: 'Test if "{{ .SourceNPMName }}" package version {{ .CurrentVersion }} has known vulnerabilities'
    kind: 'vulnerability/osv'
    disablesourceinput: true
    failwhen: true
    spec:
      ecosystem: 'npm'
      name: '{{ .SourceNPMName }}'
      version: '{{ .CurrentVersion }}'
{{- template "vulnerability" .Vulnerability }}
{{- else }}
      name: '{{ .SourceNPMName }}'
{{- if .SourceNpmrcPath }}
      npmrcpath: '{{ .SourceNpmrcPath }}'
{{- end }}
{{- if .SourceURL }}
      url: '{{ .SourceURL }}'
{{- end }}
{{- if .SourceRegistryToken }}
      registrytoken: '{{ .SourceRegistryToken }}'
{{- end }}
{{- template "age" .SourceAge }}
      versionfilter:
        kind: '{{ .SourceVersionFilterKind }}'
        pattern: '{{ .SourceVersionFilterPattern }}'
{{- if or (eq .SourceVersionFilterKind "regex/semver") (eq .SourceVersionFilterKind "regex/time") }}
        regex: '{{ .SourceVersionFilterRegex }}'
{{- end }}
{{- end }}
targets:
{{- if .TargetPackageJsonEnabled }}
  {{ .TargetID }}:
    name: '{{ .TargetName }}'
    kind: 'json'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .File }}'
      key: '{{ .TargetKey }}'
    sourceid: '{{ .SourceID }}'
{{ end }}
{{- if .TargetNPMCleanupEnabled }}
  package-lock.json:
    name: '{{ .TargetName }}'
{{- if .TargetPackageJsonEnabled }}
    dependson:
      - {{ .TargetID }}
{{ end }}
    disablesourceinput: true
    kind: shell
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      command: |-
        {{ .TargetNPMCommand }}
      changedif:
        kind: file/checksum
        spec:
          files:
            - "{{ .TargetLockFilePrefix }}package-lock.json"
            - "package.json"
      environments:
        - name: PATH
        {{- if .SourceNpmrcPath }}
        - name: NPM_CONFIG_USERCONFIG
          value: '{{ .SourceNpmrcPath }}'
        {{ end }}
      workdir: '{{ .TargetWorkdir }}'
{{ end }}
{{- if .TargetYarnCleanupEnabled }}
  yarn.lock:
    name: '{{ .TargetName }}'
{{- if .TargetPackageJsonEnabled }}
    dependson:
      - {{ .TargetID }}
{{ end }}
    disablesourceinput: true
    kind: shell
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      command: |-
        {{ .TargetYarnCommand }}
      changedif:
        kind: file/checksum
        spec:
          files:
            - "{{ .TargetLockFilePrefix }}yarn.lock"
            - "package.json"
      environments:
        - name: PATH
        {{- if .SourceNpmrcPath }}
        - name: NPM_CONFIG_USERCONFIG
          value: '{{ .SourceNpmrcPath }}'
        {{ end }}
      workdir: '{{ .TargetWorkdir }}'
{{ end }}
{{- if .TargetPnpmCleanupEnabled }}
  pnpm-lock.yaml:
    name: '{{ .TargetName }}'
{{- if .TargetPackageJsonEnabled }}
    dependson:
      - {{ .TargetID }}
{{ end }}
    disablesourceinput: true
    kind: shell
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      command: |-
        {{ .TargetPnpmCommand }}
      changedif:
        kind: file/checksum
        spec:
          files:
            - "{{ .TargetLockFilePrefix }}pnpm-lock.yaml"
            - "package.json"
      environments:
        - name: PATH
        {{- if .SourceNpmrcPath }}
        - name: NPM_CONFIG_USERCONFIG
          value: '{{ .SourceNpmrcPath }}'
        {{ end }}
      workdir: '{{ .TargetWorkdir }}'
{{ end }}
`
)

type manifestTemplateParams struct {
	ManifestName               string
	SourceID                   string
	SourceName                 string
	SourceKind                 string
	SourceNPMName              string
	SourceVersionFilterKind    string
	SourceVersionFilterPattern string
	SourceVersionFilterRegex   string
	SourceNpmrcPath            string
	SourceURL                  string
	SourceRegistryToken        string
	SourceAge                  age.Spec
	// Vulnerability switches the manifest to a security update of the package, when set.
	Vulnerability *vulnerability.Spec
	// CurrentVersion is the package version checked against the OSV database by security updates.
	CurrentVersion           string
	TargetID                 string
	TargetName               string
	TargetKey                string
	TargetPackageJsonEnabled bool
	TargetYarnCleanupEnabled bool
	TargetPnpmCleanupEnabled bool
	TargetNPMCleanupEnabled  bool
	TargetWorkdir            string
	// TargetLockFilePrefix is the path from TargetWorkdir to the lock file directory, such as "../../" for a workspace project.
	TargetLockFilePrefix string
	TargetNPMCommand     string
	TargetYarnCommand    string
	TargetPnpmCommand    string
	File                 string
	ScmID                string
}
