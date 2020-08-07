package reports

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/olblak/updateCli/pkg/config"
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

// Update report based on latest information
func (r *Report) Update(config *config.Config) {

	r.Source.Kind = config.Source.Kind
	r.Source.Name = config.Source.Name
	r.Source.Result = config.Source.Result

	i := 0
	for _, condition := range config.Conditions {
		c := &r.Conditions[i]
		c.Name = condition.Name
		c.Kind = condition.Kind
		c.Result = condition.Result
		i++
	}

	i = 0
	for _, target := range config.Targets {
		t := &r.Targets[i]
		t.Name = target.Name
		t.Kind = target.Kind
		t.Result = target.Result
		i++
	}

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
		fmt.Println("Wrong report template provided")
	}

	buffer := new(bytes.Buffer)

	err = t.Execute(buffer, r)

	if err != nil {
		return "", err
	}

	report = buffer.String()

	return report, nil
}
