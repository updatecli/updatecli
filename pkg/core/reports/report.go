package reports

import (
	"bytes"
	"text/template"

	"github.com/sirupsen/logrus"

	"github.com/olblak/updateCli/pkg/core/config"
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
{{ range .Sources }}
{{- "\t" }}{{"\t"}}{{- .Result }}  {{ .Name -}}({{- .Kind -}}){{"\n"}}
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
{{ range .Sources }}
{{- "\t" }}{{"\t"}}{{- .Result }}  {{ .Name -}}({{- .Kind -}}){{"\n"}}
{{- end }}

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
	Sources    []Stage
	Conditions []Stage
	Targets    []Stage
}

// Init init a new report for a specific configuration
//func (config *Config) InitReport() (report *Report) {
func (r *Report) Init(cfg *config.Config) {

	r.Name = cfg.Name
	r.Result = result.FAILURE

	for _, source := range cfg.Sources {
		r.Sources = append(r.Sources, Stage{
			Name:   source.Name,
			Kind:   source.Kind,
			Result: result.FAILURE,
		})
	}

	for _, condition := range cfg.Conditions {
		r.Conditions = append(r.Conditions, Stage{
			Name:   condition.Name,
			Kind:   condition.Kind,
			Result: result.FAILURE,
		})
	}

	for _, target := range cfg.Targets {
		r.Targets = append(r.Targets, Stage{
			Name:   target.Name,
			Kind:   target.Kind,
			Result: result.FAILURE,
		})
	}
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
