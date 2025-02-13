package dockerfile

var (
	// manifestTemplateLatest is the Go template used to generate
	// Updatecli manifests
	manifestTemplateLatest string = `name: 'deps(dockerfile): bump "{{ .ImageName }}" tag'
{{- if .ActionID }}
actions:
  {{ .ActionID }}:
    title: 'deps: update Docker image "{{ .ImageName }}" to "{{ "{{" }} source "{{ .SourceID }}" {{ "}}" }}"'
{{ end }}
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
    name: 'deps: update Docker image "{{ .ImageName }}" to "{{ "{{" }} source "{{ .SourceID }}" {{ "}}" }}"'
    kind: 'dockerfile'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .TargetFile }}'
      instruction:
        keyword: '{{ .TargetKeyword }}'
        matcher: '{{ .TargetMatcher }}'
    sourceid: '{{ .SourceID }}'
`
	manifestTemplateDigestAndLatest string = `name: 'deps(dockerfile): bump "{{ .ImageName }}" digest'
{{- if .ActionID }}
actions:
  {{ .ActionID }}:
    title: 'deps: update Docker image "{{ .ImageName }}" to "{{ "{{" }} source "{{ .SourceID }}" {{ "}}" }}"'
{{ end }}
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
    name: 'deps: update Docker image "{{ .ImageName }}" to "{{ "{{" }} source "{{ .SourceID }}" {{ "}}" }}"'
    kind: 'dockerfile'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .TargetFile }}'
      instruction:
        keyword: '{{ .TargetKeyword }}'
        matcher: '{{ .TargetMatcher }}'
    sourceid: '{{ .SourceID }}-digest'
`
	manifestTemplateDigest string = `name: 'deps(dockerfile): bump image "{{ .ImageName }}" digest'
{{- if .ActionID }}
actions:
  {{ .ActionID }}:
    title: 'deps: update Docker image "{{ .ImageName }}" to "{{ "{{" }} source "{{ .SourceID }}" {{ "}}" }}"'
{{ end }}
sources:
  {{ .SourceID }}-digest:
    name: 'get latest image "{{ .ImageName }}" digest'
    kind: 'dockerdigest'
    spec:
      image: '{{ .ImageName }}'
      tag: '{{ .ImageTag }}'
targets:
  {{ .TargetID }}:
    name: 'deps: bump Docker image "{{ .ImageName }}:{{ .ImageTag }}" digest'
    kind: 'dockerfile'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .TargetFile }}'
      instruction:
        keyword: '{{ .TargetKeyword }}'
        matcher: '{{ .TargetMatcher }}'
    sourceid: '{{ .SourceID }}-digest'
`
)
