package pipeline

import (
	"sort"
	"strings"
	"testing"

	"github.com/heimdalr/dag"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/pipeline/condition"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/source"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
	"github.com/updatecli/updatecli/pkg/plugins/resources/file"
)

type ResultLeaf struct {
	Id      string
	Parents []string
}

func TestSortedResourcesKeys(t *testing.T) {
	testdata := []struct {
		Name           string
		Conf           Pipeline
		ExpectedResult [][]ResultLeaf
		ExpectedErr    error
	}{{
		Name: "Scenario 1",
		Conf: Pipeline{
			Sources: map[string]source.Source{
				"1": {
					Config: source.Config{
						ResourceConfig: resource.ResourceConfig{
							DependsOn: []string{
								"2",
								"3",
							},
						},
					},
				},
				"2": {
					Config: source.Config{
						ResourceConfig: resource.ResourceConfig{
							DependsOn: []string{
								"3",
							},
						},
					},
				},
				"3": {},
			},
			Conditions: map[string]condition.Condition{
				"1": {
					Config: condition.Config{
						ResourceConfig: resource.ResourceConfig{
							DependsOn: []string{
								"2",
							},
						},
					},
				},
				"2": {
					Config: condition.Config{
						ResourceConfig: resource.ResourceConfig{
							DependsOn: []string{
								"3",
							},
						},
					},
				},
				"3": {},
			},
			Targets: map[string]target.Target{
				"1": {
					Config: target.Config{
						ResourceConfig: resource.ResourceConfig{
							DependsOn: []string{
								"2",
							},
						},
						DisableConditions: true,
					},
				},
				"2": {
					Config: target.Config{
						ResourceConfig: resource.ResourceConfig{
							DependsOn: []string{
								"3",
							},
						},
						DisableConditions: true,
					},
				},
				"3": {
					Config: target.Config{
						DisableConditions: true,
					},
				},
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
			Conf: Pipeline{
				Sources: map[string]source.Source{
					"1": {
						Config: source.Config{
							ResourceConfig: resource.ResourceConfig{
								DependsOn: []string{
									"3",
								},
							},
						},
					},
					"2": {
						Config: source.Config{
							ResourceConfig: resource.ResourceConfig{
								DependsOn: []string{
									"4",
								},
							},
						},
					},
					"3": {
						Config: source.Config{
							ResourceConfig: resource.ResourceConfig{
								DependsOn: []string{
									"2",
								},
							},
						},
					},
					"4": {},
				},
				Conditions: map[string]condition.Condition{
					"1": {
						Config: condition.Config{
							ResourceConfig: resource.ResourceConfig{
								DependsOn: []string{
									"3",
								},
							},
						},
					},
					"2": {
						Config: condition.Config{
							ResourceConfig: resource.ResourceConfig{
								DependsOn: []string{
									"4",
								},
							},
						},
					},
					"3": {
						Config: condition.Config{
							ResourceConfig: resource.ResourceConfig{
								DependsOn: []string{
									"2",
								},
							},
						},
					},
					"4": {},
				},
				Targets: map[string]target.Target{
					"1": {
						Config: target.Config{
							ResourceConfig: resource.ResourceConfig{
								DependsOn: []string{
									"3",
								},
							},
							DisableConditions: true,
						},
					},
					"2": {
						Config: target.Config{
							ResourceConfig: resource.ResourceConfig{
								DependsOn: []string{
									"4",
								},
							},
							DisableConditions: true,
						},
					},
					"3": {
						Config: target.Config{
							ResourceConfig: resource.ResourceConfig{
								DependsOn: []string{
									"2",
								},
							},
							DisableConditions: true,
						},
					},
					"4": {
						Config: target.Config{
							DisableConditions: true,
						},
					},
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
			Conf: Pipeline{
				Sources: map[string]source.Source{
					"1": {
						Config: source.Config{
							ResourceConfig: resource.ResourceConfig{
								DependsOn: []string{
									"2",
								},
							},
						},
					},
				},
				Conditions: map[string]condition.Condition{
					"2": {
						Config: condition.Config{
							ResourceConfig: resource.ResourceConfig{
								DependsOn: []string{
									"3",
								},
							},
						},
					},
				},
				Targets: map[string]target.Target{
					"3": {
						Config: target.Config{
							ResourceConfig: resource.ResourceConfig{
								DependsOn: []string{
									"4",
								},
							},
							DisableConditions: true,
						},
					},
				},
			},
			ExpectedResult: [][]ResultLeaf{},
			ExpectedErr:    ErrNotValidDependsOn,
		},
		{
			Name: "Scenario 4",
			Conf: Pipeline{
				Sources: map[string]source.Source{
					"1": {
						Config: source.Config{
							ResourceConfig: resource.ResourceConfig{
								DependsOn: []string{
									"target#2",
								},
							},
						},
					},
					"2": {
						Config: source.Config{
							ResourceConfig: resource.ResourceConfig{
								DependsOn: []string{
									"condition#3",
									"condition#2",
								},
							},
						},
					},
					"3": {
						Config: source.Config{
							ResourceConfig: resource.ResourceConfig{
								DependsOn: []string{
									"target#4",
								},
							},
						},
					},
					"4": {
						Config: source.Config{
							ResourceConfig: resource.ResourceConfig{
								DependsOn: []string{
									"target#4",
								},
							},
						},
					},
				},
				Conditions: map[string]condition.Condition{
					"1": {
						Config: condition.Config{
							ResourceConfig: resource.ResourceConfig{
								DependsOn: []string{
									"condition#2",
								},
							},
						},
					},
					"2": {
						Config: condition.Config{
							ResourceConfig: resource.ResourceConfig{
								DependsOn: []string{
									"condition#4",
								},
							},
						},
					},
					"3": {
						Config: condition.Config{
							ResourceConfig: resource.ResourceConfig{
								DependsOn: []string{
									"source#4",
									"source#3",
								},
							},
						},
					},
					"4": {
						Config: condition.Config{
							ResourceConfig: resource.ResourceConfig{
								DependsOn: []string{
									"target#4",
								},
							},
						},
					},
				},
				Targets: map[string]target.Target{
					"1": {
						Config: target.Config{
							ResourceConfig: resource.ResourceConfig{
								DependsOn: []string{
									"condition#1",
								},
							},
							DisableConditions: true,
						},
					},
					"2": {
						Config: target.Config{
							ResourceConfig: resource.ResourceConfig{
								DependsOn: []string{
									"target#1",
									"target#3",
								},
							},
							DisableConditions: true,
						},
					},
					"3": {
						Config: target.Config{
							ResourceConfig: resource.ResourceConfig{
								DependsOn: []string{
									"source#2",
								},
							},
							DisableConditions: true,
						},
					},
					"4": {
						Config: target.Config{
							DisableConditions: true,
						},
					},
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
			Conf: Pipeline{
				Sources: map[string]source.Source{
					"1": {
						Config: source.Config{
							ResourceConfig: resource.ResourceConfig{
								DependsOn: []string{
									"2",
								},
							},
						},
					},
					"2": {
						Config: source.Config{
							ResourceConfig: resource.ResourceConfig{
								DependsOn: []string{
									"1",
								},
							},
						},
					},
				},
				Conditions: map[string]condition.Condition{
					"1": {
						Config: condition.Config{
							ResourceConfig: resource.ResourceConfig{
								DependsOn: []string{
									"2",
								},
							},
						},
					},
					"2": {
						Config: condition.Config{
							ResourceConfig: resource.ResourceConfig{
								DependsOn: []string{
									"1",
								},
							},
						},
					},
				},
				Targets: map[string]target.Target{
					"1": {
						Config: target.Config{
							ResourceConfig: resource.ResourceConfig{
								DependsOn: []string{
									"2",
								},
							},
							DisableConditions: true,
						},
					},
					"2": {
						Config: target.Config{
							ResourceConfig: resource.ResourceConfig{
								DependsOn: []string{
									"1",
								},
							},
							DisableConditions: true,
						},
					},
				},
			},
			ExpectedResult: [][]ResultLeaf{},
			ExpectedErr:    ErrDependsOnLoopDetected,
		},
		{
			Name: "Scenario 6: Target Without all condition",
			Conf: Pipeline{
				Conditions: map[string]condition.Condition{
					"1": {},
					"2": {},
					"3": {},
					"4": {},
				},
				Targets: map[string]target.Target{
					"1": {},
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
			Name: "Scenario 7: Target With condition ids",
			Conf: Pipeline{
				Conditions: map[string]condition.Condition{
					"1": {
						Config: condition.Config{
							ResourceConfig: resource.ResourceConfig{
								DependsOn: []string{
									"2",
								},
							},
						},
					},
					"2": {},
					"3": {},
					"4": {
						Config: condition.Config{
							ResourceConfig: resource.ResourceConfig{
								DependsOn: []string{
									"2",
								},
							},
						},
					},
				},
				Targets: map[string]target.Target{
					"1": {
						Config: target.Config{
							DeprecatedConditionIDs: []string{"1", "4"},
						},
					},
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
			Conf: Pipeline{
				Sources: map[string]source.Source{
					"1": {
						Config: source.Config{
							ResourceConfig: resource.ResourceConfig{
								DependsOn: []string{
									"condition#1",
								},
							},
						},
					},
					"2": {
						Config: source.Config{
							ResourceConfig: resource.ResourceConfig{
								DependsOn: []string{
									"1",
								},
							},
						},
					},
				},
				Conditions: map[string]condition.Condition{
					"1": {},
					"2": {
						Config: condition.Config{
							SourceID: "2",
						},
					},
				},
				Targets: map[string]target.Target{
					"1": {
						Config: target.Config{
							SourceID:          "2",
							DisableConditions: true,
						},
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
			Conf: Pipeline{

				Sources: map[string]source.Source{
					"1": {},
				},
				Conditions: map[string]condition.Condition{
					"1": {
						Config: condition.Config{
							DisableSourceInput: true,
						}},
				},
				Targets: map[string]target.Target{
					"1": {
						Config: target.Config{
							DisableSourceInput: true,
						},
					},
					"5": {
						Config: target.Config{
							ResourceConfig: resource.ResourceConfig{
								DependsOn: []string{
									"1",
								},
							},
							DisableSourceInput: true,
						},
					},
					"6": {
						Config: target.Config{
							ResourceConfig: resource.ResourceConfig{
								DependsOn: []string{
									"1",
								},
							},
							DependsOnChange:    true,
							DisableSourceInput: true,
						},
					},
					"7": {
						Config: target.Config{
							ResourceConfig: resource.ResourceConfig{
								DependsOn: []string{
									"5",
								},
							},
							DependsOnChange:    true,
							DisableSourceInput: true,
						},
					},
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
			Conf: Pipeline{
				Sources: map[string]source.Source{
					"1": {},
					"2": {},
					"3": {},
					"4": {},
				},
				Conditions: map[string]condition.Condition{
					"1": {
						Config: condition.Config{
							ResourceConfig: resource.ResourceConfig{
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
			Conf: Pipeline{
				Sources: map[string]source.Source{
					"1": {},
				},
				Conditions: map[string]condition.Condition{
					"1": {
						Config: condition.Config{
							ResourceConfig: resource.ResourceConfig{
								DependsOn: []string{"source#1"},
							},
							SourceID: "1",
						},
					},
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
	}

	for _, data := range testdata {
		t.Run(data.Name, func(t *testing.T) {
			p := Pipeline{
				Sources:    data.Conf.Sources,
				Conditions: data.Conf.Conditions,
				Targets:    data.Conf.Targets,
			}
			gotSortedDag, err := p.SortedResources()

			if err != nil && data.ExpectedErr != nil {
				if strings.Compare(err.Error(), data.ExpectedErr.Error()) != 0 {
					t.Errorf("Unexpected error:\nExpected:\t\t%q\nGot:\t\t\t%q",
						data.ExpectedErr,
						err.Error())
				}
			} else if err != nil && data.ExpectedErr == nil {
				t.Errorf("Unexpected error:\nExpected:\t\tnil\nGot:\t\t\t%q",
					err.Error())

			} else if err == nil && data.ExpectedErr != nil {
				t.Errorf("Unexpected error:\nExpected:\t\t%q\nGot:\t\t\tnil",
					data.ExpectedErr)
			}

			if gotSortedDag == nil {
				return
			}
			results, err := gotSortedDag.GetOrderedDescendants(rootVertex)

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
