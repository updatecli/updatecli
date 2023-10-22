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
{{ if .Err }}
{{- .Result }} {{ .Name -}}:{{"\n"}}
{{ "\t"}}Error: {{ .Err}}
{{ else }}
{{- .Result }} {{ .Name -}}:{{"\n"}}
{{- if .Sources -}}
{{- "\t"}}Source:
{{ range $ID,$source := .Sources }}
{{- "\t" }}{{"\t"}}{{- $source.Result }} [{{ $ID }}] {{ $source.Name }}{{"\n"}}
{{- end -}}
{{- end -}}

{{- if .Conditions -}}
{{- "\t" }}Condition:
{{ range $ID, $condition := .Conditions }}
{{- "\t" }}{{"\t"}}{{- $condition.Result }} [{{ $ID }}] {{ $condition.Name }}{{"\n"}}
{{- end -}}
{{- end -}}
{{- if .Targets -}}
{{- "\t" -}}Target:
{{ range $ID, $target := .Targets }}
{{- "\t" }}{{"\t"}}{{- $target.Result }} [{{ $ID }}] {{ $target.Name }}{{"\n"}}
{{- end }}
{{ end }}
{{- if .ReportURL }}
{{- "\t"}}=> Report available on {{ .ReportURL -}}{{"\n"}}
{{- end -}}
{{- end -}}
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
		switch report.Result {

		case result.SUCCESS:
			successCounter++

		case result.FAILURE:
			failedCounter++

		case result.SKIPPED:
			skippedCounter++

		case result.ATTENTION:
			changedCounter++

		default:
			logrus.Infof("Unknown report result %q with named %q.", report.Result, report.Name)
		}
	}

	return successCounter, changedCounter, failedCounter, skippedCounter
}

func (r *Report) SetResult(inputResult string) error {

	switch r.Result {

	case result.SUCCESS:
		switch inputResult {
		case result.FAILURE:
			r.Result = result.FAILURE
		case result.ATTENTION:
			r.Result = result.ATTENTION
		}
	case result.ATTENTION:
		switch inputResult {
		case result.FAILURE:
			r.Result = result.FAILURE
		}

	case result.SKIPPED:
		r.Result = inputResult

	case result.FAILURE:
		// nothing else to do

	default:
		return fmt.Errorf("something went wrong in handling resource result, unknowned report result %q", r.Result)
	}
	return nil
}
