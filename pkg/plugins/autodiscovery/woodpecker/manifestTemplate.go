package woodpecker

const (
	// manifestTemplateMirrorCondition checks that the mirror serves the version
	// the target is about to write, when the image is looked up upstream.
	manifestTemplateMirrorCondition string = `{{- if .MirrorCondition }}
conditions:
  {{ .SourceID }}-mirror:
    name: 'check image "{{ .ImageName }}:{{ "{{" }} source "{{ .MirrorConditionSourceID }}" {{ "}}" }}" is available on the mirror'
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
	// manifestTemplateLatest is the Go template used to generate Woodpecker manifests for tag updates
	manifestTemplateLatest string = `name: 'deps(woodpecker): bump "{{ .ImageName }}" tag'
{{- if .ActionID }}
actions:
  {{ .ActionID }}:
    title: 'deps: update Woodpecker image "{{ .ImageName }}" to "{{ "{{" }} source "{{ .SourceID }}" {{ "}}" }}"'
{{ end }}
sources:
  {{ .SourceID }}:
    name: 'get latest image tag for "{{ .SourceImageName }}"'
    kind: 'dockerimage'
    spec:
      image: '{{ .SourceImageName }}'
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
` + manifestTemplateMirrorCondition + `targets:
  {{ .TargetID }}:
    name: 'deps: update Woodpecker image "{{ .ImageName }}" to "{{ "{{" }} source "{{ .SourceID }}" {{ "}}" }}"'
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
	// manifestTemplateDigestAndLatest is the Go template used to generate Woodpecker manifests for digest+tag updates
	manifestTemplateDigestAndLatest string = `name: 'deps(woodpecker): bump "{{ .ImageName }}" digest'
{{- if .ActionID }}
actions:
  {{ .ActionID }}:
    title: 'deps: update Woodpecker image "{{ .ImageName }}" to "{{ "{{" }} source "{{ .SourceID }}" {{ "}}" }}"'
{{ end }}
sources:
  {{ .SourceID }}:
    name: 'get latest image tag for "{{ .SourceImageName }}"'
    kind: 'dockerimage'
    spec:
      image: '{{ .SourceImageName }}'
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
  {{ .SourceID }}-digest:
    name: 'get latest image "{{ .SourceImageName }}" digest'
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
    name: 'deps: update Woodpecker image "{{ .ImageName }}" to "{{ "{{" }} source "{{ .SourceID }}" {{ "}}" }}"'
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
	// manifestTemplateDigest is the Go template used to generate Woodpecker manifests for digest-only updates
	manifestTemplateDigest string = `name: 'deps(woodpecker): bump image "{{ .ImageName }}" digest'
{{- if .ActionID }}
actions:
  {{ .ActionID }}:
    title: 'deps: update Woodpecker image "{{ .ImageName }}:{{ .ImageTag }}" digest'
{{ end }}
sources:
  {{ .SourceID }}-digest:
    name: 'get latest image "{{ .SourceImageName }}" digest'
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
    name: 'deps: bump Woodpecker image "{{ .ImageName }}:{{ .ImageTag }}" digest'
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
