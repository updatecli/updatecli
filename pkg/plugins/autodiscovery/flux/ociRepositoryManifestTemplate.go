package flux

const (
	// ociRepositoryManifestTemplateLatest is the Go template used to generate Flux manifests for ocirepository resources
	ociRepositoryManifestTemplateLatest string = `name: 'deps(flux): bump ociRepository "{{ .OCIName }}"'
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
targets:
  oci:
    name: 'deps(flux): bump OCI repository "{{ .OCIName }}"'
    kind: 'yaml'
{{- if .ScmID }}
    scmid: {{ .ScmID }}
{{ end }}
    spec:
      file: '{{ .File }}'
      key: '$.spec.ref.tag'
    sourceid: 'oci'
`
	ociRepositoryManifestTemplateDigestAndLatest string = `name: 'deps(flux): bump ociRepository "{{ .OCIName }}"'
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
  oci-digest:
    name: 'Get latest "{{ .OCIName }}" OCI artifact digest'
    kind: 'dockerdigest'
    spec:
      image: '{{ .OCIName }}'
      tag: '{{ "{{" }} source "oci" {{ "}}" }}'
    dependson:
      - 'oci'
targets:
  oci:
    name: 'deps(flux): bump OCI repository "{{ .OCIName }}"'
    kind: 'yaml'
{{- if .ScmID }}
    scmid: {{ .ScmID }}
{{ end }}
    spec:
      file: '{{ .File }}'
      key: '$.spec.ref.tag'
    sourceid: 'oci-digest'
`

	ociRepositoryManifestTemplateDigest string = `name: 'deps(flux): bump ociRepository "{{ .OCIName }}"'
sources:
  oci-digest:
    name: 'Get latest "{{ .OCIName }}" OCI artifact digest'
    kind: 'dockerdigest'
    spec:
      image: '{{ .OCIName }}'
      tag: '{{ .OCIVersion }}'
targets:
  oci:
    name: 'deps(flux): bump OCI repository "{{ .OCIName }}"'
    kind: 'yaml'
{{- if .ScmID }}
    scmid: {{ .ScmID }}
{{ end }}
    spec:
      file: '{{ .File }}'
      key: '$.spec.ref.tag'
    sourceid: 'oci-digest'
`
)
