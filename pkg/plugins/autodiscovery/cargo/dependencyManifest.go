package cargo

const (
	// dependencyManifest is the Go template used to generate
	// the manifests to update the cargo file
	dependencyManifestWithCargoUpgrade string = `name: '{{ .ManifestName }}'
{{- if .ActionID }}
actions:
  {{ .ActionID }}:
    title: '{{ .TargetName }}'
{{ end }}
sources:
  {{ .SourceID }}:
    name: '{{ .SourceName }}'
    kind: 'cargopackage'
    spec:
      package: '{{ .DependencyName }}'
      versionfilter:
        kind: '{{ .SourceVersionFilterKind }}'
        pattern: '{{ .SourceVersionFilterPattern }}'
{{- if .WithRegistry }}
      registry:
        url: '{{ .RegistryURL }}'
        rootdir: '{{ .RegistryRootDir }}'
        auth:
          token: '{{ .RegistryAuthToken }}'
          headerFormat : '{{ .RegistryHeaderFormat }}'
{{- if .RegistrySCMID }}
    scmid: '{{ .RegistrySCMID }}'
{{- end }}
{{- end }}
  {{ .ExistingSourceID }}:
    name: '{{ .ExistingSourceName }}'
    kind: 'toml'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{- end }}
    spec:
      file: '{{ .File }}'
      Key: '{{ .ExistingSourceKey }}'
conditions:
  {{ .ConditionID }}:
    name: 'Test if version of "{{ .DependencyName }}" {{"{{"}} source "{{ .ExistingSourceID }}" {{"}}"}} differs from {{"{{"}} source "{{ .SourceID }}" {{"}}"}}'
    kind: 'shell'
    sourceid: '{{ .SourceID }}'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{- end }}
    spec:
      command: 'test {{"{{"}} source "{{ .ExistingSourceID }}" {{"}}"}} != '
targets:
  {{ .TargetID }}:
    name: '{{ .TargetName }}'
    kind: 'shell'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{- end }}
    spec:
      command: |
        ARGS=""
        if [ "$DRY_RUN" = "true" ]; then
          ARGS="--dry-run"
        fi
        cargo upgrade $ARGS --manifest-path {{ .File }} --package {{ .DependencyName }}@{{"{{"}} source "{{ .SourceID }}" {{"}}"}}
        cargo update $ARGS --manifest-path {{ .File }} {{ .DependencyName }}@{{"{{"}} source "{{ .SourceID }}" {{"}}"}} --precise {{"{{"}} source "{{ .SourceID }}" {{"}}"}}
      changedif:
        kind: file/checksum
        spec:
          files:
            - "{{ .File }}"
            - "Cargo.lock"
    disablesourceinput: true
`
	dependencyManifestWithoutCargoUpgrade string = `name: '{{ .ManifestName }}'
{{- if .ActionID }}
actions:
  {{ .ActionID }}:
    title: '{{ .TargetName }}'
{{ end }}
sources:
  {{ .SourceID }}:
    name: '{{ .SourceName }}'
    kind: 'cargopackage'
    spec:
      package: '{{ .DependencyName }}'
      versionfilter:
        kind: '{{ .SourceVersionFilterKind }}'
        pattern: '{{ .SourceVersionFilterPattern }}'
{{- if .WithRegistry }}
      registry:
        url: '{{ .RegistryURL }}'
        rootdir: '{{ .RegistryRootDir }}'
        auth:
          token: '{{ .RegistryAuthToken }}'
          headerFormat : '{{ .RegistryHeaderFormat }}'
{{- if .RegistrySCMID }}
    scmid: '{{ .RegistrySCMID }}'
{{- end }}
{{- end }}
  {{ .ExistingSourceID }}:
    name: '{{ .ExistingSourceName }}'
    kind: 'toml'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{- end }}
    spec:
      file: '{{ .File }}'
      Key: '{{ .ExistingSourceKey }}'
conditions:
  {{ .ConditionID }}:
    name: 'Test if version of "{{ .DependencyName }}" {{"{{"}} source "{{ .ExistingSourceID }}" {{"}}"}} differs from {{"{{"}} source "{{ .SourceID }}" {{"}}"}}'
    kind: 'shell'
    sourceid: '{{ .SourceID }}'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{- end }}
    spec:
      command: 'test {{"{{"}} source "{{ .ExistingSourceID }}" {{"}}"}} != '
targets:
  {{ .TargetID }}:
    name: '{{ .TargetName }}'
    kind: 'toml'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{- end }}
    spec:
      file: '{{ .TargetFile }}'
      key: '{{ .TargetKey }}'
    sourceid: '{{ .SourceID }}'
{{- if .TargetCargoCleanupEnabled }}
  Cargo.lock:
    name: Update Cargo lockfile Cargo.lock
    dependson:
      - {{ .TargetID }}
    disablesourceinput: true
    kind: shell
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{- end }}
    spec:
      command: cargo generate-lockfile
      environments:
        - name: PATH
      changedif:
        kind: file/checksum
        spec:
          files:
            - "Cargo.lock"
      workdir: '{{ .TargetWorkdir }}'
{{- end }}
`
)
