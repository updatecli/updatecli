package githubaction

const (
	// workflowManifestGiteaTemplate is the Go template used to generate Gitea action workflow manifests
	workflowManifestGiteaTemplate string = `name: 'deps: bump {{ .ActionName }} Gitea workflow'
{{- if .ActionID }}
actions:
  {{ .ActionID }}:
    title: 'deps: update {{ .ActionName }} to {{ "{{" }} source "release" {{ "}}" }}'
{{ end }}

sources:
  release:
    dependson:
      - 'condition#release:and'
    name: 'Get latest Gitea Release for {{ .ActionName }}'
    kind: 'gitearelease'
    spec:
      owner: '{{ .Owner }}'
      repository: '{{ .Repository }}'
      url: '{{ .URL }}'
      token: '{{ .Token }}'
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
      versionfilter:
        kind: '{{ .VersionFilterKind }}'
        pattern: '{{ .VersionFilterPattern }}'

conditions:
  release:
    name: 'Check if {{ .ActionName }}@{{ .Reference }} is a Gitea release'
    kind: 'gitearelease'
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
    name: 'deps(gitea): bump {{ .ActionName }} from {{ .Reference }} to {{ "{{" }} source "release" {{ "}}" }}'
    kind: 'yaml'
    sourceid: 'release'
    transformers:
      - addprefix: '{{ .ActionName }}@'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .File }}'
      key: '$.jobs.{{ .JobID }}.steps[{{ .StepID }}].uses'
      engine: 'yamlpath'

  tag:
    dependson:
      - 'condition#tag:and'
    disableconditions: true
    name: 'deps(gitea): bump {{ .ActionName }} from {{ .Reference }} to {{ "{{" }} source "tag" {{ "}}" }}'
    kind: 'yaml'
    sourceid: 'tag'
    transformers:
      - addprefix: '{{ .ActionName }}@'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .File }}'
      key: '$.jobs.{{ .JobID }}.steps[{{ .StepID }}].uses'
      engine: 'yamlpath'

  branch:
    dependson:
      - 'condition#branch:and'
    disableconditions: true
    name: 'deps(gitea): bump {{ .ActionName }} from {{ .Reference }} to {{ "{{" }} source "branch" {{ "}}" }}'
    kind: yaml
    sourceid: branch
    transformers:
      - addprefix: '{{ .ActionName }}@'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .File }}'
      key: '$.jobs.{{ .JobID }}.steps[{{ .StepID }}].uses'
      engine: 'yamlpath'
`
)
