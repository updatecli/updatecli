package reports

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/result"
)

const reportsTpl string = `
=============================

REPORTS:

{{ range . }}
{{ if  .Err }}
{{- .Result }} {{ .Name -}}:{{"\n"}}
{{ "\t"}}Error: {{ .Err}}
{{ else }}
{{- .Result }} {{ .Name -}}:{{"\n"}}
{{- "\t"}}Sources:
{{ range $ID,$source := .Sources }}
{{- "\t" }}{{"\t"}}{{- $source.Result }} [{{ $ID }}] {{ $source.Name -}}({{- $source.Kind -}}){{"\n"}}
{{- end }}

{{- if .Conditions -}}
{{- "\t" }}Condition:
{{ range $ID, $condition := .Conditions }}
{{- "\t" }}{{"\t"}}{{- $condition.Result }} [{{ $ID }}] {{ $condition.Name -}}({{- $condition.Kind -}}){{"\n"}}
{{- end -}}
{{- end -}}

{{- "\t" -}}Target:
{{ range $ID, $target := .Targets }}
{{- "\t" }}{{"\t"}}{{- $target.Result }} [{{ $ID }}]  {{ $target.Name -}}({{- $target.Kind -}}){{"\n"}}
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

	logrus.Infof(reports)

	return nil
}

// Summary display a summary of
func (r *Reports) Summary() (int, int, int, error) {
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
		} else if report.Result == result.ATTENTION {
			changedCounter++
		} else {
			logrus.Infof("Unknown report result '%s'", report.Result)
		}
	}

	if failedCounter > 0 {
		return successCounter,
			changedCounter,
			failedCounter,
			fmt.Errorf("%d/%d job(s) failed", failedCounter, counter)
	}

	return successCounter, changedCounter, failedCounter, nil
}
