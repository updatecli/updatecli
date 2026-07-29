package flux

const (
	// ociRepositoryManifestTemplateLatest is the Go template used to generate Flux manifests for ocirepository resources without digest
	ociRepositoryManifestTemplateLatest string = `name: 'deps(flux): bump ociRepository "{{ .OCIName }}"'
{{- if .ActionID }}
actions:
  {{ .ActionID }}:
    title: 'deps: update OCI repository to {{ "{{" }} source "oci" {{ "}}" }}'
{{- end }}
sources:
  oci:
    name: 'Get latest "{{ .OCIName }}" OCI artifact tag'
    kind: 'dockerimage'
    spec:
      image: '{{ .OCIName }}'
      tagfilter: '{{ .TagFilter }}'
      versionfilter:
        kind: '{{ .VersionFilterKind }}'
        pattern: '{{ .VersionFilterPattern }}'
        {{- if or (eq .VersionFilterKind "regex/semver") (eq .VersionFilterKind "regex/time") }}
        regex: '{{ .VersionFilterRegex }}'
        {{- end }}
      {{- if .RegistryUsername }}
      username: '{{ .RegistryUsername }}'
      {{- end }}
      {{- if .RegistryPassword }}
      password: '{{ .RegistryPassword }}'
      {{- end }}
      {{- if .RegistryToken }}
      token: '{{ .RegistryToken }}'
      {{- end }}
targets:
  oci:
    name: 'deps(flux): bump OCI repository "{{ .OCIName }}"'
    kind: 'yaml'
{{- if .ScmID }}
    scmid: {{ .ScmID }}
{{- end }}
    spec:
      file: '{{ .File }}'
      key: '$.spec.ref.tag'
      documentindex: {{ .TargetYAMLDocument }}
    sourceid: 'oci'
`
	// ociRepositoryManifestTemplateDigestAndLatest is the Go template used to generate Flux manifests for ocirepository resources with digest and latest
	ociRepositoryManifestTemplateDigestAndLatest string = `name: 'deps(flux): bump ociRepository "{{ .OCIName }}"'
{{- if .ActionID }}
actions:
  {{ .ActionID }}:
    title: 'deps: update OCI repository digest for {{ .OCIName }}:{{ "{{" }} source "oci" {{ "}}" }}'
{{- end }}
sources:
  oci:
    name: 'Get latest "{{ .OCIName }}" OCI artifact tag'
    kind: 'dockerimage'
    spec:
      image: '{{ .OCIName }}'
      tagfilter: '{{ .TagFilter }}'
      versionfilter:
        kind: '{{ .VersionFilterKind }}'
        pattern: '{{ .VersionFilterPattern }}'
        {{- if or (eq .VersionFilterKind "regex/semver") (eq .VersionFilterKind "regex/time") }}
        regex: '{{ .VersionFilterRegex }}'
        {{- end }}
      {{- if .RegistryUsername }}
      username: '{{ .RegistryUsername }}'
      {{- end }}
      {{- if .RegistryPassword }}
      password: '{{ .RegistryPassword }}'
      {{- end }}
      {{- if .RegistryToken }}
      token: '{{ .RegistryToken }}'
      {{- end }}
  oci-digest:
    name: 'Get latest "{{ .OCIName }}" OCI artifact digest'
    kind: 'dockerdigest'
    spec:
      image: '{{ .OCIName }}'
      tag: '{{ "{{" }} source "oci" {{ "}}" }}'
      hidetag: true
      {{- if .RegistryUsername }}
      username: '{{ .RegistryUsername }}'
      {{- end }}
      {{- if .RegistryPassword }}
      password: '{{ .RegistryPassword }}'
      {{- end }}
      {{- if .RegistryToken }}
      token: '{{ .RegistryToken }}'
      {{- end }}
    transformers:
      - trimprefix: '@'
    dependson:
      - 'oci'
targets:
  oci-tag:
    name: 'deps(flux): bump OCI repository "{{ .OCIName }}" tag'
    kind: 'yaml'
{{- if .ScmID }}
    scmid: {{ .ScmID }}
{{- end }}
    spec:
      file: '{{ .File }}'
      key: '$.spec.ref.tag'
      documentindex: {{ .TargetYAMLDocument }}
    sourceid: 'oci'
  oci-digest:
    name: 'deps(flux): bump OCI repository "{{ .OCIName }}" digest'
    kind: 'yaml'
{{- if .ScmID }}
    scmid: {{ .ScmID }}
{{- end }}
    spec:
      file: '{{ .File }}'
      key: '$.spec.ref.digest'
      documentindex: {{ .TargetYAMLDocument }}
    sourceid: 'oci-digest'
`
	// ociRepositoryManifestTemplateDigest is the Go template used to generate Flux manifests for ocirepository resources with digest without updating the tag.
	ociRepositoryManifestTemplateDigest string = `name: 'deps(flux): bump ociRepository "{{ .OCIName }}"'
{{- if .ActionID }}
actions:
  {{ .ActionID }}:
    title: 'deps: update OCI repository digest for {{ .OCIName }}:{{ .OCIVersion }}'
{{- end }}
sources:
  oci-digest:
    name: 'Get latest "{{ .OCIName }}" OCI artifact digest'
    kind: 'dockerdigest'
    spec:
      image: '{{ .OCIName }}'
      tag: '{{ .OCIVersion }}'
      hidetag: true
      {{- if .RegistryUsername }}
      username: '{{ .RegistryUsername }}'
      {{- end }}
      {{- if .RegistryPassword }}
      password: '{{ .RegistryPassword }}'
      {{- end }}
      {{- if .RegistryToken }}
      token: '{{ .RegistryToken }}'
      {{- end }}
    transformers:
      - trimprefix: '@'
targets:
  oci-digest:
    name: 'deps(flux): bump OCI repository "{{ .OCIName }}" digest'
    kind: 'yaml'
{{- if .ScmID }}
    scmid: {{ .ScmID }}
{{- end }}
    spec:
      file: '{{ .File }}'
      key: '$.spec.ref.digest'
      documentindex: {{ .TargetYAMLDocument }}
    sourceid: 'oci-digest'
`
)
