package reports

import (
	"bytes"
	"github.com/sirupsen/logrus"
	"text/template"

	"github.com/olblak/updateCli/pkg/core/result"
)

const (
	// CONDITIONREPORTTEMPLATE defines
	CONDITIONREPORTTEMPLATE string = `
{{- "\t" }}Condition:
{{ range .Conditions }}
{{- "\t" }}{{"\t"}}{{- .Result }}  {{ .Name -}}({{- .Kind -}}){{"\n"}}
{{- end -}}
`
	// TARGETREPORTTEMPLATE ...
	TARGETREPORTTEMPLATE string = `
{{- "\t" -}}Target:
{{ range .Targets }}
{{- "\t" }}{{"\t"}}{{- .Result }}  {{ .Name -}}({{- .Kind -}}){{"\n"}}
{{- end }}
`
	// SOURCEREPORTTEMPLATE ...
	SOURCEREPORTTEMPLATE string = `
{{- "\t"}}Source:
{{ "\t"}}{{"\t"}}{{- .Source.Result }}  {{ .Source.Name -}}({{- .Source.Kind -}}){{"\n"}}
`

	// REPORTTEMPLATE ...
	REPORTTEMPLATE string = `
=============================

REPORTS:

{{ if  .Err }}
{{- .Result }} {{ .Name -}}{{"\n"}}
{{ "\t"}}Error: {{ .Err}}
{{ else }}
{{- .Result }} {{ .Name -}}{{"\n"}}
{{- "\t"}}Source:
{{ "\t"}}{{"\t"}}{{- .Source.Result }}  {{ .Source.Name -}}({{- .Source.Kind -}}){{"\n"}}

{{- if .Conditions -}}
{{- "\t" }}Condition:
{{ range .Conditions }}
{{- "\t" }}{{"\t"}}{{- .Result }}  {{ .Name -}}({{- .Kind -}}){{"\n"}}
{{- end -}}
{{- end -}}

{{- "\t" -}}Target:
{{ range .Targets }}
{{- "\t" }}{{"\t"}}{{- .Result }}  {{ .Name -}}({{- .Kind -}}){{"\n"}}
{{- end }}
{{ end }}
`
)

// Report contains a list of Rules
type Report struct {
	Name       string
	Err        string
	Result     string
	Source     Stage
	Conditions []Stage
	Targets    []Stage
}

// Init init a new report for a specific configuration
//func (config *Config) InitReport() (report *Report) {
func Init(
	name string,
	source Stage,
	conditions []Stage,
	targets []Stage,
) (report Report) {

	report.Name = name
	report.Result = result.FAILURE

	report.Source = Stage{
		Name:   source.Name,
		Kind:   source.Kind,
		Result: result.FAILURE,
	}

	for _, condition := range conditions {
		report.Conditions = append(report.Conditions, Stage{
			Name:   condition.Name,
			Kind:   condition.Kind,
			Result: result.FAILURE,
		})
	}

	for _, target := range targets {
		report.Targets = append(report.Targets, Stage{
			Name:   target.Name,
			Kind:   target.Kind,
			Result: result.FAILURE,
		})
	}

	return report
}

// String return a report as a string
func (r *Report) String(mode string) (report string, err error) {
	t := &template.Template{}

	switch mode {
	case "conditions":
		t = template.Must(template.New("reports").Parse(CONDITIONREPORTTEMPLATE))
	case "source":
		t = template.Must(template.New("reports").Parse(SOURCEREPORTTEMPLATE))
	case "targets":
		t = template.Must(template.New("reports").Parse(TARGETREPORTTEMPLATE))
	case "all":
		t = template.Must(template.New("reports").Parse(REPORTTEMPLATE))
	default:
		logrus.Infof("Wrong report template provided")
	}

	buffer := new(bytes.Buffer)

	err = t.Execute(buffer, r)

	if err != nil {
		return "", err
	}

	report = buffer.String()

	return report, nil
}
