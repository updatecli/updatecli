package githubaction

const (
	// manifestTemplate is the Go template used to generate Docker compose manifests
	manifestTemplateDockerLatest string = `name: 'deps: bump Docker image "{{ .ImageName }}"'
sources:
  image:
    name: 'get latest image tag for "{{ .ImageName }}"'
    kind: 'dockerimage'
    spec:
      image: '{{ .ImageName }}'
      tagfilter: '{{ .TagFilter }}'
      versionfilter:
        kind: '{{ .VersionFilterKind }}'
        pattern: '{{ .VersionFilterPattern }}'
targets:
  workflow:
    name: 'deps: bump Docker image "{{ .ImageName }}" to {{ "{{" }} source "image" {{ "}}" }}'
    kind: 'yaml'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .TargetFile }}'
      key: '{{ .TargetKey }}'
    sourceid: 'image'
    transformers:
      - addprefix: '{{ .TargetPrefix }}'
`
	manifestTemplateDockerDigestAndLatest string = `name: 'deps: bump Docker image digest for "{{ .ImageName }}"'
sources:
  image:
    name: 'get latest image tag for "{{ .ImageName }}"'
    kind: 'dockerimage'
    spec:
      image: '{{ .ImageName }}'
      tagfilter: '{{ .TagFilter }}'
      versionfilter:
        kind: '{{ .VersionFilterKind }}'
        pattern: '{{ .VersionFilterPattern }}'

  digest:
    name: 'get image "{{ .ImageName }}" digest for tag {{ "{{" }} source "image" {{ "}}" }}'
    kind: 'dockerdigest'
    spec:
      image: '{{ .ImageName }}'
      tag: '{{ "{{" }} source "image" {{ "}}" }}'
    dependson:
      - 'image'
targets:
  workflow:
    name: 'deps: bump Docker image digest for "{{ .ImageName }}" to {{ "{{" }} source "digest" {{ "}}" }}'
    kind: 'yaml'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .TargetFile }}'
      key: '{{ .TargetKey }}'
    sourceid: 'digest'
    transformers:
      - addprefix: '{{ .TargetPrefix }}'
`
	manifestTemplateDockerDigest string = `name: 'deps: bump Docker image digest "{{ .ImageName }}"'
sources:
  digest:
    name: 'get latest image "{{ .ImageName }}" digest'
    kind: 'dockerdigest'
    spec:
      image: '{{ .ImageName }}'
      tag: '{{ .ImageTag }}'
targets:
  workflow:
    name: 'deps: bump Docker image Docker image "{{ .ImageName }}" digest to {{ "{{" }} source "digest" {{ "}}" }}'
    kind: 'yaml'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .TargetFile }}'
      key: '{{ .TargetKey }}'
    sourceid: 'digest'
    transformers:
      - addprefix: '{{ .TargetPrefix }}'
`
)
