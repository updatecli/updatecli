package reports

import (
	"bytes"
	"slices"
	"sort"
	"text/template"

	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/result"
)

const reportsTpl string = `
=============================

SUMMARY:


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

// Sort reports by result in the following order
// SUCCESS > SKIPPED > ATTENTION > FAILURE > UNKNOWN
func (r *Reports) Sort() {
	reports := *r

	sort.SliceStable(reports, func(i, j int) bool {
		resultToInteger := func(state string) int {
			switch state {
			case result.SUCCESS:
				return 0
			case result.SKIPPED:
				return 1
			case result.ATTENTION:
				return 2
			case result.FAILURE:
				return 3
			default:
				return 4
			}
		}

		return resultToInteger(reports[i].Result) < resultToInteger(reports[j].Result)
	})

}

// Summary display a summary of
func (r *Reports) Summary() (successCounter, changedCounter, failedCounter, skippedCounter int, actionLinks []string) {

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

		for _, action := range report.Actions {
			if !slices.Contains(actionLinks, action.Link) && action.Link != "" {
				actionLinks = append(actionLinks, action.Link)
			}
		}
	}

	return successCounter, changedCounter, failedCounter, skippedCounter, actionLinks
}
