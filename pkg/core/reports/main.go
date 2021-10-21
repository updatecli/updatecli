package reports

import (
	"bytes"
	"strings"
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
func (r *Reports) Summary() (successCounter, changedCounter, failedCounter, skippedCounter int) {

	reports := *r

	for _, report := range reports {
		if strings.Compare(report.Result, result.SUCCESS) == 0 {
			successCounter++
		} else if strings.Compare(report.Result, result.FAILURE) == 0 {
			failedCounter++
		} else if strings.Compare(report.Result, result.ATTENTION) == 0 {
			changedCounter++
		} else if strings.Compare(report.Result, result.SKIPPED) == 0 {
			skippedCounter++
		} else {
			logrus.Infof("Unknown report result %q with named %q.", report.Result, report.Name)
		}
	}

	return successCounter, changedCounter, failedCounter, skippedCounter
}
