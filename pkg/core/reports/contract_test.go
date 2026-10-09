package reports

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/updatecli/updatecli/pkg/core/result"
)

var updateGolden = flag.Bool("update", false, "update golden files")

// TestReportJSONContract pins the JSON document published to Udash.
//
// Udash decodes this payload, so any change to the golden file is a change to
// the Udash wire format and must be reviewed as such. Regenerate it with:
//
//	go test ./pkg/core/reports/ -run TestReportJSONContract -update
func TestReportJSONContract(t *testing.T) {
	r := Report{
		Name:       "Bump example",
		Labels:     map[string]string{"dependency.type": "golang-module"},
		Err:        "",
		Graph:      "",
		Result:     result.ATTENTION,
		ID:         "report-id",
		PipelineID: "pipeline-id",
		Actions: map[string]*Action{
			"default": {
				ID:          "action-id",
				Title:       "Bump example",
				Description: "description",
				Targets: []ActionTarget{
					{ID: "target-id", Title: "target", Description: "description"},
				},
				PipelineURL: &PipelineURL{URL: "https://ci.example.com/job/1", Name: "ci"},
				Link:        "https://github.com/example/example/pull/1",
			},
		},
		Sources: map[string]*result.Source{
			"default": {
				Name:        "Get latest version",
				Result:      result.SUCCESS,
				Information: "1.2.3",
				ID:          "default",
				Config:      map[string]string{"kind": "golang/module"},
			},
		},
		Conditions: map[string]*result.Condition{
			"default": {Name: "Check module", Result: result.SUCCESS, Pass: true},
		},
		Targets: map[string]*result.Target{
			"default": {
				Name:           "Update go.mod",
				Result:         result.ATTENTION,
				Information:    "1.2.2",
				NewInformation: "1.2.3",
				Files:          []string{"go.mod"},
			},
		},
		ReportURL: "",
		CI:        &CIData{URL: "https://ci.example.com/job/1", Name: "ci"},
		// Left empty so the golden file does not change with every release.
		UpdatecliVersion: "",
	}

	// Encoded the same way udash.Publish does.
	got, err := json.MarshalIndent(r, "", "  ")
	require.NoError(t, err)

	golden := filepath.Join("testdata", "udash_report.golden.json")

	if *updateGolden {
		require.NoError(t, os.WriteFile(golden, got, 0o600))
	}

	expected, err := os.ReadFile(golden)
	require.NoError(t, err, "run with -update to create the golden file")

	assert.JSONEq(t, string(expected), string(got))
}

func TestReportStringShowsReportURL(t *testing.T) {
	r := Report{
		Name:      "pipeline",
		Result:    result.SUCCESS,
		ReportURL: "https://udash.example.com/pipeline/reports/abc",
	}

	got, err := r.String("all")
	require.NoError(t, err)
	assert.Contains(t, got, "Report available on https://udash.example.com/pipeline/reports/abc")
}
