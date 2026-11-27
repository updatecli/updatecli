package githubaction

const (
	// manifestTemplateDockerMirrorCondition checks that the mirror serves the version
	// the target is about to write, when the image is looked up upstream.
	manifestTemplateDockerMirrorCondition string = `{{- if .MirrorCondition }}
conditions:
  mirror:
    name: 'check Docker image "{{ .ImageName }}:{{ "{{" }} source "{{ .MirrorConditionSourceID }}" {{ "}}" }}" is available on the mirror'
    kind: 'dockerimage'
    sourceid: '{{ .MirrorConditionSourceID }}'
    spec:
      image: '{{ .ImageName }}'
      {{- if .MirrorUsername }}
      username: '{{ .MirrorUsername }}'
      {{- end }}
      {{- if .MirrorPassword }}
      password: '{{ .MirrorPassword }}'
      {{- end }}
      {{- if .MirrorToken }}
      token: '{{ .MirrorToken }}'
      {{- end }}
{{- end }}
`
	// manifestTemplate is the Go template used to generate Docker compose manifests
	manifestTemplateDockerLatest string = `name: 'deps: bump Docker image "{{ .ImageName }}"'
{{- if .ActionID }}
actions:
  {{ .ActionID }}:
    title: 'deps: update Docker image to {{ "{{" }} source "image" {{ "}}" }}'
{{ end }}

sources:
  image:
    name: 'get latest image tag for "{{ .SourceImageName }}"'
    kind: 'dockerimage'
    spec:
      image: '{{ .SourceImageName }}'
      tagfilter: '{{ .TagFilter }}'
      versionfilter:
        kind: '{{ .VersionFilterKind }}'
        pattern: '{{ .VersionFilterPattern }}'
` + manifestTemplateDockerMirrorCondition + `targets:
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
{{- if .ActionID }}
actions:
  {{ .ActionID }}:
    title: 'deps: update Docker image to {{ "{{" }} source "image" {{ "}}" }}'
{{ end }}

sources:
  image:
    name: 'get latest image tag for "{{ .SourceImageName }}"'
    kind: 'dockerimage'
    spec:
      image: '{{ .SourceImageName }}'
      tagfilter: '{{ .TagFilter }}'
      versionfilter:
        kind: '{{ .VersionFilterKind }}'
        pattern: '{{ .VersionFilterPattern }}'

  digest:
    name: 'get image "{{ .SourceImageName }}" digest for tag {{ "{{" }} source "image" {{ "}}" }}'
    kind: 'dockerdigest'
    spec:
      image: '{{ .SourceImageName }}'
      tag: '{{ "{{" }} source "image" {{ "}}" }}'
    dependson:
      - 'image'
` + manifestTemplateDockerMirrorCondition + `targets:
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
{{- if .ActionID }}
actions:
  {{ .ActionID }}:
    title: 'deps: update Docker image to {{ "{{" }} source "digest" {{ "}}" }}'
{{ end }}

sources:
  digest:
    name: 'get latest image "{{ .SourceImageName }}" digest'
    kind: 'dockerdigest'
    spec:
      image: '{{ .SourceImageName }}'
      tag: '{{ .ImageTag }}'
` + manifestTemplateDockerMirrorCondition + `targets:
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
