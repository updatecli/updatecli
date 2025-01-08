package githubaction

const (
	// workflowManifestGitHubTemplate is the Go template used to generate GitHub action workflow manifests
	workflowManifestGitHubTemplate string = `name: 'deps: bump {{ .ActionName }} GitHub workflow'

sources:
  release:
    dependson:
      - 'condition#release:and'
    name: 'Get latest GitHub Release for {{ .ActionName }}'
    kind: 'githubrelease'
    spec:
      owner: '{{ .Owner }}'
      repository: '{{ .Repository }}'
      url: '{{ .URL }}'
      token: '{{ .Token }}'
{{- if .Digest }}
      key: 'hash'
{{- end }}
      versionfilter:
        kind: '{{ .VersionFilterKind }}'
        pattern: '{{ .VersionFilterPattern }}'

  tag:
    dependson:
      - 'condition#tag:and'
    name: 'Get latest tag for {{ .ActionName }}'
    kind: 'gittag'
    spec:
      url: "{{ .URL }}/{{ .Owner }}/{{ .Repository }}.git"
      password: '{{ .Token }}'
{{- if .Digest }}
      key: 'hash'
{{- end }}
      versionfilter:
        kind: '{{ .VersionFilterKind }}'
        pattern: '{{ .VersionFilterPattern }}'

  branch:
    dependson:
      - 'condition#branch:and'
    name: 'Get latest branch for {{ .ActionName }}'
    kind: 'gitbranch'
    spec:
      url: "{{ .URL }}/{{ .Owner }}/{{ .Repository }}.git"
      password: '{{ .Token }}'
{{- if .Digest }}
      key: 'hash'
{{- end }}
      versionfilter:
        kind: '{{ .VersionFilterKind }}'
        pattern: '{{ .VersionFilterPattern }}'

conditions:
  release:
    name: 'Check if {{ .ActionName }}@{{ .Reference }} is a GitHub release'
    kind: 'githubrelease'
    disablesourceinput: true
    spec:
      owner: '{{ .Owner }}'
      repository: '{{ .Repository }}'
      url: '{{ .URL }}'
      token: '{{ .Token }}'
      tag: '{{ .Reference }}'

  tag:
    name: 'Check if {{ .ActionName }}@{{ .Reference }} is a tag'
    kind: 'gittag'
    disablesourceinput: true
    spec:
      url: "{{ .URL }}/{{ .Owner }}/{{ .Repository }}.git"
      password: '{{ .Token }}'
      versionfilter:
        kind: 'regex'
        pattern: '^{{ .Reference }}$'

  branch:
    name: 'Check if {{ .ActionName }}@{{ .Reference }} is a branch'
    kind: 'gitbranch'
    disablesourceinput: true
    spec:
      branch: '{{ .Reference }}'
      url: "{{ .URL }}/{{ .Owner }}/{{ .Repository }}.git"
      password: '{{ .Token }}'

targets:
  release:
    dependson:
      - 'condition#release:and'
    disableconditions: true
    name: 'deps(github): bump Action release for {{ .ActionName }} from {{ .Reference }} to {{ "{{" }} source "release" {{ "}}" }}'
    kind: 'yaml'
    sourceid: 'release'
    transformers:
      - addprefix: '"{{ .ActionName }}@'
      - addsuffix: '"'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .File }}'
      key: '$.jobs.{{ .JobID }}.steps[{{ .StepID }}].uses'

  tag:
    dependson:
      - 'condition#tag:and'
    disableconditions: true
    name: 'deps(github): bump Action tag for {{ .ActionName }} from {{ .Reference }} to {{ "{{" }} source "tag" {{ "}}" }}'
    kind: 'yaml'
    sourceid: 'tag'
    transformers:
      - addprefix: '"{{ .ActionName }}@'
      - addsuffix: '"'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .File }}'
      key: '$.jobs.{{ .JobID }}.steps[{{ .StepID }}].uses'

  branch:
    dependson:
      - 'condition#branch:and'
    disableconditions: true
    name: 'deps(github): bump Action branch for {{ .ActionName }} from {{ .Reference }} to {{ "{{" }} source "branch" {{ "}}" }}'
    kind: yaml
    sourceid: branch
    transformers:
      - addprefix: '"{{ .ActionName }}@'
      - addsuffix: '"'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .File }}'
      key: '$.jobs.{{ .JobID }}.steps[{{ .StepID }}].uses'
`
)
