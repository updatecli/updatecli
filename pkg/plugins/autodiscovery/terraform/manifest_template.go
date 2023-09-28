package terraform

// terraformProviderManifestTemplate is the Go template used to generate Terraform manifest update
var terraformProviderManifestTemplate string = `name: 'Bump Terraform provider {{ .Provider }} version'
sources:
  latestVersion:
    name: 'Get latest version of the {{ .Provider }} provider'
    kind: terraform/registry
    spec:
      type: provider
      namespace: {{ .ProviderNamespace }}
      name: {{ .ProviderName }}
      versionfilter:
        kind: '{{ .VersionFilterKind }}'
        pattern: '{{ .VersionFilterPattern }}'
targets:
  terraformLockVersion:
    name: {{ .TargetName }}
    kind: terraform/lock
    sourceid: latestVersion
    spec:
      file: '{{ .TerraformLockFile }}'
      provider: '{{ .Provider }}'
      platforms:
{{- range .Platforms}}
        - {{.}}
{{- end }}
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{- end }}
`
