package helm

const (
	// manifestTemplateMirrorCondition checks that the mirror serves the version
	// the target is about to write, when the image is looked up upstream.
	// It is inserted at the end of the "conditions" section.
	manifestTemplateMirrorCondition string = `{{- if .MirrorCondition }}
  {{ .TargetID }}-mirror:
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
{{- end }}`
	manifestTemplateLatest string = `name: 'deps(helm): bump image "{{ .ImageName }}" tag for chart "{{ .ChartName }}"'
{{- if .ActionID }}
actions:
  {{ .ActionID }}:
    title: 'deps(helm): update Docker image {{ .ImageName }} to {{ "{{" }} source "image" {{ "}}" }}'
{{ end }}
sources:
  image:
    name: 'get latest image tag for "{{ .SourceImageName }}"'
    kind: 'dockerimage'
    spec:
      image: '{{ .SourceImageName }}'
      tagfilter: '{{ .SourceTagFilter }}'
      versionfilter:
        kind: '{{ .SourceVersionFilterKind }}'
        pattern: '{{ .SourceVersionFilterPattern }}'
{{- if or (eq .SourceVersionFilterKind "regex/semver") (eq .SourceVersionFilterKind "regex/time") }}
        regex: '{{ .SourceVersionFilterRegex }}'
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
conditions:
{{- if .HasRegistry }}
  {{ .ConditionRegistryID }}:
    disablesourceinput: true
    name: '{{ .ConditionRegistryName }}'
    kind: 'yaml'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .File }}'
      key: '{{ .ConditionRegistryKey }}'
      value: '{{ .ConditionRegistryValue }}'
{{- end }}
  {{ .ConditionRepositoryID }}:
    disablesourceinput: true
    name: '{{ .ConditionRepositoryName }}'
    kind: 'yaml'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .File }}'
      key: '{{ .ConditionRepositoryKey }}'
      value: '{{ .ConditionRepositoryValue }}'
` + manifestTemplateMirrorCondition + `
targets:
  {{ .TargetID }}:
    name: 'deps(helm): bump image "{{ .ImageName }}" tag'
    kind: 'helmchart'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .TargetFile }}'
      name: '{{ .TargetChartName }}'
      key: '{{ .TargetKey }}'
      skippackaging: {{ .TargetChartSkipPackaging }}
      versionincrement: '{{ .TargetChartVersionIncrement }}'
    sourceid: 'image'
`
	manifestTemplateDigestAndLatest string = `name: 'deps(helm): bump image "{{ .ImageName }}" digest for chart "{{ .ChartName }}"'
{{- if .ActionID }}
actions:
  {{ .ActionID }}:
    title: 'deps(helm): update Docker image {{ .ImageName }} to {{ "{{" }} source "image" {{ "}}" }}'
{{ end }}
sources:
  image:
    name: 'get latest "{{ .SourceImageName }}" container tag'
    kind: 'dockerimage'
    spec:
      image: '{{ .SourceImageName }}'
      tagfilter: '{{ .SourceTagFilter }}'
      versionfilter:
        kind: '{{ .SourceVersionFilterKind }}'
        pattern: '{{ .SourceVersionFilterPattern }}'
        {{- if or (eq .SourceVersionFilterKind "regex/semver") (eq .SourceVersionFilterKind "regex/time") }}
        regex: '{{ .SourceVersionFilterRegex }}'
        {{- end }}
      {{- if .RegistryToken }}
      token: '{{ .RegistryToken }}'
      {{- end }}
      {{- if .RegistryUsername }}
      username: '{{ .RegistryUsername }}'
      {{- end }}
      {{- if .RegistryPassword }}
      password: '{{ .RegistryPassword }}'
      {{- end }}
  image-digest:
    name: 'get latest image "{{ .SourceImageName }}" digest'
    kind: 'dockerdigest'
    spec:
      image: '{{ .SourceImageName }}'
      tag: '{{ "{{" }} source "image" {{ "}}" }}'
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
      - 'image'
conditions:
{{- if .HasRegistry }}
  {{ .ConditionRegistryID }}:
    disablesourceinput: true
    name: '{{ .ConditionRegistryName }}'
    kind: 'yaml'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .File }}'
      key: '{{ .ConditionRegistryKey }}'
      value: '{{ .ConditionRegistryValue }}'
{{- end }}
  {{ .ConditionRepositoryID }}:
    disablesourceinput: true
    name: '{{ .ConditionRepositoryName }}'
    kind: 'yaml'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .File }}'
      key: '{{ .ConditionRepositoryKey }}'
      value: '{{ .ConditionRepositoryValue }}'
` + manifestTemplateMirrorCondition + `
targets:
  {{ .TargetID }}:
    name: 'deps(helm): bump image "{{ .ImageName }}" digest'
    kind: 'helmchart'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .TargetFile }}'
      name: '{{ .TargetChartName }}'
      key: '{{ .TargetKey }}'
      skippackaging: {{ .TargetChartSkipPackaging }}
      versionincrement: '{{ .TargetChartVersionIncrement }}'
    sourceid: 'image-digest'
`
	manifestTemplateDigest string = `name: 'deps(helm): bump image "{{ .ImageName }}" digest for chart "{{ .ChartName }}"'
{{- if .ActionID }}
actions:
  {{ .ActionID }}:
    title: 'deps(helm): update Docker image {{ .ImageName }} digest'
{{ end }}
sources:
  image-digest:
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
conditions:
{{- if .HasRegistry }}
  {{ .ConditionRegistryID }}:
    disablesourceinput: true
    name: '{{ .ConditionRegistryName }}'
    kind: 'yaml'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .File }}'
      key: '{{ .ConditionRegistryKey }}'
      value: '{{ .ConditionRegistryValue }}'
{{- end }}
  {{ .ConditionRepositoryID }}:
    disablesourceinput: true
    name: '{{ .ConditionRepositoryName }}'
    kind: 'yaml'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .File }}'
      key: '{{ .ConditionRepositoryKey }}'
      value: '{{ .ConditionRepositoryValue }}'
` + manifestTemplateMirrorCondition + `
targets:
  {{ .TargetID }}:
    name: 'deps(helm): bump image "{{ .ImageName }}" digest'
    kind: 'helmchart'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .TargetFile }}'
      name: '{{ .TargetChartName }}'
      key: '{{ .TargetKey }}'
      skippackaging: {{ .TargetChartSkipPackaging }}
      versionincrement: '{{ .TargetChartVersionIncrement }}'
    sourceid: 'image-digest'
`
)
