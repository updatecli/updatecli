package githubaction

// workflowManifestActionPartials holds the template blocks shared by the GitHub and Gitea
// action manifests. Each source, condition, or target block starts with its own newlines
// and ends without one, so callers invoke them with "{{- template ... }}".
//
// Blocks marked "(source)" take an actionSourceParams, as returned by
// actionManifestParams.Source, the others take an actionManifestParams.
const workflowManifestActionPartials string = `
{{- /* versionfilter (source) emits the version filter of a source spec */ -}}
{{- define "versionfilter" }}
      versionfilter:
        kind: '{{ .FilterKind }}'
        {{- if ne .FilterKind "latest" }}
        pattern: '{{ .FilterPattern }}'
        {{- end }}
        {{- if or (eq .FilterKind "regex/semver") (eq .FilterKind "regex/time") }}
        regex: '{{ .VersionFilterRegex }}'
        {{- end }}
{{- end }}

{{- /* gitAuth emits the credentials of a gittag or gitbranch spec */ -}}
{{- define "gitAuth" }}
      url: "{{ .URL }}/{{ .Owner }}/{{ .Repository }}.git"
      username: '{{ .GitUsername }}'
      password: '{{ .Token }}'
{{- end }}

{{- /* gitDigestSource (source) resolves the commit of the reference found by its source */ -}}
{{- define "gitDigestSource" }}

  {{ .SourceID }}_digest:
    dependson:
      - 'condition#{{ .SourceID }}:and'
    name: 'Get commit for {{ .ActionName }} {{ .SourceID }}'
    kind: '{{ .GitKind }}'
    spec:
{{- template "gitAuth" . }}
      lsremote: true
      key: 'hash'
      versionfilter:
        kind: 'regex'
        pattern: '^\Q{{ "{{" }} source "{{ .SourceID }}" {{ "}}" }}\E$'
{{- end }}

{{- /* gitSource (source) finds the latest tag or branch of the action */ -}}
{{- define "gitSource" }}

  {{ .SourceID }}:
    dependson:
      - 'condition#{{ .SourceID }}:and'
    name: 'Get latest {{ .SourceID }} for {{ .ActionName }}'
    kind: '{{ .GitKind }}'
    spec:
{{- template "gitAuth" . }}
{{- if and .Age.IsZero (ne .FilterKind "latest") }}
      lsremote: true
{{- end }}
{{- template "age" .Age }}
{{- template "versionfilter" . }}
{{- if .Digest }}
{{- template "gitDigestSource" . }}
{{- end }}
{{- end }}

{{- /* gitConditions gives releases priority over tags, and tags over branches */ -}}
{{- define "gitConditions" }}

  not_release:
    <<: *release
    failwhen: true

  tag: &tag
    dependson:
      - 'condition#not_release:and'
    name: 'Check if {{ .ActionName }}@{{ .CheckReference }} is a tag'
    kind: 'gittag'
    disablesourceinput: true
    spec:
{{- template "gitAuth" . }}
      lsremote: true
      versionfilter:
        kind: 'regex'
        pattern: '^\Q{{ .CheckReference }}\E$'

  not_tag:
    <<: *tag
    failwhen: true

  branch:
    dependson:
      - 'condition#not_tag:and'
    name: 'Check if {{ .ActionName }}@{{ .CheckReference }} is a branch'
    kind: 'gitbranch'
    disablesourceinput: true
    spec:
      branch: '{{ .CheckReference }}'
{{- template "gitAuth" . }}
      lsremote: true
{{- end }}

{{- /* target (source) updates the workflow step from its source */ -}}
{{- define "target" }}
  {{ .SourceID }}:
    dependson:
      - 'condition#{{ .SourceID }}:and'
    disableconditions: true
    name: '{{ .TargetTitle }} from {{ .Reference }} to {{ "{{" }} source "{{ .TargetSourceID }}" {{ "}}" }}{{ if .Digest }} (Pinned from {{ "{{" }} source "{{ .SourceID }}" {{ "}}" }}){{ end }}'
    kind: 'yaml'
    sourceid: '{{ .TargetSourceID }}'
    transformers:
      - addprefix: '{{ .ActionName }}@'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    spec:
      file: '{{ .File }}'
      key: '{{ .TargetKey }}'
{{- if .Digest }}
      comment: '{{ "{{" }} source "{{ .SourceID }}" {{ "}}" }}'
{{- end }}
{{- end }}`
