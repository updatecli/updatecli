package reports

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/olblak/updateCli/pkg/config"
	"github.com/olblak/updateCli/pkg/result"
)

const reportsTpl string = `
=============================

REPORTS:

{{ range . }}
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

// New init a new reports
func New(config *config.Config) (report Report) {

	report.Result = result.FAILURE

	report.Source = Stage{
		Name:   config.Source.Name,
		Kind:   config.Source.Kind,
		Result: result.FAILURE,
	}

	for _, condition := range config.Conditions {
		report.Conditions = append(report.Conditions, Stage{
			Name:   condition.Name,
			Kind:   condition.Kind,
			Result: result.FAILURE,
		})
	}

	for _, target := range config.Targets {
		report.Targets = append(report.Targets, Stage{
			Name:   target.Name,
			Kind:   target.Kind,
			Result: result.FAILURE,
		})
	}

	return report
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
