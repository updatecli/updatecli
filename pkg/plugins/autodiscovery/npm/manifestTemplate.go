package npm

var (
	// manifestTemplate is the Go template used to generate Fleet manifests
	manifestTemplate string = `name: '{{ .ManifestName }}'
sources:
  {{ .SourceID }}:
    name: '{{ .SourceName }}'
    kind: '{{ .SourceKind }}'
    spec:
      name: '{{ .SourceNPMName }}'
      versionfilter:
        kind: '{{ .SourceVersionFilterKind }}'
        pattern: '{{ .SourceVersionFilterPattern }}'
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
    name: Update NPM lockfile package-lock.json
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
      command: |
        {{ .TargetNPMCommand }}
      changedif:
        kind: file/checksum
        spec:
          files:
            - "package-lock.json"
            - "package.json"
      environments:
       - name: PATH
         inherit: true
      workdir: '{{ .TargetWorkdir }}'
{{ end }}
{{- if .TargetYarnCleanupEnabled }}
  yarn.lock:
    name: Update Yarn lockfile yarn.lock
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
      command: |
        {{ .TargetYarnCommand }}
      changedif:
        kind: file/checksum
        spec:
          files:
            - "yarn.lock"
      environments:
       - name: PATH
         inherit: true
      workdir: '{{ .TargetWorkdir }}'
{{ end }}
`
)
