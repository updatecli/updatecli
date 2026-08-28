package pipeline

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func TestExtractDepsFromTemplate(t *testing.T) {
	testdata := []struct {
		Name           string
		Template       string
		ExpectedResult []string
		ExpectedErr    string
	}{{
		Name: "Scenario 1",
		Template: `
{{ source "sourceId1" }}
		{{ condition "conditionid1" }}
		{{ target "targetid1" }}
		{{ source "sourceId2" }}
		{{ target "targetid2" }}
		{{ condition "conditionid2" }}
        `,
		ExpectedResult: []string{
			"source#sourceId1", "source#sourceId2",
			"condition#conditionid1", "condition#conditionid2",
			"target#targetid1", "target#targetid2",
		},
	},
		{
			// {{ pipeline "..." }} takes a path into the pipeline configuration, so
			// the dependency is the resource that path reads from.
			Name: "Pipeline queries rooted at a resource map",
			Template: `
{{ pipeline "sources.sourceId1.output" }}
		{{ pipeline "Conditions.conditionid1.Result" }}
		{{ pipeline "TARGETS.targetid1.Output" }}
        `,
			ExpectedResult: []string{
				"source#sourceId1", "condition#conditionid1", "target#targetid1",
			},
		},
		{
			// Any other query depends on nothing, and used to produce an
			// unresolvable `pipeline#...` dependency.
			Name: "Pipeline queries not rooted at a resource map",
			Template: `
{{ pipeline "name" }}
		{{ pipeline "scms.default.url" }}
		{{ pipeline "sources." }}
        `,
			ExpectedResult: []string{},
		},
	}

	for _, data := range testdata {

		t.Run(data.Name, func(t *testing.T) {
			got, err := ExtractDepsFromTemplate(data.Template)
			if data.ExpectedErr != "" {
				require.EqualError(t, err, data.ExpectedErr)
			} else {
				require.NoError(t, err)
			}
			gotIDs := []string{}
			for _, dependency := range got {
				gotIDs = append(gotIDs, dependency.ID)
			}

			sort.Strings(data.ExpectedResult)
			sort.Strings(gotIDs)
			require.Equal(t, data.ExpectedResult, gotIDs)
		})
	}
}

func TestShouldSkipResource(t *testing.T) {
	testdata := []struct {
		Name           string
		Leaf           Node
		DepsResults    map[string]*Node
		ExpectedResult bool
	}{
		{
			Name: "target depending on a successful source runs",
			Leaf: Node{
				ID:        "target#mytarget",
				Category:  targetCategory,
				DependsOn: []Dependency{{ID: "source#mysource", Operator: andBooleanOperator}},
			},
			DepsResults: map[string]*Node{
				"source#mysource": {ID: "source#mysource", Category: sourceCategory, Result: result.SUCCESS},
			},
			ExpectedResult: false,
		},
		{
			// A skipped source has no value to provide, so its dependents would otherwise
			// consume an empty source input.
			Name: "target depending on a skipped source is skipped",
			Leaf: Node{
				ID:        "target#mytarget",
				Category:  targetCategory,
				DependsOn: []Dependency{{ID: "source#mysource", Operator: andBooleanOperator}},
			},
			DepsResults: map[string]*Node{
				"source#mysource": {ID: "source#mysource", Category: sourceCategory, Result: result.SKIPPED},
			},
			ExpectedResult: true,
		},
		{
			Name: "condition depending on a skipped source is skipped",
			Leaf: Node{
				ID:        "condition#mycondition",
				Category:  conditionCategory,
				DependsOn: []Dependency{{ID: "source#mysource", Operator: andBooleanOperator}},
			},
			DepsResults: map[string]*Node{
				"source#mysource": {ID: "source#mysource", Category: sourceCategory, Result: result.SKIPPED},
			},
			ExpectedResult: true,
		},
		{
			Name: "target depending on a failed source is skipped",
			Leaf: Node{
				ID:        "target#mytarget",
				Category:  targetCategory,
				DependsOn: []Dependency{{ID: "source#mysource", Operator: andBooleanOperator}},
			},
			DepsResults: map[string]*Node{
				"source#mysource": {ID: "source#mysource", Category: sourceCategory, Result: result.FAILURE},
			},
			ExpectedResult: true,
		},
		{
			// A skipped target only means it had nothing to change, which mustn't stop
			// the resources depending on it.
			Name: "target depending on a skipped target runs",
			Leaf: Node{
				ID:        "target#mytarget",
				Category:  targetCategory,
				DependsOn: []Dependency{{ID: "target#myothertarget", Operator: andBooleanOperator}},
			},
			DepsResults: map[string]*Node{
				"target#myothertarget": {ID: "target#myothertarget", Category: targetCategory, Result: result.SKIPPED},
			},
			ExpectedResult: false,
		},
		{
			Name: "resource without any dependency runs",
			Leaf: Node{
				ID:       "target#mytarget",
				Category: targetCategory,
			},
			ExpectedResult: false,
		},
	}

	p := Pipeline{}

	for _, tt := range testdata {
		t.Run(tt.Name, func(t *testing.T) {
			require.Equal(t, tt.ExpectedResult, p.shouldSkipResource(&tt.Leaf, tt.DepsResults))
		})
	}
}
