package reports

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func newIDTestReport(targetName string) Report {
	return Report{
		Name: "Update Golang version",
		Sources: map[string]*result.Source{
			"go": {Config: map[string]any{"kind": "golang"}},
		},
		Conditions: map[string]*result.Condition{
			"docker": {Config: map[string]any{"kind": "dockerimage", "spec": map[string]any{"image": "golang"}}},
		},
		Targets: map[string]*result.Target{
			"gomod": {
				Config: map[string]any{"name": targetName, "kind": "golang/gomod"},
				Scm: result.SCM{
					URL:    "https://github.com/updatecli/updatecli.git",
					Branch: result.GitBranch{Source: "main", Working: "updatecli_main_golang", Target: "main"},
				},
			},
			"precommit": {Config: map[string]any{"name": targetName, "kind": "yaml"}},
		},
	}
}

// legacyReportID is the report ID as UpdateID computed it before FreezeID existed.
func legacyReportID(t *testing.T, r Report) string {
	hash := func(input any) string {
		data, err := json.Marshal(input)
		require.NoError(t, err)
		return fmt.Sprintf("%x", sha256.Sum256(data))
	}

	reportHash := []string{r.Name}
	for _, id := range slices.Sorted(maps.Keys(r.Conditions)) {
		reportHash = append(reportHash, hash(r.Conditions[id].Config), hash(r.Conditions[id].Scm))
	}
	for _, id := range slices.Sorted(maps.Keys(r.Sources)) {
		reportHash = append(reportHash, hash(r.Sources[id].Config), hash(r.Sources[id].Scm))
	}
	for _, id := range slices.Sorted(maps.Keys(r.Targets)) {
		reportHash = append(reportHash, hash(r.Targets[id].Config), hash(r.Targets[id].Scm))
	}

	return fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(reportHash, "0"))))
}

// TestUpdateIDWithoutFreezeKeepsLegacyID ensures reports of manifests that never freeze
// their ID keep the ID they had, so their history isn't split.
func TestUpdateIDWithoutFreezeKeepsLegacyID(t *testing.T) {
	r := newIDTestReport("deps(go): update Go version to 1.25.2")
	expected := legacyReportID(t, newIDTestReport("deps(go): update Go version to 1.25.2"))

	require.NoError(t, r.UpdateID())
	assert.Equal(t, expected, r.ID)

	// Running it twice must not change anything, even though the scm IDs are now set.
	require.NoError(t, r.UpdateID())
	assert.Equal(t, expected, r.ID)
}

// TestUpdateIDUsesFrozenID ensures values rendered at runtime, such as {{ source "go" }},
// don't change the report ID while the resource IDs still reflect the rendered configuration.
func TestUpdateIDUsesFrozenID(t *testing.T) {
	unrendered := newIDTestReport(`deps(go): update Go version to {{ source "go" }}`)
	require.NoError(t, unrendered.FreezeID())
	frozenID := unrendered.stableID

	var reportIDs, targetIDs []string
	for _, version := range []string{"1.25.2", "1.25.3"} {
		r := newIDTestReport(`deps(go): update Go version to {{ source "go" }}`)
		require.NoError(t, r.FreezeID())

		// Rendering the configuration replaces the target configuration
		r.Targets["gomod"].Config = map[string]any{"name": "deps(go): update Go version to " + version, "kind": "golang/gomod"}
		r.Targets["gomod"].Scm.BranchReset = true

		require.NoError(t, r.UpdateID())
		reportIDs = append(reportIDs, r.ID)
		targetIDs = append(targetIDs, r.Targets["gomod"].ID)
	}

	assert.Equal(t, []string{frozenID, frozenID}, reportIDs)
	assert.NotEqual(t, targetIDs[0], targetIDs[1], "resource IDs must reflect the rendered configuration")
}

// TestFreezeIDMatchesLegacyIDWithoutTemplates ensures a manifest without runtime values
// keeps the report ID it had before the ID was frozen.
func TestFreezeIDMatchesLegacyIDWithoutTemplates(t *testing.T) {
	r := newIDTestReport("deps(go): update Go version")
	require.NoError(t, r.FreezeID())
	require.NoError(t, r.UpdateID())

	assert.Equal(t, legacyReportID(t, newIDTestReport("deps(go): update Go version")), r.ID)
}
