package reports

import (
	"crypto/sha256"
	"fmt"

	"bytes"
	"text/template"

	"encoding/json"

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
{{- "\t" }}{{"\t"}}{{- $target.Result }} [{{ $ID }}] {{ $target.Name -}}({{- $target.Kind -}}){{"\n"}}
{{- end }}
`
	// SOURCEREPORTTEMPLATE ...
	SOURCEREPORTTEMPLATE string = `
{{- "\t"}}Source:
{{ range $ID,$source := .Sources }}
{{- "\t" }}{{"\t"}}{{- $source.Result }} [{{ $ID }}] {{ $source.Name -}}({{- $source.Kind -}}){{"\n"}}
{{- end }}
`

	// REPORTTEMPLATE ...
	REPORTTEMPLATE string = `
=============================

REPORTS:

{{ if .Err }}
{{- .Result }} {{ .Name -}}{{"\n"}}
{{ "\t"}}Error: {{ .Err}}
{{ else }}
{{- .Result }} {{ .Name -}}{{"\n"}}
{{- if .reportURL }}
Report available on {{ .reportURL -}}{{"\n"}}
{{- end }}
{{- "\t"}}Source:
{{ range $ID, $source := .Sources }}
{{- "\t" }}{{"\t"}}{{- $source.Result }} [{{ $ID }}] {{ $source.Name }} (kind: {{ $source.Kind -}}){{"\n"}}
{{- end }}

{{- if .Conditions -}}
{{- "\t" }}Condition:
{{ range $ID, $condition := .Conditions }}
{{- "\t" }}{{"\t"}}{{- $condition.Result }} [{{ $ID }}] {{ $condition.Name }} (kind: {{ $condition.Kind -}}){{"\n"}}
{{- end -}}
{{- end -}}

{{- "\t" -}}Target:
{{ range $ID,$target := .Targets }}
{{- "\t" }}{{"\t"}}{{- $target.Result }} [{{ $ID}}] {{ $target.Name -}} (kind: {{ $target.Kind -}}){{"\n"}}
{{- end }}
{{ end }}
`
)

// Report contains a list of Rules
type Report struct {
	Name   string
	Err    string
	Result string
	// ID defines the report ID
	ID string
	// PipelineID represents the Updatecli manifest pipelineID
	PipelineID string
	Actions    map[string]*Action
	Sources    map[string]*result.Source
	Conditions map[string]*result.Condition
	Targets    map[string]*result.Target
	ReportURL  string
}

// String returns a report as a string
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

func (r *Report) UpdateID() error {
	var err error

	r.ID, err = getSha256HashFromStruct(*r)
	if err != nil {
		return err
	}

	for i, action := range r.Actions {
		action.ID, err = getSha256HashFromStruct(action)
		if err != nil {
			return err
		}

		r.Actions[i] = action
	}

	for i, condition := range r.Conditions {
		condition.ID, err = getSha256HashFromStruct(condition)
		if err != nil {
			return err
		}

		/*
			Always generate a SCM Id even if the scm is empty.
			I think this information could be useful to quickly identify this scenario
			That being said, I may revisit this decision in the future
		*/
		condition.Scm.ID, err = getSha256HashFromStruct(condition.Scm)
		if err != nil {
			return err
		}

		r.Conditions[i] = condition
	}

	for i, source := range r.Sources {
		source.ID, err = getSha256HashFromStruct(source)
		if err != nil {
			return err
		}

		/*
			Always generate a SCM Id even if the scm is empty.
			I think this information could be useful to quickly identify this scenario
			That being said, I may revisit this decision in the future
		*/
		source.Scm.ID, err = getSha256HashFromStruct(source.Scm)
		if err != nil {
			return err
		}

		r.Sources[i] = source
	}

	for i, target := range r.Targets {
		target.ID, err = getSha256HashFromStruct(target)
		if err != nil {
			return err
		}

		/*
			Always generate a SCM Id even if the scm is empty.
			I think this information could be useful to quickly identify this scenario
			That being said, I may revisit this decision in the future
		*/
		target.Scm.ID, err = getSha256HashFromStruct(target.Scm)
		if err != nil {
			return err
		}

		r.Targets[i] = target
	}

	return nil
}

func getSha256HashFromStruct(input interface{}) (string, error) {

	data, err := json.Marshal(input)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", sha256.Sum256(data)), nil
}
