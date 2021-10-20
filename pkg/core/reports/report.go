package reports

import (
	"bytes"
	"text/template"

	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/result"
)

const (
	// CONDITIONREPORTTEMPLATE defines
	CONDITIONREPORTTEMPLATE string = `
{{- "\t" }}Condition:
{{ range $ID, $condition := .Conditions }}
{{- "\t" }}{{"\t"}}{{- $condition.Result }} [{{ $ID }}] {{ $condition.Name -}}({{- $condition.Kind -}}){{"\n"}}
{{- end -}}
`
	// TARGETREPORTTEMPLATE ...
	TARGETREPORTTEMPLATE string = `
{{- "\t" -}}Target:
{{ range $ID, $target := .Targets }}
{{- "\t" }}{{"\t"}}{{- $target.Result }} [{{ $ID }}]  {{ $target.Name -}}({{- $target.Kind -}}){{"\n"}}
{{- end }}
`
	// SOURCEREPORTTEMPLATE ...
	SOURCEREPORTTEMPLATE string = `
{{- "\t"}}Source:
{{ range $ID,$source := .Sources }}
{{- "\t" }}{{"\t"}}{{- $source.Result }} [{{ $ID }}]  {{ $source.Name -}}({{- $source.Kind -}}){{"\n"}}
{{- end }}
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
{{ range $ID, $source := .Sources }}
{{- "\t" }}{{"\t"}}{{- $source.Result }} [{{ $ID }}] {{ $source.Name -}}({{- $source.Kind -}}){{"\n"}}
{{- end }}

{{- if .Conditions -}}
{{- "\t" }}Condition:
{{ range  $ID, $condition := .Conditions }}
{{- "\t" }}{{"\t"}}{{- $condition.Result }} [{{ $ID }}] {{ $condition.Name -}}({{- $condition.Kind -}}){{"\n"}}
{{- end -}}
{{- end -}}

{{- "\t" -}}Target:
{{ range $ID,$target := .Targets }}
{{- "\t" }}{{"\t"}}{{- $target.Result }} [{{ $ID}}] {{ $target.Name -}}({{- $target.Kind -}}){{"\n"}}
{{- end }}
{{ end }}
`
)

// Report contains a list of Rules
type Report struct {
	Name       string
	Err        string
	Result     string
	Sources    map[string]Stage
	Conditions map[string]Stage
	Targets    map[string]Stage
}

// Init init a new report for a specific configuration
//func (config *Config) InitReport() (report *Report) {
func (r *Report) Init(name string, sourceNbr, conditionNbr, targetNbr int) {

	r.Name = name
	r.Result = result.FAILURE

	r.Sources = make(map[string]Stage, sourceNbr)
	r.Conditions = make(map[string]Stage, conditionNbr)
	r.Targets = make(map[string]Stage, targetNbr)
}

// String return a report as a string
func (r *Report) String(mode string) (report string, err error) {
	t := &template.Template{}

	switch mode {
	case "conditions":
		t = template.Must(template.New("reports").Parse(CONDITIONREPORTTEMPLATE))
	case "sources":
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
