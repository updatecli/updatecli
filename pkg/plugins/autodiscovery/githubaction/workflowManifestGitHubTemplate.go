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
      versionfilter:
        kind: '{{ .VersionFilterKind }}'
        pattern: '{{ .VersionFilterPattern }}'

{{- if .Digest }}

  release_digest:
    dependson:
      - 'condition#release:and'
    name: 'Get latest GitHub Release for {{ .ActionName }}'
    kind: 'githubrelease'
    spec:
      owner: '{{ .Owner }}'
      repository: '{{ .Repository }}'
      url: '{{ .URL }}'
      token: '{{ .Token }}'
      key: 'hash'
      versionfilter:
        kind: '{{ .VersionFilterKind }}'
        pattern: '{{ .VersionFilterPattern }}'

{{- end }}

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

{{- if .Digest }}

  tag_digest:
    dependson:
      - 'condition#tag:and'
    name: 'Get latest tag for {{ .ActionName }}'
    kind: 'gittag'
    spec:
      url: "{{ .URL }}/{{ .Owner }}/{{ .Repository }}.git"
      password: '{{ .Token }}'
      key: 'hash'
      versionfilter:
        kind: '{{ .VersionFilterKind }}'
        pattern: '{{ .VersionFilterPattern }}'

{{- end }}

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

{{- if .Digest }}

  branch_digest:
    dependson:
      - 'condition#branch:and'
    name: 'Get latest branch for {{ .ActionName }}'
    kind: 'gitbranch'
    spec:
      url: "{{ .URL }}/{{ .Owner }}/{{ .Repository }}.git"
      password: '{{ .Token }}'
      key: 'hash'
      versionfilter:
        kind: '{{ .VersionFilterKind }}'
        pattern: '{{ .VersionFilterPattern }}'

{{- end }}

conditions:
  release:
    name: 'Check if {{ .ActionName }}@{{ if .Digest }}{{ .PinReference }}{{ else }}{{ .Reference }}{{ end }} is a GitHub release'
    kind: 'githubrelease'
    disablesourceinput: true
    spec:
      owner: '{{ .Owner }}'
      repository: '{{ .Repository }}'
      url: '{{ .URL }}'
      token: '{{ .Token }}'
      tag: '{{ if .Digest }}{{ .PinReference }}{{ else }}{{ .Reference }}{{ end }}'

  tag:
    name: 'Check if {{ .ActionName }}@{{ if .Digest }}{{ .PinReference }}{{ else }}{{ .Reference }}{{ end }} is a tag'
    kind: 'gittag'
    disablesourceinput: true
    spec:
      url: "{{ .URL }}/{{ .Owner }}/{{ .Repository }}.git"
      password: '{{ .Token }}'
      versionfilter:
        kind: 'regex'
        pattern: '^{{ if .Digest }}{{ .PinReference }}{{ else }}{{ .Reference }}{{ end }}$'

  branch:
    name: 'Check if {{ .ActionName }}@{{ if .Digest }}{{ .PinReference }}{{ else }}{{ .Reference }}{{ end }} is a branch'
    kind: 'gitbranch'
    disablesourceinput: true
    spec:
      branch: '{{ if .Digest }}{{ .PinReference }}{{ else }}{{ .Reference }}{{ end }}'
      url: "{{ .URL }}/{{ .Owner }}/{{ .Repository }}.git"
      password: '{{ .Token }}'

targets:
  release:
    dependson:
      - 'condition#release:and'
    disableconditions: true
    name: 'deps(github): bump Action release for {{ .ActionName }} from {{ .Reference }} to {{ "{{" }} source "release{{if .Digest}}_digest{{end}}" {{ "}}" }}{{ if .Digest }} (Pinned from {{ "{{" }} source "release" {{ "}}" }}){{ end }}'
    kind: 'yaml'
    sourceid: 'release{{if .Digest}}_digest{{end}}'
    transformers:
      - addprefix: '"{{ .ActionName }}@'
      - addsuffix: '"'
{{- if .Digest }}
      - addsuffix: '  # pinned from {{ "{{" }} source "release" {{ "}}" }} by updatecli (do-not-remove-comment)'
{{- end }}
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
    name: 'deps(github): bump Action tag for {{ .ActionName }} from {{ .Reference }} to {{ "{{" }} source "tag{{if .Digest}}_digest{{end}}" {{ "}}" }}{{ if .Digest }} (Pinned from {{ "{{" }} source "tag" {{ "}}" }}){{ end }}'
    kind: 'yaml'
    sourceid: 'tag{{if .Digest}}_digest{{end}}'
    transformers:
      - addprefix: '"{{ .ActionName }}@'
      - addsuffix: '"'
{{- if .Digest }}
      - addsuffix: '  # pinned from {{ "{{" }} source "tag" {{ "}}" }} by updatecli (do-not-remove-comment)'
{{- end }}
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
    name: 'deps(github): bump Action branch for {{ .ActionName }} from {{ .Reference }} to {{ "{{" }} source "branch{{if .Digest}}_digest{{end}}" {{ "}}" }}{{ if .Digest }} (Pinned from {{ "{{" }} source "branch" {{ "}}" }}){{ end }}'
    kind: yaml
    sourceid: branch{{if .Digest}}_digest{{end}}
    transformers:
      - addprefix: '"{{ .ActionName }}@'
      - addsuffix: '"'
{{- if .Digest }}
      - addsuffix: '  # pinned from {{ "{{" }} source "branch" {{ "}}" }} by updatecli (do-not-remove-comment)'
{{- end }}
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .File }}'
      key: '$.jobs.{{ .JobID }}.steps[{{ .StepID }}].uses'
`
)
