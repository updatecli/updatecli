package ko

const (
	// manifestTemplateMirrorCondition checks that the mirror serves the version
	// the target is about to write, when the image is looked up upstream.
	manifestTemplateMirrorCondition string = `{{- if .MirrorCondition }}
conditions:
  {{ .SourceID }}-mirror:
    name: 'check container image "{{ .ImageName }}:{{ "{{" }} source "{{ .MirrorConditionSourceID }}" {{ "}}" }}" is available on the mirror'
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
	// manifestTemplateLatest is the Go template used to generate Kubernetes manifests
	manifestTemplateLatest string = `name: '{{ .ManifestName }}'
{{- if .ActionID }}
actions:
  {{ .ActionID }}:
    title: '{{ .TargetName }}'
{{ end }}
sources:
  {{ .SourceID }}:
    name: 'get latest container image tag for "{{ .SourceImageName }}"'
    kind: 'dockerimage'
    spec:
      image: '{{ .SourceImageName }}'
      tagfilter: '{{ .SourceTagFilter }}'
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
` + manifestTemplateMirrorCondition + `targets:
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
    name: 'get latest container image tag for "{{ .SourceImageName }}"'
    kind: 'dockerimage'
    spec:
      image: '{{ .SourceImageName }}'
      tagfilter: '{{ .SourceTagFilter }}'
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
  {{ .SourceID }}-digest:
    name: 'get latest container image digest for "{{ .SourceImageName }}:{{ .ImageTag }}"'
    kind: 'dockerdigest'
    spec:
      image: '{{ .SourceImageName }}'
      tag: '{{ "{{" }} source "{{ .SourceID }}" {{ "}}" }}'
      {{- if .RegistryUsername }}
      username: '{{ .RegistryUsername }}'
      {{- end }}
      {{- if .RegistryPassword }}
      password: '{{ .RegistryPassword }}'
      {{- end }}
      {{- if .RegistryToken }}
      token: '{{ .RegistryToken }}'
      {{- end }}
    dependson:
      - '{{ .SourceID }}'
` + manifestTemplateMirrorCondition + `targets:
  {{ .TargetID }}:
    name: 'deps: bump container image digest for "{{ .ImageName }}:{{ .ImageTag}}"'
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
    title: 'deps: bump container image digest for "{{ .ImageName }}:{{ .ImageTag}}"'
{{ end }}
sources:
  {{ .SourceID }}-digest:
    name: 'get latest container image digest for "{{ .SourceImageName }}:{{ .ImageTag }}"'
    kind: 'dockerdigest'
    spec:
      image: '{{ .SourceImageName }}'
      tag: '{{ .ImageTag }}'
      {{- if .RegistryUsername }}
      username: '{{ .RegistryUsername }}'
      {{- end }}
      {{- if .RegistryPassword }}
      password: '{{ .RegistryPassword }}'
      {{- end }}
      {{- if .RegistryToken }}
      token: '{{ .RegistryToken }}'
      {{- end }}
` + manifestTemplateMirrorCondition + `targets:
  {{ .TargetID }}:
    name: 'deps: bump container image digest for "{{ .ImageName }}:{{ .ImageTag}}"'
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
)
