package githubaction

const (
	// workflowManifestGiteaTemplate is the Go template used to generate Gitea action workflow manifests.
	// It relies on the blocks defined in workflowManifestActionPartials.
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
{{- template "versionfilter" (.Source "release") }}
{{- if .Digest }}
{{- template "gitDigestSource" (.Source "release") }}
{{- end }}
{{- template "gitSource" (.Source "tag") }}
{{- template "gitSource" (.Source "branch") }}

conditions:
  release: &release
    name: 'Check if {{ .ActionName }}@{{ .CheckReference }} is a Gitea release'
    kind: 'gitea/release'
    disablesourceinput: true
    spec:
      owner: '{{ .Owner }}'
      repository: '{{ .Repository }}'
      url: '{{ .URL }}'
      token: '{{ .Token }}'
      tag: '{{ .CheckReference }}'
{{- template "gitConditions" . }}

targets:
{{- template "target" (.Source "release") }}
{{ template "target" (.Source "tag") }}
{{ template "target" (.Source "branch") }}
`
)
