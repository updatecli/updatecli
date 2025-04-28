package nomad

const (
	// manifestTemplate is the Go template used to generate Docker compose manifests
	manifestTemplateLatest string = `name: 'deps(dockercompose): bump "{{ .ImageName }}" tag'
{{- if .ActionID }}
actions:
  {{ .ActionID }}:
    title: 'deps: update Nomad job task "{{ .ImageName }}" to "{{ "{{" }} source "{{ .SourceID }}" {{ "}}" }}"'
{{- end }}
sources:
  {{ .SourceID }}:
    name: 'get latest image tag for "{{ .ImageName }}"'
    kind: 'dockerimage'
    spec:
      image: '{{ .ImageName }}'
      tagfilter: '{{ .TagFilter }}'
      versionfilter:
        kind: '{{ .VersionFilterKind }}'
        pattern: '{{ .VersionFilterPattern }}'
targets:
  {{ .TargetID }}:
    name: 'deps(nomad): update Docker image "{{ .ImageName }}" to "{{ "{{" }} source "{{ .SourceID }}" {{ "}}" }}"'
    kind: 'hcl'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{- end }}
    spec:
      file: '{{ .TargetFile }}'
      path: '{{ .TargetPath }}'
    sourceid: '{{ .SourceID }}'
{{- if .TargetPrefix }}
    transformers:
      - addprefix: '{{ .TargetPrefix }}'
{{- end }}
`
	manifestTemplateDigestAndLatest string = `name: 'deps(nomad): update Docker image digest "{{ .ImageName }}"'
{{- if .ActionID }}
actions:
  {{ .ActionID }}:
    title: 'deps(nomad): update Docker image "{{ .ImageName }}" to "{{ "{{" }} source "{{ .SourceID }}" {{ "}}" }}"'
{{- end }}
sources:
  {{ .SourceID }}:
    name: 'get latest image tag for "{{ .ImageName }}"'
    kind: 'dockerimage'
    spec:
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
    name: 'deps(nomad): update Docker image "{{ .ImageName }}" to "{{ "{{" }} source "{{ .SourceID }}" {{ "}}" }}"'
    kind: 'hcl'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{- end }}
    spec:
      file: '{{ .TargetFile }}'
      path: '{{ .TargetPath }}'
    sourceid: '{{ .SourceID }}-digest'
{{- if .TargetPrefix }}
    transformers:
      - addprefix: '{{ .TargetPrefix }}'
{{- end }}
`
	manifestTemplateDigest string = `name: 'deps(nomad): update Docker image digest for "{{ .ImageName }}" digest'
{{- if .ActionID }}
actions:
  {{ .ActionID }}:
    title: 'deps(nomad): update Docker image digest for "{{ .ImageName }}:{{ .ImageTag }}"'
{{- end }}
sources:
  {{ .SourceID }}-digest:
    name: 'get latest image "{{ .ImageName }}" digest'
    kind: 'dockerdigest'
    spec:
      image: '{{ .ImageName }}'
      tag: '{{ .ImageTag }}'
targets:
  {{ .TargetID }}:
    name: 'deps(nomad): update Docker image digest for "{{ .ImageName }}:{{ .ImageTag }}"'
    kind: 'hcl'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{- end }}
    spec:
      file: '{{ .TargetFile }}'
      path: '{{ .TargetPath }}'
    sourceid: '{{ .SourceID }}-digest'
{{- if .TargetPrefix }}
    transformers:
      - addprefix: '{{ .TargetPrefix }}'
{{- end }}
`
)
