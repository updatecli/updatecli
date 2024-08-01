package githubaction

const (
	// workflowManifestTemplate is the Go template used to generate GitHub action workflow manifests
	workflowManifestTemplate string = `name: 'deps: bump {{ .ActionName }} GitHub workflow'

scms:
  github-action:
    kind: 'git'
    spec:
      url: "https://github.com/{{ .Owner }}/{{ .Repository }}.git"
      password: '{{ .Token }}'

sources:
  release:
    name: 'Get latest GitHub Release for {{ .ActionName }}'
    kind: 'githubrelease'
    spec:
      owner: '{{ .Owner }}'
      repository: '{{ .Repository }}'
      token: '{{ .Token }}'
      versionfilter:
        kind: '{{ .VersionFilterKind }}'
        pattern: '{{ .VersionFilterPattern }}'

  tag:
    name: 'Get latest tag for {{ .ActionName }}'
    kind: 'gittag'
    scmid: 'github-action'
    spec:
      versionfilter:
        kind: '{{ .VersionFilterKind }}'
        pattern: '{{ .VersionFilterPattern }}'

  branch:
    name: 'Get latest branch for {{ .ActionName }}'
    kind: 'gitbranch'
    scmid: 'github-action'
    spec:
      versionfilter:
        kind: '{{ .VersionFilterKind }}'
        pattern: '{{ .VersionFilterPattern }}'

conditions:
  release:
    name: 'Check if {{ .ActionName }}@{{ .Reference }} is a GitHub release'
    kind: 'githubrelease'
    disablesourceinput: true
    spec:
      owner: {{ .Owner }}
      repository: {{ .Repository }}
      token: '{{ .Token }}'
      tag: '{{ "{{" }} source "release" {{ "}}" }}'

  tag:
    name: 'Check if {{ .ActionName }}@{{ .Reference }} is a tag'
    kind: 'gittag'
    scmid: 'github-action'
    disablesourceinput: true
    spec:
      tag: '{{ .Reference }}'

  branch:
    name: 'Check if {{ .ActionName }}@{{ .Reference }} is a branch'
    kind: 'gitbranch'
    scmid: 'github-action'
    disablesourceinput: true
    spec:
      branch: '{{ .Reference }}'

targets:
  release:
    name: 'deps(github): bump Action release for {{ .ActionName }} from {{ .Reference }} to {{ "{{" }} source "release" {{ "}}" }}'
    kind: 'yaml'
    sourceid: 'release'
    conditionids:
      - 'release'
    transformers:
      - addprefix: '"{{ .ActionName }}@'
      - addsuffix: '"'
    spec:
      file: '{{ .File }}'
      key: '$.jobs.{{ .JobID }}.steps[{{ .StepID }}].uses'

  tag:
    name: 'deps(github): bump Action tag for {{ .ActionName }} from {{ .Reference }} to {{ "{{" }} source "tag" {{ "}}" }}'
    kind: 'yaml'
    sourceid: 'tag'
    conditionids:
      - 'tag'
    transformers:
      - addprefix: '"{{ .ActionName }}@'
      - addsuffix: '"'
    spec:
      file: '{{ .File }}'
      key: '$.jobs.{{ .JobID }}.steps[{{ .StepID }}].uses'

  branch:
    name: 'deps(github): bump Action branch for {{ .ActionName }} from {{ .Reference }} to {{ "{{" }} source "branch" {{ "}}" }}'
    kind: yaml
    sourceid: branch
    conditionids:
      - branch
    transformers:
      - addprefix: '"{{ .ActionName }}@'
      - addsuffix: '"'
    spec:
      file: '{{ .File }}'
      key: '$.jobs.{{ .JobID }}.steps[{{ .StepID }}].uses'
`
)
