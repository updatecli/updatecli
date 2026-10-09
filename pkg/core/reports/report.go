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
	"github.com/updatecli/updatecli/pkg/plugins/utils/ci"
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

// CIData contains information about the CI pipeline where the report is generated
type CIData struct {
	URL  string `json:",omitempty"`
	Name string `json:",omitempty"`
}

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
	CI         *CIData
	// UpdatecliVersion is the version of Updatecli which produced the report
	UpdatecliVersion string
	// stableID holds the report ID frozen by FreezeID, before the configuration was rendered
	stableID string
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

// UpdateCIJob updates the report with CI job information if available
func (r *Report) UpdateCIJob() error {
	detectedCi, err := ci.New()
	if err != nil {
		return err
	}

	if detectedCi == nil {
		// No CI pipeline detected
		return nil
	}

	r.CI = &CIData{
		Name: detectedCi.Name(),
		URL:  detectedCi.URL(),
	}
	return nil
}

// UpdateID sets the report ID along with the ID of every resource and of its scm.
//
// The report ID identifies the pipeline manifest the report comes from, regardless of the
// result of the pipeline or of the values rendered at runtime, such as {{ source "id" }}.
// It's different from the pipelineID which is used to identify an update scenario which
// could be the result of multiple Updatecli manifests.
//
// The resource IDs are computed from the configuration as it was rendered, while the report
// ID is the one frozen by FreezeID when it was called.
func (r *Report) UpdateID() error {
	resourceIDs, err := r.resourceIDs()
	if err != nil {
		return err
	}

	for id, condition := range r.Conditions {
		condition.ID = resourceIDs[conditionKey(id)].config
		condition.Scm.ID = resourceIDs[conditionKey(id)].scm
	}

	for id, source := range r.Sources {
		source.ID = resourceIDs[sourceKey(id)].config
		source.Scm.ID = resourceIDs[sourceKey(id)].scm
	}

	for id, target := range r.Targets {
		target.ID = resourceIDs[targetKey(id)].config
		target.Scm.ID = resourceIDs[targetKey(id)].scm
	}

	if r.stableID != "" {
		r.ID = r.stableID
		return nil
	}

	r.ID, err = r.computeID()
	return err
}

// FreezeID computes the report ID from the report as it is now, and keeps it for UpdateID.
//
// It is meant to be called before the configuration is rendered with the values only known
// at runtime, such as {{ source "id" }}. Otherwise a resource named after the version it
// updates would get a new report ID every time that version changes.
func (r *Report) FreezeID() error {
	id, err := r.computeID()
	if err != nil {
		return err
	}

	r.stableID = id
	return nil
}

// resourceHashes contains the hash of a resource configuration and the hash of its scm.
type resourceHashes struct {
	config string
	scm    string
}

func conditionKey(id string) string { return "condition#" + id }
func sourceKey(id string) string    { return "source#" + id }
func targetKey(id string) string    { return "target#" + id }

// resourceIDs returns the hashes of every resource of the report, without modifying it.
func (r *Report) resourceIDs() (map[string]resourceHashes, error) {
	hashes := make(map[string]resourceHashes, len(r.Conditions)+len(r.Sources)+len(r.Targets))

	hash := func(key string, config any, scm result.SCM) error {
		configID, err := getSha256HashFromStruct(config)
		if err != nil {
			return err
		}

		// The scm ID is itself derived from the scm, so it is cleared first to keep the
		// hash the same whether or not it was already set.
		scm.ID = ""
		/*
			Always generate a SCM Id even if the scm is empty.
			I think this information could be useful to quickly identify this scenario
			That being said, I may revisit this decision in the future
		*/
		scmID, err := getSha256HashFromStruct(scm)
		if err != nil {
			return err
		}

		hashes[key] = resourceHashes{config: configID, scm: scmID}
		return nil
	}

	for id, condition := range r.Conditions {
		if err := hash(conditionKey(id), condition.Config, condition.Scm); err != nil {
			return nil, err
		}
	}

	for id, source := range r.Sources {
		if err := hash(sourceKey(id), source.Config, source.Scm); err != nil {
			return nil, err
		}
	}

	for id, target := range r.Targets {
		if err := hash(targetKey(id), target.Config, target.Scm); err != nil {
			return nil, err
		}
	}

	return hashes, nil
}

// computeID returns the report ID derived from the pipeline name and from the configuration
// of every resource and of its scm, without modifying the report.
func (r *Report) computeID() (string, error) {
	resourceIDs, err := r.resourceIDs()
	if err != nil {
		return "", err
	}

	reportHash := []string{}

	if r.Name != "" {
		reportHash = append(reportHash, r.Name)
	}

	// Resources are sorted by their ID to make sure that the hash is always the same
	for _, id := range slices.Sorted(maps.Keys(r.Conditions)) {
		reportHash = append(reportHash, resourceIDs[conditionKey(id)].config, resourceIDs[conditionKey(id)].scm)
	}

	for _, id := range slices.Sorted(maps.Keys(r.Sources)) {
		reportHash = append(reportHash, resourceIDs[sourceKey(id)].config, resourceIDs[sourceKey(id)].scm)
	}

	for _, id := range slices.Sorted(maps.Keys(r.Targets)) {
		reportHash = append(reportHash, resourceIDs[targetKey(id)].config, resourceIDs[targetKey(id)].scm)
	}

	return fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(reportHash, "0")))), nil
}

func getSha256HashFromStruct(input interface{}) (string, error) {

	data, err := json.Marshal(input)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", sha256.Sum256(data)), nil
}
