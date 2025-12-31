package terragrunt

// terragruntModuleManifestTemplate is the Go template used to generate Terragrunt manifest update
var terragruntModuleManifestTemplate = `name: 'Bump Terragrunt module {{ .Module }} version'
{{- if .ActionID }}
actions:
  {{ .ActionID }}:
    title: '{{ .TargetName }}'
{{ end }}
{{- if .ModuleSourceScm }}
scms:
  {{ .ModuleSourceScm }}:
    kind: 'git'
    spec:
      url: '{{ .ModuleSourceScmUrl }}'
{{- if .GitHubToken }}
      password: '{{ .GitHubToken }}'
{{- end }}
{{- end }}
sources:
  latestVersion:
    name: 'Get latest version of the {{ .Module }} module'
    kind: '{{ .SourceTypeKind }}'
{{- if .Transformers }}
    transformers:
{{- end }}
{{- range .Transformers }}
      - {{.Kind}}: '{{.Value}}'
{{- end }}
    spec:
      versionfilter:
        kind: '{{ .VersionFilterKind }}'
        pattern: '{{ .VersionFilterPattern }}'
        {{- if or (eq .VersionFilterKind "regex/semver") (eq .VersionFilterKind "regex/time") }}
        regex: '{{ .VersionFilterRegex }}'
        {{- end }}
{{- if eq .SourceType "registry" }}
      type: 'module'
{{- if .ModuleHost }}
      hostname: '{{ .ModuleHost }}'
{{- end }}
      namespace: '{{ .ModuleNamespace }}'
      name: '{{ .ModuleName }}'
      targetsystem: '{{ .ModuleTargetSystem }}'
{{- else if eq .SourceType "git" }}
    scmid: '{{ .ModuleSourceScm }}'
{{- end }}
targets:
  terragruntModuleFile:
    name: '{{ .TargetName }}'
    kind: 'hcl'
    sourceid: 'latestVersion'
    spec:
      file: '{{ .TerragruntModulePath }}'
      path: '{{ .TargetPath }}'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{- end }}
`
