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
{{- if .TargetNPMCleanupEnabled }}
  package-lock.json:
    name: Update NPM lockfile package-lock.json
    dependson:
      - {{ .TargetID }}
    disablesourceinput: true
    kind: shell
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      command: npm install --package-lock-only
      changedif:
        kind: file/checksum
        spec:
          files:
            - "package-lock.json"
      workdir: '{{ .TargetWorkdir }}'
{{ end }}
{{- if .TargetYarnCleanupEnabled }}
  yarn.lock:
    name: Update Yarn lockfile yarn.lock
    dependson:
      - {{ .TargetID }}
    disablesourceinput: true
    kind: shell
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      command: yarn install --mode update-lockfile
      changedif:
        kind: file/checksum
        spec:
          files:
            - "yarn.lock"
      workdir: '{{ .TargetWorkdir }}'
{{ end }}
`
)
