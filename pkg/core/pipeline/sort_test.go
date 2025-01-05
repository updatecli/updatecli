package pipeline

import (
	"sort"
	"testing"

	"github.com/heimdalr/dag"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/condition"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/source"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
	"github.com/updatecli/updatecli/pkg/plugins/resources/file"
	"github.com/updatecli/updatecli/pkg/plugins/resources/shell"
)

type ResultLeaf struct {
	Id      string
	Parents []string
}

func TestSortedResourcesKeys(t *testing.T) {
	testdata := []struct {
		Name           string
		Conf           Pipeline
		Sources        map[string]source.Config
		Conditions     map[string]condition.Config
		Targets        map[string]target.Config
		ExpectedResult [][]ResultLeaf
		ExpectedErr    error
	}{{
		Name: "Scenario 1",
		Sources: map[string]source.Config{
			"1": {
				ResourceConfig: resource.ResourceConfig{
					Kind: "shell",
					DependsOn: []string{
						"2",
						"3",
					},
				},
			},
			"2": {
				ResourceConfig: resource.ResourceConfig{
					Kind: "shell",
					DependsOn: []string{
						"3",
					},
				},
			},
			"3": {
				ResourceConfig: resource.ResourceConfig{
					Kind: "shell",
				},
			},
		},
		Conditions: map[string]condition.Config{
			"1": {
				ResourceConfig: resource.ResourceConfig{
					Kind: "shell",
					DependsOn: []string{
						"2",
					},
				},
				DisableSourceInput: true,
			},
			"2": {
				ResourceConfig: resource.ResourceConfig{
					Kind: "shell",
					DependsOn: []string{
						"3",
					},
				},
				DisableSourceInput: true,
			},
			"3": {
				ResourceConfig: resource.ResourceConfig{
					Kind: "shell",
				},
				DisableSourceInput: true,
			},
		},
		Targets: map[string]target.Config{
			"1": {
				ResourceConfig: resource.ResourceConfig{
					Kind: "shell",
					DependsOn: []string{
						"2",
					},
				},
				DisableConditions:  true,
				DisableSourceInput: true,
			},
			"2": {
				ResourceConfig: resource.ResourceConfig{
					Kind: "shell",
					DependsOn: []string{
						"3",
					},
				},
				DisableConditions:  true,
				DisableSourceInput: true,
			},
			"3": {
				ResourceConfig: resource.ResourceConfig{
					Kind: "shell",
				},
				DisableConditions:  true,
				DisableSourceInput: true,
			},
		},
		ExpectedResult: [][]ResultLeaf{
			{
				{Id: "source#3", Parents: []string{"root"}},
				{Id: "condition#3", Parents: []string{"root"}},
				{Id: "target#3", Parents: []string{"root"}},
			},
			{
				{Id: "source#2", Parents: []string{"root", "source#3"}},
				{Id: "condition#2", Parents: []string{"root", "condition#3"}},
				{Id: "target#2", Parents: []string{"root", "target#3"}},
			},
			{
				{Id: "source#1", Parents: []string{"root", "source#2", "source#3"}},
				{Id: "condition#1", Parents: []string{"root", "condition#2"}},
				{Id: "target#1", Parents: []string{"root", "target#2"}},
			},
		},
	},
		{
			Name: "Scenario 2",
			Sources: map[string]source.Config{
				"1": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
						DependsOn: []string{
							"3",
						},
					},
				},
				"2": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
						DependsOn: []string{
							"4",
						},
					},
				},
				"3": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
						DependsOn: []string{
							"2",
						},
					},
				},
				"4": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
					},
				},
			},
			Conditions: map[string]condition.Config{
				"1": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
						DependsOn: []string{
							"3",
						},
					},
					DisableSourceInput: true,
				},
				"2": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
						DependsOn: []string{
							"4",
						},
					},
					DisableSourceInput: true,
				},
				"3": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
						DependsOn: []string{
							"2",
						},
					},
					DisableSourceInput: true,
				},
				"4": {

					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
					},
					DisableSourceInput: true,
				},
			},
			Targets: map[string]target.Config{
				"1": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
						DependsOn: []string{
							"3",
						},
					},
					DisableSourceInput: true,
					DisableConditions:  true,
				},
				"2": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
						DependsOn: []string{
							"4",
						},
					},
					DisableSourceInput: true,
					DisableConditions:  true,
				},
				"3": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
						DependsOn: []string{
							"2",
						},
					},
					DisableSourceInput: true,
					DisableConditions:  true,
				},
				"4": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
					},
					DisableConditions:  true,
					DisableSourceInput: true,
				},
			},
			ExpectedResult: [][]ResultLeaf{
				{
					{Id: "source#4", Parents: []string{"root"}},
					{Id: "condition#4", Parents: []string{"root"}},
					{Id: "target#4", Parents: []string{"root"}},
				},
				{
					{Id: "source#2", Parents: []string{"root", "source#4"}},
					{Id: "condition#2", Parents: []string{"root", "condition#4"}},
					{Id: "target#2", Parents: []string{"root", "target#4"}},
				},
				{
					{Id: "source#3", Parents: []string{"root", "source#2"}},
					{Id: "condition#3", Parents: []string{"root", "condition#2"}},
					{Id: "target#3", Parents: []string{"root", "target#2"}},
				},
				{
					{Id: "source#1", Parents: []string{"root", "source#3"}},
					{Id: "condition#1", Parents: []string{"root", "condition#3"}},
					{Id: "target#1", Parents: []string{"root", "target#3"}},
				},
			},
		},
		{
			Name: "Scenario 3",
			Sources: map[string]source.Config{
				"1": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
						DependsOn: []string{
							"2",
						},
					},
				},
			},
			Conditions: map[string]condition.Config{
				"2": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
						DependsOn: []string{
							"3",
						},
					},
					DisableSourceInput: true,
				},
			},
			Targets: map[string]target.Config{
				"3": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
						DependsOn: []string{
							"4",
						},
					},
					DisableSourceInput: true,
					DisableConditions:  true,
				},
			},
			ExpectedResult: [][]ResultLeaf{},
			ExpectedErr:    ErrNotValidDependsOn,
		},
		{
			Name: "Scenario 4",
			Sources: map[string]source.Config{
				"1": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
						DependsOn: []string{
							"target#2",
						},
					},
				},
				"2": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
						DependsOn: []string{
							"condition#3",
							"condition#2",
						},
					},
				},
				"3": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
						DependsOn: []string{
							"target#4",
						},
					},
				},
				"4": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
						DependsOn: []string{
							"target#4",
						},
					},
				},
			},
			Conditions: map[string]condition.Config{
				"1": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
						DependsOn: []string{
							"condition#2",
						},
					},
					DisableSourceInput: true,
				},
				"2": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
						DependsOn: []string{
							"condition#4",
						},
					},
					DisableSourceInput: true,
				},
				"3": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
						DependsOn: []string{
							"source#4",
							"source#3",
						},
					},
					DisableSourceInput: true,
				},
				"4": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
						DependsOn: []string{
							"target#4",
						},
					},
					DisableSourceInput: true,
				},
			},
			Targets: map[string]target.Config{
				"1": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
						DependsOn: []string{
							"condition#1",
						},
					},
					DisableConditions:  true,
					DisableSourceInput: true,
				},
				"2": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
						DependsOn: []string{
							"target#1",
							"target#3",
						},
					},
					DisableConditions:  true,
					DisableSourceInput: true,
				},
				"3": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
						DependsOn: []string{
							"source#2",
						},
					},
					DisableConditions:  true,
					DisableSourceInput: true,
				},
				"4": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
					},
					DisableConditions:  true,
					DisableSourceInput: true,
				},
			},
			ExpectedResult: [][]ResultLeaf{
				{
					{Id: "target#4", Parents: []string{"root"}},
				},
				{
					{Id: "source#4", Parents: []string{"root", "target#4"}},
					{Id: "source#3", Parents: []string{"root", "target#4"}},
					{Id: "condition#4", Parents: []string{"root", "target#4"}},
				},
				{
					{Id: "condition#3", Parents: []string{"root", "source#3", "source#4"}},
					{Id: "condition#2", Parents: []string{"root", "condition#4"}},
				},
				{
					{Id: "source#2", Parents: []string{"root", "condition#2", "condition#3"}},
					{Id: "condition#1", Parents: []string{"root", "condition#2"}},
				},
				{
					{Id: "target#3", Parents: []string{"root", "source#2"}},
					{Id: "target#1", Parents: []string{"root", "condition#1"}},
				},
				{
					{Id: "target#2", Parents: []string{"root", "target#3", "target#1"}},
				},
				{
					{Id: "source#1", Parents: []string{"root", "target#2"}},
				},
			},
		},
		{
			Name: "Scenario 5",
			Sources: map[string]source.Config{
				"1": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
						DependsOn: []string{
							"2",
						},
					},
				},
				"2": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
						DependsOn: []string{
							"1",
						},
					},
				},
			},
			Conditions: map[string]condition.Config{
				"1": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
						DependsOn: []string{
							"2",
						},
					},
					DisableSourceInput: true,
				},
				"2": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
						DependsOn: []string{
							"1",
						},
					},
					DisableSourceInput: true,
				},
			},
			Targets: map[string]target.Config{
				"1": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
						DependsOn: []string{
							"2",
						},
					},
					DisableConditions:  true,
					DisableSourceInput: true,
				},
				"2": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
						DependsOn: []string{
							"1",
						},
					},
					DisableConditions:  true,
					DisableSourceInput: true,
				},
			},
			ExpectedResult: [][]ResultLeaf{},
			ExpectedErr:    ErrDependsOnLoopDetected,
		},
		{
			Name: "Scenario 6: Target Without all condition",
			Conditions: map[string]condition.Config{
				"1": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
					},
					DisableSourceInput: true,
				},
				"2": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
					},
					DisableSourceInput: true,
				},
				"3": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
					},
					DisableSourceInput: true,
				},
				"4": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
					},
					DisableSourceInput: true,
				},
			},
			Targets: map[string]target.Config{
				"1": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
					},
					DisableSourceInput: true,
				},
			},
			ExpectedResult: [][]ResultLeaf{
				{
					{Id: "condition#1", Parents: []string{"root"}},
					{Id: "condition#2", Parents: []string{"root"}},
					{Id: "condition#3", Parents: []string{"root"}},
					{Id: "condition#4", Parents: []string{"root"}},
				},
				{
					{Id: "target#1", Parents: []string{"root", "condition#1", "condition#2", "condition#3", "condition#4"}},
				},
			},
		},
		{
			Name: "Scenario 7: Target With deprecated condition ids",
			Conditions: map[string]condition.Config{
				"1": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
						DependsOn: []string{
							"2",
						},
					},
					DisableSourceInput: true,
				},
				"2": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
					},
					DisableSourceInput: true,
				},
				"3": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
					},
					DisableSourceInput: true,
				},
				"4": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
						DependsOn: []string{
							"2",
						},
					},
					DisableSourceInput: true,
				},
			},
			Targets: map[string]target.Config{
				"1": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
					},
					DeprecatedConditionIDs: []string{"1", "4"},
					DisableSourceInput:     true,
				},
			},
			ExpectedResult: [][]ResultLeaf{
				{
					{Id: "condition#2", Parents: []string{"root"}},
					{Id: "condition#3", Parents: []string{"root"}},
				},
				{
					{Id: "condition#1", Parents: []string{"root", "condition#2"}},
					{Id: "condition#4", Parents: []string{"root", "condition#2"}},
				},
				{
					{Id: "target#1", Parents: []string{"root", "condition#1", "condition#4"}},
				},
			},
		},
		{
			Name: "Scenario 8: Source Id creates an inferred deps",
			Sources: map[string]source.Config{
				"1": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
						DependsOn: []string{
							"condition#1",
						},
					},
				},
				"2": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
						DependsOn: []string{
							"1",
						},
					},
				},
			},
			Conditions: map[string]condition.Config{
				"1": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
					},
					DisableSourceInput: true,
				},
				"2": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
					},
					SourceID: "2",
				},
			},
			Targets: map[string]target.Config{
				"1": {
					SourceID:          "2",
					DisableConditions: true,
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
					},
				},
			},
			ExpectedResult: [][]ResultLeaf{
				{
					{Id: "condition#1", Parents: []string{"root"}},
				},
				{
					{Id: "source#1", Parents: []string{"root", "condition#1"}},
				},
				{
					{Id: "source#2", Parents: []string{"root", "source#1"}},
				},
				{
					{Id: "condition#2", Parents: []string{"root", "source#2"}},
					{Id: "target#1", Parents: []string{"root", "source#2"}},
				},
			},
		},
		{
			Name: "Scenario 9: DependsOnChange",
			Sources: map[string]source.Config{
				"1": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
					},
				},
			},
			Conditions: map[string]condition.Config{
				"1": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
					},
					DisableSourceInput: true,
				},
			},
			Targets: map[string]target.Config{
				"1": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
					},
					DisableSourceInput: true,
				},
				"5": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
						DependsOn: []string{
							"1",
						},
					},
					DisableSourceInput: true,
				},
				"6": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
						DependsOn: []string{
							"1",
						},
					},
					DependsOnChange:    true,
					DisableSourceInput: true,
				},
				"7": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
						DependsOn: []string{
							"5",
						},
					},
					DependsOnChange:    true,
					DisableSourceInput: true,
				},
			},
			ExpectedResult: [][]ResultLeaf{
				{
					{Id: "source#1", Parents: []string{"root"}},
					{Id: "condition#1", Parents: []string{"root"}},
				},
				{
					{Id: "target#1", Parents: []string{"root", "condition#1"}},
				},
				{
					{Id: "target#5", Parents: []string{"root", "condition#1", "target#1"}},
					{Id: "target#6", Parents: []string{"root", "condition#1", "target#1"}},
				},
				{
					{Id: "target#7", Parents: []string{"root", "condition#1", "target#5"}},
				},
			},
		},
		{
			Name: "Scenario 10: Runtime Dependency",
			Sources: map[string]source.Config{
				"1": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
					},
				},
				"2": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
					},
				},
				"3": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
					},
				},
				"4": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
					},
				},
			},
			Conditions: map[string]condition.Config{
				"1": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "file",
						Spec: file.Spec{
							Files: []string{
								"{{ source \"1\" }}",
								"{{ source \"2\" }}",
								"{{ source \"3\" }}",
								"{{ source \"4\" }}",
							},
						},
					},
					DisableSourceInput: true,
				},
			},
			ExpectedResult: [][]ResultLeaf{
				{
					{Id: "source#1", Parents: []string{"root"}},
					{Id: "source#2", Parents: []string{"root"}},
					{Id: "source#3", Parents: []string{"root"}},
					{Id: "source#4", Parents: []string{"root"}},
				},
				{
					{Id: "condition#1", Parents: []string{"root", "source#1", "source#2", "source#3", "source#4"}},
				},
			},
		},
		{
			Name: "Scenario 11: SourceID and Explicit Dep",
			Sources: map[string]source.Config{
				"1": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
					},
				},
			},
			Conditions: map[string]condition.Config{
				"1": {
					ResourceConfig: resource.ResourceConfig{
						Kind:      "shell",
						DependsOn: []string{"source#1"},
					},
					SourceID: "1",
				},
			},
			ExpectedResult: [][]ResultLeaf{
				{
					{Id: "source#1", Parents: []string{"root"}},
				},
				{
					{Id: "condition#1", Parents: []string{"root", "source#1"}},
				},
			},
		},
		{
			Name: "Scenario 12: Runtime Dependency in sources",
			Sources: map[string]source.Config{
				"1": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
					},
				},
				"2": {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
						Spec: shell.Spec{
							Command: "echo {{ source \"1\"}}",
						},
					},
				},
			},
			ExpectedResult: [][]ResultLeaf{
				{
					{Id: "source#1", Parents: []string{"root"}},
				},
				{
					{Id: "source#2", Parents: []string{"root", "source#1"}},
				},
			},
		},
	}

	for i := range testdata {
		data := &testdata[i]
		t.Run(data.Name, func(t *testing.T) {
			sources := map[string]source.Source{}
			for k, v := range data.Sources {
				err := v.Validate()
				if err != nil {
					logrus.Errorf("Failed to validate source config %s", err)
				}
				sources[k] = source.Source{
					Config: v,
				}
			}
			conditions := map[string]condition.Condition{}
			for k, v := range data.Conditions {
				err := v.Validate()
				if err != nil {
					logrus.Errorf("Failed to validate source config %s", err)
				}
				conditions[k] = condition.Condition{
					Config: v,
				}
			}
			targets := map[string]target.Target{}
			for k, v := range data.Targets {
				err := v.Validate()
				if err != nil {
					logrus.Errorf("Failed to validate source config %s", err)
				}
				targets[k] = target.Target{
					Config: v,
				}
			}
			p := Pipeline{
				Sources:    sources,
				Conditions: conditions,
				Targets:    targets,
				Config: &config.Config{
					Spec: config.Spec{
						Sources:    data.Sources,
						Conditions: data.Conditions,
						Targets:    data.Targets,
					},
				},
			}
			// _ = p.Update()
			gotSortedDag, err := p.SortedResources()

			require.Equal(t, data.ExpectedErr, err)

			if gotSortedDag == nil {
				return
			}
			results, _ := gotSortedDag.GetOrderedDescendants(rootVertex)
			compareDag(t, data.ExpectedResult, results, gotSortedDag)
		})
	}
}

