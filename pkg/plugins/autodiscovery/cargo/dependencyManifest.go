package cargo

const (
	// dependencyManifest is the Go template used to generate
	// the Helm chart manifests specific for Helm dependencies
	dependencyManifest string = `name: '{{ .ManifestName }}'
sources:
  {{ .SourceID }}:
    name: '{{ .SourceName }}'
    kind: 'cargopackage'
    spec:
      package: '{{ .DependencyName }}'
      versionFilter:
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
    name: 'Ensure Cargo chart named "{{ .DependencyName }}" is specified'
    kind: 'toml'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{- end }}
    spec:
      file: '{{ .File }}'
      Query: '{{ .ConditionQuery }}'
    sourceid: '{{ .ExistingSourceID }}'
targets:
  {{ .TargetID }}:
    name: 'Bump crate dependency "{{ .DependencyName }}" for crate "{{ .CrateName }}"'
    kind: 'toml'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{- end }}
    spec:
      file: '{{ .TargetFile }}'
      key: '{{ .TargetKey }}'
    sourceid: '{{ .SourceID }}'
`
)
