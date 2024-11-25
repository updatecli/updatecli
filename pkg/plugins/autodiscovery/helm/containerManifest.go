package helm

const (
	manifestTemplateLatest string = `name: 'deps(helm): bump image "{{ .ImageName }}" tag for chart "{{ .ChartName }}"'
sources:
  {{ .SourceID }}:
    name: 'get latest image tag for "{{ .ImageName }}"'
    kind: 'dockerimage'
    spec:
      image: '{{ .SourceImageName }}'
      tagfilter: '{{ .SourceTagFilter }}'
      versionfilter:
        kind: '{{ .SourceVersionFilterKind }}'
        pattern: '{{ .SourceVersionFilterPattern }}'
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
    sourceid: '{{ .SourceID }}'
`
	manifestTemplateDigestAndLatest string = `name: 'deps(helm): bump image "{{ .ImageName }}" digest for chart "{{ .ChartName }}"'
sources:
  {{ .SourceID }}:
    name: 'get latest "{{ .ImageName }}" container tag'
    kind: 'dockerimage'
    spec:
      image: '{{ .SourceImageName }}'
      tagfilter: '{{ .SourceTagFilter }}'
      versionfilter:
        kind: '{{ .SourceVersionFilterKind }}'
        pattern: '{{ .SourceVersionFilterPattern }}'
  {{ .SourceID }}-digest:
    name: 'get latest image "{{ .ImageName }}" digest'
    kind: 'dockerdigest'
    spec:
      image: '{{ .ImageName }}'
      tag: '{{ "{{" }} source "{{ .SourceID }}" {{ "}}" }}'
    dependson:
      - '{{ .SourceID }}'
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
    sourceid: '{{ .SourceID }}-digest'
`
	manifestTemplateDigest string = `name: 'deps(helm): bump image "{{ .ImageName }}" digest for chart "{{ .ChartName }}"'
sources:
  {{ .SourceID }}-digest:
    name: 'get latest image "{{ .ImageName }}" digest'
    kind: 'dockerdigest'
    spec:
      image: '{{ .ImageName }}'
      tag: '{{ .ImageTag }}'
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
    sourceid: '{{ .SourceID }}-digest'
`
)
