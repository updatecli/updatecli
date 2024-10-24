package reports

// markdownReportTemplate is the Go template used to generate markdown report
var markdownReportTemplate string = `# {{ .PipelineTitle }}

{{- range .Targets}}

## {{ .Title }}

{{- if .Description }}

{{ .Description }}
{{- end}}

{{- range .Changelogs}}

### {{ .Title }}

{{- if .Description }}

` + "```" + `
{{ .Description }}
` + "```" + `
{{- end}}

{{- end}}

{{- end}}

{{- if .PipelineUrl }}

[{{ .PipelineUrl.Name }}]({{ .PipelineUrl.URL }})
{{- end}}`
