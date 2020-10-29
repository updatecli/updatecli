package reports

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/olblak/updateCli/pkg/result"
)

const reportsTpl string = `
=============================

REPORTS:

{{ range . }}
{{ if  .Err }}
{{- .Result }} {{ .Name -}}{{"\n"}}
{{ "\t"}}Error: {{ .Err}}
{{ else }}
{{- .Result }} {{ .Name -}}{{"\n"}}
{{- "\t"}}Source:
{{ "\t"}}{{"\t"}}{{- .Source.Result }}  {{ .Source.Name -}}({{- .Source.Kind -}}){{"\n"}}

{{- "\t" }}Condition:
{{ range .Conditions }} 
{{- "\t" }}{{"\t"}}{{- .Result }}  {{ .Name -}}({{- .Kind -}}){{"\n"}}
{{- end -}}

{{- "\t" -}}Target:
{{ range .Targets }} 
{{- "\t" }}{{"\t"}}{{- .Result }}  {{ .Name -}}({{- .Kind -}}){{"\n"}}
{{- end }}
{{ end }}
{{ end }}
`

// Reports contains a list of report
type Reports []Report

// Show return a small reports of what has been changed
func (r *Reports) Show() error {
	t := template.Must(template.New("reports").Parse(reportsTpl))

	reports := ""

	buffer := new(bytes.Buffer)

	err := t.Execute(buffer, r)

	reports = buffer.String()

	if err != nil {
		return err
	}

	fmt.Println(reports)

	return nil
}

// Summary display a summary of
func (r *Reports) Summary() error {
	counter := 0
	successCounter := 0
	changedCounter := 0
	failedCounter := 0

	reports := *r

	for _, report := range reports {
		counter++
		if report.Result == result.SUCCESS {
			successCounter++
		} else if report.Result == result.FAILURE {
			failedCounter++
		} else if report.Result == result.CHANGED {
			changedCounter++
		} else {
			fmt.Printf("Unknown report result '%s'\n", report.Result)
		}
	}

	fmt.Printf("Run Summary\n")
	fmt.Printf("===========\n\n")
	fmt.Printf("%d job run\n", counter)
	fmt.Printf("%d job succeed\n", successCounter)
	fmt.Printf("%d job failed\n", failedCounter)
	fmt.Printf("%d job applied changes\n", changedCounter)

	return nil
}
