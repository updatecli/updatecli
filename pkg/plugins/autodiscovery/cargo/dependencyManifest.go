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
  {{ .ExistingSourceID }}:
    name: '{{ .ExistingSourceName }}'
    kind: 'toml'
    spec:
      file: '{{ .File }}'
      Key: '{{ .ExistingSourceKey }}'
conditions:
  {{ .ConditionID }}:
    name: 'Ensure Cargo chart named "{{ .DependencyName }}" is specified'
    kind: 'toml'
    spec:
      file: '{{ .File }}'
      Query: '{{ .ConditionQuery }}'
    sourceid: '{{ .ExistingSourceID }}'
targets:
  {{ .TargetID }}:
    name: 'Bump crate dependency "{{ .DependencyName }}" for crate "{{ .CrateName }}"'
    kind: 'toml'
    spec:
      file: '{{ .TargetFile }}'
      key: '{{ .TargetKey }}'
    sourceid: '{{ .SourceID }}'
`
)
