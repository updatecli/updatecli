package kubernetes

const (
	// manifestTemplateLatest is the Go template used to generate Kubernetes manifests
	manifestTemplateLatest string = `name: '{{ .ManifestName }}'
{{- if .ActionID }}
actions:
  {{ .ActionID }}:
    title: 'deps: bump container image "{{ .ImageName }}" to {{ "{{" }} source "{{ .SourceID }}" {{ "}}" }}'
{{ end }}
sources:
  {{ .SourceID }}:
    name: 'get latest container image tag for "{{ .ImageName }}"'
    kind: 'dockerimage'
    spec:
      image: '{{ .ImageName }}'
      tagfilter: '{{ .SourceTagFilter }}'
      versionfilter:
        kind: '{{ .VersionFilterKind }}'
        pattern: '{{ .VersionFilterPattern }}'
{{- if or (eq .VersionFilterKind "regex/semver") (eq .VersionFilterKind "regex/time") }}
        regex: '{{ .VersionFilterRegex }}'
{{- end }}
targets:
  {{ .TargetID }}:
    name: 'deps: bump container image "{{ .ImageName }}" to {{ "{{" }} source "{{ .SourceID }}" {{ "}}" }}'
    kind: 'yaml'
{{- if .ScmID }}
    scmid: {{ .ScmID }}
{{ end }}
    spec:
      file: '{{ .TargetFile }}'
      key: "{{ .TargetKey}}"
    sourceid: '{{ .SourceID }}'
    transformers:
      - addprefix: '{{ .TargetPrefix }}'
`
	manifestTemplateDigestAndLatest string = `name: '{{ .ManifestName }}'
{{- if .ActionID }}
actions:
  {{ .ActionID }}:
    title: 'deps: bump container image digest for "{{ .ImageName }}:{{ "{{" }} source "{{ .SourceID }}" {{ "}}" }}"'
{{ end }}
sources:
  {{ .SourceID }}:
    name: 'get latest container image tag for "{{ .ImageName }}"'
    kind: 'dockerimage'
    spec:
      image: '{{ .ImageName }}'
      tagfilter: '{{ .SourceTagFilter }}'
      versionfilter:
        kind: '{{ .VersionFilterKind }}'
        pattern: '{{ .VersionFilterPattern }}'
{{- if or (eq .VersionFilterKind "regex/semver") (eq .VersionFilterKind "regex/time") }}
        regex: '{{ .VersionFilterRegex }}'
{{- end }}
  {{ .SourceID }}-digest:
    name: 'get latest container image digest for "{{ .ImageName }}:{{ .ImageTag }}"'
    kind: 'dockerdigest'
    spec:
      image: '{{ .ImageName }}'
      tag: '{{ "{{" }} source "{{ .SourceID }}" {{ "}}" }}'
    dependson:
      - '{{ .SourceID }}'
targets:
  {{ .TargetID }}:
    name: 'deps: bump container image digest for "{{ .ImageName }}:{{ "{{" }} source "{{ .SourceID }}" {{ "}}" }}"'
    kind: 'yaml'
{{- if .ScmID }}
    scmid: {{ .ScmID }}
{{ end }}
    spec:
      file: '{{ .TargetFile }}'
      key: "{{ .TargetKey}}"
    sourceid: '{{ .SourceID }}-digest'
    transformers:
      - addprefix: '{{ .TargetPrefix }}'
`
	manifestTemplateDigest string = `name: '{{ .ManifestName }}'
{{- if .ActionID }}
actions:
  {{ .ActionID }}:
    title: 'deps: bump container image digest for "{{ .ImageName }}:{{ .ImageTag }}"'
{{ end }}
sources:
  {{ .SourceID }}-digest:
    name: 'get latest container image digest for "{{ .ImageName }}:{{ .ImageTag }}"'
    kind: 'dockerdigest'
    spec:
      image: '{{ .ImageName }}'
      tag: '{{ .ImageTag }}'
targets:
  {{ .TargetID }}:
    name: 'deps: bump container image digest for "{{ .ImageName }}:{{ .ImageTag }}"'
    kind: 'yaml'
{{- if .ScmID }}
    scmid: {{ .ScmID }}
{{ end }}
    spec:
      file: '{{ .TargetFile }}'
      key: '{{ .TargetKey}}'
    sourceid: '{{ .SourceID }}-digest'
    transformers:
      - addprefix: '{{ .TargetPrefix }}'
`
)
