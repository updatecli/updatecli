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
    kind: 'gitea/release'
    spec:
      owner: '{{ .Owner }}'
      repository: '{{ .Repository }}'
      url: '{{ .URL }}'
      token: '{{ .Token }}'
{{- template "age" .Age }}
      versionfilter:
        kind: '{{ .VersionFilterKind }}'
        pattern: '{{ .VersionFilterPattern }}'
        {{- if or (eq .VersionFilterKind "regex/semver") (eq .VersionFilterKind "regex/time") }}
        regex: '{{ .VersionFilterRegex }}'
        {{- end }}

{{- if .Digest }}

  release_digest:
    dependson:
      - 'condition#release:and'
    name: 'Get commit for {{ .ActionName }} release'
    kind: 'gittag'
    spec:
      url: "{{ .URL }}/{{ .Owner }}/{{ .Repository }}.git"
      username: '{{ if .Token }}oauth2{{ end }}'
      password: '{{ .Token }}'
      lsremote: true
      key: 'hash'
      versionfilter:
        kind: 'regex'
        pattern: '^\Q{{ "{{" }} source "release" {{ "}}" }}\E$'

{{- end }}

  tag:
    dependson:
      - 'condition#tag:and'
    name: 'Get latest tag for {{ .ActionName }}'
    kind: 'gittag'
    spec:
      url: "{{ .URL }}/{{ .Owner }}/{{ .Repository }}.git"
      username: '{{ if .Token }}oauth2{{ end }}'
      password: '{{ .Token }}'
{{- if and .Age.IsZero (ne .VersionFilterKind "latest") }}
      lsremote: true
{{- end }}
{{- template "age" .Age }}
      versionfilter:
        kind: '{{ .VersionFilterKind }}'
        pattern: '{{ .VersionFilterPattern }}'
        {{- if or (eq .VersionFilterKind "regex/semver") (eq .VersionFilterKind "regex/time") }}
        regex: '{{ .VersionFilterRegex }}'
        {{- end }}

{{- if .Digest }}

  tag_digest:
    dependson:
      - 'condition#tag:and'
    name: 'Get commit for {{ .ActionName }} tag'
    kind: 'gittag'
    spec:
      url: "{{ .URL }}/{{ .Owner }}/{{ .Repository }}.git"
      username: '{{ if .Token }}oauth2{{ end }}'
      password: '{{ .Token }}'
      lsremote: true
      key: 'hash'
      versionfilter:
        kind: 'regex'
        pattern: '^\Q{{ "{{" }} source "tag" {{ "}}" }}\E$'

{{- end }}

  branch:
    dependson:
      - 'condition#branch:and'
    name: 'Get latest branch for {{ .ActionName }}'
    kind: 'gitbranch'
    spec:
      url: "{{ .URL }}/{{ .Owner }}/{{ .Repository }}.git"
      username: '{{ if .Token }}oauth2{{ end }}'
      password: '{{ .Token }}'
{{- if and .Age.IsZero (ne .BranchVersionFilterKind "latest") }}
      lsremote: true
{{- end }}
{{- template "age" .Age }}
      versionfilter:
        kind: '{{ .BranchVersionFilterKind }}'
        pattern: '{{ .BranchVersionFilterPattern }}'
        {{- if or (eq .BranchVersionFilterKind "regex/semver") (eq .BranchVersionFilterKind "regex/time") }}
        regex: '{{ .VersionFilterRegex }}'
        {{- end }}

{{- if .Digest }}

  branch_digest:
    dependson:
      - 'condition#branch:and'
    name: 'Get commit for {{ .ActionName }} branch'
    kind: 'gitbranch'
    spec:
      url: "{{ .URL }}/{{ .Owner }}/{{ .Repository }}.git"
      username: '{{ if .Token }}oauth2{{ end }}'
      password: '{{ .Token }}'
      lsremote: true
      key: 'hash'
      versionfilter:
        kind: 'regex'
        pattern: '^\Q{{ "{{" }} source "branch" {{ "}}" }}\E$'

{{- end }}

conditions:
  release:
    name: 'Check if {{ .ActionName }}@{{ .CheckReference }} is a Gitea release'
    kind: 'gitea/release'
    disablesourceinput: true
    spec:
      owner: '{{ .Owner }}'
      repository: '{{ .Repository }}'
      url: '{{ .URL }}'
      token: '{{ .Token }}'
      tag: '{{ .CheckReference }}'

  tag:
    name: 'Check if {{ .ActionName }}@{{ .CheckReference }} is a tag'
    kind: 'gittag'
    disablesourceinput: true
    spec:
      url: "{{ .URL }}/{{ .Owner }}/{{ .Repository }}.git"
      username: '{{ if .Token }}oauth2{{ end }}'
      password: '{{ .Token }}'
      lsremote: true
      versionfilter:
        kind: 'regex'
        pattern: '^\Q{{ .CheckReference }}\E$'

  branch:
    name: 'Check if {{ .ActionName }}@{{ .CheckReference }} is a branch'
    kind: 'gitbranch'
    disablesourceinput: true
    spec:
      branch: '{{ .CheckReference }}'
      url: "{{ .URL }}/{{ .Owner }}/{{ .Repository }}.git"
      username: '{{ if .Token }}oauth2{{ end }}'
      password: '{{ .Token }}'
      lsremote: true

targets:
  release:
    dependson:
      - 'condition#release:and'
    disableconditions: true
    name: 'deps(gitea): bump {{ .ActionName }} from {{ .Reference }} to {{ "{{" }} source "release{{if .Digest}}_digest{{end}}" {{ "}}" }}{{ if .Digest }} (Pinned from {{ "{{" }} source "release" {{ "}}" }}){{ end }}'
    kind: 'yaml'
    sourceid: 'release{{if .Digest}}_digest{{end}}'
    transformers:
      - addprefix: '{{ .ActionName }}@'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .File }}'
      key: '{{ .TargetKey }}'
{{- if .Digest }}
      comment: '{{ "{{" }} source "release" {{ "}}" }}'
{{- end }}

  tag:
    dependson:
      - 'condition#tag:and'
    disableconditions: true
    name: 'deps(gitea): bump {{ .ActionName }} from {{ .Reference }} to {{ "{{" }} source "tag{{if .Digest}}_digest{{end}}" {{ "}}" }}{{ if .Digest }} (Pinned from {{ "{{" }} source "tag" {{ "}}" }}){{ end }}'
    kind: 'yaml'
    sourceid: 'tag{{if .Digest}}_digest{{end}}'
    transformers:
      - addprefix: '{{ .ActionName }}@'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .File }}'
      key: '{{ .TargetKey }}'
{{- if .Digest }}
      comment: '{{ "{{" }} source "tag" {{ "}}" }}'
{{- end }}

  branch:
    dependson:
      - 'condition#branch:and'
    disableconditions: true
    name: 'deps(gitea): bump {{ .ActionName }} from {{ .Reference }} to {{ "{{" }} source "branch{{if .Digest}}_digest{{end}}" {{ "}}" }}{{ if .Digest }} (Pinned from {{ "{{" }} source "branch" {{ "}}" }}){{ end }}'
    kind: yaml
    sourceid: branch{{if .Digest}}_digest{{end}}
    transformers:
      - addprefix: '{{ .ActionName }}@'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .File }}'
      key: '{{ .TargetKey }}'
{{- if .Digest }}
      comment: '{{ "{{" }} source "branch" {{ "}}" }}'
{{- end }}
`
)
