package reports

// markdownReportTemplate is the Go template used to generate markdown report
var markdownReportTemplate string = `# {{ .PipelineTitle }}

Pipeline ID: ` + "`" + `{{ .ID }}` + "`" + `

{{- range .Targets}}

## {{ .Title }}

Target ID: ` + "`" + `{{ .ID }}` + "`" + `

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

{{- if .PipelineURL }}

Pipeline URL: [{{ .PipelineURL.Name }}]({{ .PipelineURL.URL }})
{{- end}}`
