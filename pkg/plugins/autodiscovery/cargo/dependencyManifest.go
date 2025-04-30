package cargo

const (
	// dependencyManifest is the Go template used to generate
	// the manifests to update the cargo file
	dependencyManifest string = `name: '{{ .ManifestName }}'
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
      file: '{{ .CargoFile }}'
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
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{- end }}
{{- if .CargoUpgradeAvailable }}
    kind: 'shell'
    spec:
      command: |
        ARGS=""
        if [ "$DRY_RUN" = "true" ]; then
          ARGS="--dry-run"
        fi
        cargo upgrade $ARGS --manifest-path {{ .CargoFile }} --package {{ .DependencyName }}@{{"{{"}} source "{{ .SourceID }}" {{"}}"}}
{{- if .CargoLockFile }}
        cargo update $ARGS --manifest-path {{ .CargoFile }} {{ .DependencyName }}@{{"{{"}} source "{{ .SourceID }}" {{"}}"}} --precise {{"{{"}} source "{{ .SourceID }}" {{"}}"}}
{{- end }}
      changedif:
        kind: file/checksum
        spec:
          files:
            - "{{ .CargoFile }}"
{{- if .CargoLockFile}}
            - "{{ .CargoLockFile }}"
{{- end }}
    disablesourceinput: true
{{- else }}
    kind: 'toml'
    spec:
      file: '{{ .CargoFile }}'
      key: '{{ .TargetKey }}'
    sourceid: '{{ .SourceID }}'
{{- if .CargoLockFile }}
  lockfile:
    name: '{{ .CargoLockTargetName }}'
    kind: 'shell'
    dependson:
      - target#{{ .TargetID }}
    spec:
      command: |
        ARGS=""
        if [ "$DRY_RUN" = "true" ]; then
          ARGS="--dry-run"
        fi
        cargo update $ARGS --manifest-path {{ .CargoFile }} {{ .DependencyName }}@{{"{{"}} source "{{ .SourceID }}" {{"}}"}} --precise {{"{{"}} source "{{ .SourceID }}" {{"}}"}}
      changedif:
        kind: file/checksum
        spec:
          files:
            - "{{ .CargoLockFile }}"
    disablesourceinput: true
{{- end }}
{{- end }}
`
)
