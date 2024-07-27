package terragrunt

// terraformProviderManifestTemplate is the Go template used to generate Terraform manifest update
var terragruntModuleManifestTemplate = `name: 'Bump Terraform module {{ .Module }} version'
{{- if .ModuleSourceScm }}
scms:
  {{ .ModuleSourceScm }}:
    kind: "git"
    spec:
      url: {{ .ModuleSourceScmUrl }}
{{- end }}
sources:
  latestVersion:
    name: 'Get latest version of the {{ .Module }} module'
    kind: {{ .SourceTypeKind }}
{{- if .Transformers }}
    transformers:
{{- end }}
{{- range .Transformers }}
      - {{.Kind}}: {{.Value}}
{{- end }}
    spec:
      versionfilter:
        kind: '{{ .VersionFilterKind }}'
        pattern: '{{ .VersionFilterPattern }}'
{{- if eq .SourceType "registry" }}
      type: module
{{- if .ModuleHost }}
      hostname: {{ .ModuleHost }}
{{- end }}
      namespace: {{ .ModuleNamespace }}
      name: {{ .ModuleName }}
      targetsystem: {{ .ModuleTargetSystem }}
{{- else if eq .SourceType "git" }}
    scmid: {{ .ModuleSourceScm }}
{{- end }}
targets:
  terragruntModuleFile:
    name: {{ .TargetName }}
    kind: hcl
    sourceid: latestVersion
    spec:
      file: '{{ .TerragruntModulePath }}'
      path: '{{ .TargetPath }}'
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{- end }}
`
