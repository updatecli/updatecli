package dockercompose

const (
	// manifestTemplate is the Go template used to generate Docker compose manifests
	manifestTemplateLatest string = `name: 'deps(dockercompose): bump "{{ .ImageName }}" tag'
sources:
  {{ .SourceID }}:
    name: 'get latest image tag for "{{ .ImageName }}"'
    kind: 'dockerimage'
    spec:
{{- if .ImageArchitecture }}
      architecture: '{{ .ImageArchitecture }}'
{{- end }}
      image: '{{ .ImageName }}'
      tagfilter: '{{ .TagFilter }}'
      versionfilter:
        kind: '{{ .VersionFilterKind }}'
        pattern: '{{ .VersionFilterPattern }}'
targets:
  {{ .TargetID }}:
    name: 'deps(dockercompose): bump "{{ .ImageName }}" tag'
    kind: 'yaml'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .TargetFile }}'
      key: '{{ .TargetKey }}'
    sourceid: '{{ .SourceID }}'
    transformers:
      - addprefix: '{{ .TargetPrefix }}'
`
	manifestTemplateDigestAndLatest string = `name: 'deps(dockercompose): bump "{{ .ImageName }}" digest'
sources:
  {{ .SourceID }}:
    name: 'get latest image tag for "{{ .ImageName }}"'
    kind: 'dockerimage'
    spec:
{{- if .ImageArchitecture }}
      architecture: '{{ .ImageArchitecture }}'
{{- end }}
      image: '{{ .ImageName }}'
      tagfilter: '{{ .TagFilter }}'
      versionfilter:
        kind: '{{ .VersionFilterKind }}'
        pattern: '{{ .VersionFilterPattern }}'
  {{ .SourceID }}-digest:
    name: 'get latest image "{{ .ImageName }}" digest'
    kind: 'dockerdigest'
    spec:
      image: '{{ .ImageName }}'
      tag: '{{ "{{" }} source "{{ .SourceID }}" {{ "}}" }}'
    dependson:
      - '{{ .SourceID }}'
targets:
  {{ .TargetID }}:
    name: 'deps(dockercompose): bump "{{ .ImageName }}" digest'
    kind: 'yaml'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .TargetFile }}'
      key: '{{ .TargetKey }}'
    sourceid: '{{ .SourceID }}-digest'
    transformers:
      - addprefix: '{{ .TargetPrefix }}'
`
	manifestTemplateDigest string = `name: 'deps(dockercompose): bump image "{{ .ImageName }}" digest'
sources:
  {{ .SourceID }}-digest:
    name: 'get latest image "{{ .ImageName }}" digest'
    kind: 'dockerdigest'
    spec:
      image: '{{ .ImageName }}'
      tag: '{{ .ImageTag }}'
targets:
  {{ .TargetID }}:
    name: 'deps(dockercompose): bump image "{{ .ImageName }}" digest'
    kind: 'yaml'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .TargetFile }}'
      key: '{{ .TargetKey }}'
    sourceid: '{{ .SourceID }}-digest'
    transformers:
      - addprefix: '{{ .TargetPrefix }}'
`
)
