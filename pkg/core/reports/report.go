package reports

import (
	"crypto/sha256"
	"fmt"
	"maps"
	"slices"
	"strings"

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

// Report contains the result of the execution of a pipeline
type Report struct {
	Name   string
	Labels map[string]string
	Err    string
	Graph  string
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

// UpdateID generates a unique ID for the report based on the content of the report
// Ideally the report ID should be the same regardless of the result of the pipeline.
// The goal is to be able to identify what report corresponds to what pipeline manifest.
// It's different from the pipelineID which is used to identify an update scenario which
// could be the result of multiple Updatecli manifests.
func (r *Report) UpdateID() error {
	var err error

	r.ID, err = getSha256HashFromStruct(*r)
	if err != nil {
		return err
	}

	reportHash := []string{}

	if r.Name != "" {
		reportHash = append(reportHash, r.Name)
	}

	// We need to sort the actions by their ID to make sure that the hash is always the same
	for _, i := range slices.Sorted(maps.Keys(r.Actions)) {
		action := r.Actions[i]
		action.ID, err = getSha256HashFromStruct(action)
		if err != nil {
			return err
		}

		reportHash = append(reportHash, action.ID)

		r.Actions[i] = action
	}

	// We need to sort the conditions by their ID to make sure that the hash is always the same
	for _, i := range slices.Sorted(maps.Keys(r.Conditions)) {
		condition := r.Conditions[i]
		condition.ID, err = getSha256HashFromStruct(condition.Config)
		if err != nil {
			return err
		}

		reportHash = append(reportHash, condition.ID)

		/*
			Always generate a SCM Id even if the scm is empty.
			I think this information could be useful to quickly identify this scenario
			That being said, I may revisit this decision in the future
		*/
		condition.Scm.ID, err = getSha256HashFromStruct(condition.Scm)
		if err != nil {
			return err
		}

		reportHash = append(reportHash, condition.Scm.ID)

		r.Conditions[i] = condition
	}

	// We need to sort the conditions by their ID to make sure that the hash is always the same
	for _, i := range slices.Sorted(maps.Keys(r.Sources)) {
		source := r.Sources[i]
		source.ID, err = getSha256HashFromStruct(source.Config)
		if err != nil {
			return err
		}

		reportHash = append(reportHash, source.ID)

		/*
			Always generate a SCM Id even if the scm is empty.
			I think this information could be useful to quickly identify this scenario
			That being said, I may revisit this decision in the future
		*/
		source.Scm.ID, err = getSha256HashFromStruct(source.Scm)
		if err != nil {
			return err
		}

		reportHash = append(reportHash, source.Scm.ID)
		r.Sources[i] = source
	}

	// We need to sort the conditions by their ID to make sure that the hash is always the same
	for _, i := range slices.Sorted(maps.Keys(r.Targets)) {
		target := r.Targets[i]
		target.ID, err = getSha256HashFromStruct(target.Config)
		if err != nil {
			return err
		}

		reportHash = append(reportHash, target.ID)

		/*
			Always generate a SCM Id even if the scm is empty.
			I think this information could be useful to quickly identify this scenario
			That being said, I may revisit this decision in the future
		*/
		target.Scm.ID, err = getSha256HashFromStruct(target.Scm)
		if err != nil {
			return err
		}

		reportHash = append(reportHash, target.Scm.ID)

		r.Targets[i] = target
	}

	r.ID = fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(reportHash, "0"))))

	// If the report doesn't have any configuration then we need to generate a hash based on the report itself
	// This is not ideal because it means that the report ID will be different each time Updatecli is executed
	// and that the result is different.
	if r.ID == "" {
		r.ID, err = getSha256HashFromStruct(*r)
		if err != nil {
			return err
		}
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
