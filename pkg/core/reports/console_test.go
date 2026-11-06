package reports

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/updatecli/updatecli/pkg/core/result"
)

// TestRender pins the console summary printed at the end of a run.
// Regenerate the golden files with:
//
//	go test ./pkg/core/reports/ -run TestRender -update
func TestRender(t *testing.T) {
	scm := result.SCM{
		URL:    "https://github.com/updatecli/updatecli.git",
		Branch: result.GitBranch{Target: "updatecli_main_abc"},
	}

	testdata := []struct {
		name    string
		reports Reports
	}{
		{
			name: "no_action",
			reports: Reports{
				{
					Name:   "Bump golang",
					Result: result.SUCCESS,
					Sources: map[string]*result.Source{
						"latest": {Name: "Get latest golang", Result: result.SUCCESS},
						"alpha":  {Name: "Get alpha", Result: result.SUCCESS, Scm: scm},
					},
					Conditions: map[string]*result.Condition{
						"exists": {Name: "Check image", Result: result.SUCCESS},
					},
					Targets: map[string]*result.Target{
						"gomod": {Name: "Update go.mod", Result: result.SUCCESS, Scm: scm},
					},
				},
			},
		},
		{
			name: "error_and_report_url",
			reports: Reports{
				{
					Name:       "Broken pipeline",
					PipelineID: "broken",
					Result:     result.FAILURE,
					Err:        "something went wrong",
				},
				{
					Name:      "Reported pipeline",
					Result:    result.ATTENTION,
					ReportURL: "https://udash.example.com/pipeline/reports/abc",
					Targets: map[string]*result.Target{
						"file": {Name: "Update file", Result: result.ATTENTION},
					},
				},
			},
		},
		{
			name: "actions",
			reports: Reports{
				{
					Name:   "Bump golang",
					Result: result.ATTENTION,
					Targets: map[string]*result.Target{
						"gomod": {Name: "Update go.mod", Result: result.ATTENTION, Scm: scm},
					},
					Actions: map[string]*Action{
						"default": {
							Title: "deps: bump golang to 1.27",
							Link:  "https://github.com/updatecli/updatecli/pull/123",
						},
						"dryrun": {
							PipelineTitle: "Bump golang",
							DryRun:        true,
						},
					},
				},
			},
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.reports.render()
			golden := filepath.Join("testdata", "console_"+tt.name+".golden")

			if *updateGolden {
				require.NoError(t, os.WriteFile(golden, []byte(got), 0o600))
			}

			expected, err := os.ReadFile(golden)
			require.NoError(t, err, "run with -update to create the golden file")
			assert.Equal(t, string(expected), got)
		})
	}
}