// compareDag allows to test that the resulted orderedDescendant of a dag
// match the expected result.
// As the sibling order is not guarantee, we need to  handle that
func compareDag(t *testing.T, expected [][]ResultLeaf, got []string, d *dag.DAG) {
	//{ Keep track of the current index in 'got'
	index := 0
	// Iterate over each sublist in 'expected'
	for _, sublist := range expected {
		// If remaining 'got' is smaller than current 'expected' sublist, return false
		require.GreaterOrEqual(t, len(got), index+len(sublist))

		// Extract Ids from sublist
		sublistStr := []string{}
		for _, i := range sublist {
			sublistStr = append(sublistStr, i.Id)
		}
		// Extract the slice from 'got' to compare with current sublist
		gotSublist := got[index : index+len(sublist)]

		// Sort both the current 'expected' sublist and the corresponding 'got' sublist
		sort.Strings(sublistStr)
		sort.Strings(gotSublist)

		// require.Equal(t, sublistStr, gotSublist)
		// Check if they have the required dependencies
		for _, l := range sublist {
			parents, _ := d.GetParents(l.Id)
			parentIds := []string{}
			for k := range parents {
				parentIds = append(parentIds, k)
			}
			expectedParents := []string{}
			expectedParents = append(expectedParents, l.Parents...)
			sort.Strings(expectedParents)
			sort.Strings(parentIds)
			require.Equal(t, expectedParents, parentIds)
		}

		// Move the index forward by the length of the current sublist
		index += len(sublist)
	}

	// If we've processed all 'expected' sublists and matched them correctly, return true
	require.Equal(t, index, len(got))
}
