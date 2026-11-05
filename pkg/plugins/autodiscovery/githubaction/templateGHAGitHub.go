package githubaction

const (
	// workflowManifestGitHubTemplate is the Go template used to generate GitHub action workflow manifests.
	// It relies on the blocks defined in workflowManifestActionPartials.
	workflowManifestGitHubTemplate string = `name: 'deps: bump {{ .ActionName }} GitHub workflow'
{{- if .ActionID }}
actions:
  {{ .ActionID }}:
    title: 'deps: update {{ .ActionName }} to {{ "{{" }} source "release" {{ "}}" }}'
{{ end }}

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
{{- template "age" .Age }}
{{- template "versionfilter" (.Source "release") }}
{{- if .Digest }}

  release_digest:
    dependson:
      - 'condition#release:and'
    name: 'Get commit for {{ .ActionName }} release'
    kind: 'githubrelease'
    spec:
      owner: '{{ .Owner }}'
      repository: '{{ .Repository }}'
      url: '{{ .URL }}'
      token: '{{ .Token }}'
      key: 'taghash'
      versionfilter:
        kind: 'regex'
        pattern: '^\Q{{ "{{" }} source "release" {{ "}}" }}\E$'
{{- end }}
{{- template "gitSource" (.Source "tag") }}
{{- template "gitSource" (.Source "branch") }}

conditions:
  release: &release
    name: 'Check if {{ .ActionName }}@{{ .CheckReference }} is a GitHub release'
    kind: 'githubrelease'
    disablesourceinput: true
    spec:
      owner: '{{ .Owner }}'
      repository: '{{ .Repository }}'
      url: '{{ .URL }}'
      token: '{{ .Token }}'
      tag: '{{ .CheckReference }}'
{{- /* An explicit type filter disables the fallback to Git tags. */}}
      typefilter:
        release: true
{{- template "gitConditions" . }}

targets:
{{- template "target" (.Source "release") }}
{{ template "target" (.Source "tag") }}
{{ template "target" (.Source "branch") }}
`
)
