package pipeline

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/condition"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/source"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
)

func TestSortedResourcesKeys(t *testing.T) {
	testdata := []struct {
		Name           string
		Conf           Pipeline
		ExpectedResult [][]string
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
		ExpectedResult: [][]string{
			{"source#3", "condition#3", "target#3"},
			{"source#2", "condition#2", "target#2"},
			{"source#1", "condition#1", "target#1"},
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
			ExpectedResult: [][]string{
				{"source#4",
					"condition#4",
					"target#4"},
				{"source#2",
					"condition#2",
					"target#2"},
				{"source#3",
					"condition#3",
					"target#3"},
				{"source#1",
					"condition#1",
					"target#1"},
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
			ExpectedResult: [][]string{},
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
			ExpectedResult: [][]string{
				{"target#4"},
				{"source#4", "source#3", "condition#4"},
				{"condition#3", "condition#2"},
				{"source#2", "condition#1"},
				{"target#3", "target#1"},
				{"target#2"},
				{"source#1"},
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
			ExpectedResult: [][]string{},
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
			ExpectedResult: [][]string{
				{"condition#1", "condition#2", "condition#3", "condition#4"},
				{"target#1"},
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
			ExpectedResult: [][]string{
				{"condition#2", "condition#3"},
				{"condition#1", "condition#4"},
				{"target#1"},
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
			ExpectedResult: [][]string{
				{"condition#1"},
				{"source#1"},
				{"source#2"},
				{"condition#2", "target#1"},
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
			ExpectedResult: [][]string{
				{"source#1", "condition#1"},
				{"target#1"},
				{"target#5"},
				{"target#6", "target#7"},
			},
		},
	}

	for _, data := range testdata {
		t.Run(data.Name, func(t *testing.T) {
			// Test Source
			gotSortedDag, err := SortedResources(&data.Conf.Sources, &data.Conf.Conditions, &data.Conf.Targets)

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
			logrus.Errorf("Got: %s", results)
			logrus.Errorf("Expected: %s", data.ExpectedResult)

			err = compareDag(data.ExpectedResult, results)
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

		})
	}
}

// compareDag allows to test that the resulted orderedDescendant of a dag
// match the expected result.
// As the sibling order is not guarantee, we need to  handle that
func compareDag(expected [][]string, got []string) (err error) {
	//{ Keep track of the current index in 'got'
	index := 0
	// Iterate over each sublist in 'expected'
	for _, sublist := range expected {
		// If remaining 'got' is smaller than current 'expected' sublist, return false
		if index+len(sublist) > len(got) {
			return fmt.Errorf("Not enough elem remaining in array")
		}

		// Extract the slice from 'got' to compare with current sublist
		gotSublist := got[index : index+len(sublist)]

		// Sort both the current 'expected' sublist and the corresponding 'got' sublist
		sort.Strings(sublist)
		sort.Strings(gotSublist)

		// Check if they match
		if !reflect.DeepEqual(sublist, gotSublist) {
			return fmt.Errorf("Sibblings are not equals")
		}

		// Move the index forward by the length of the current sublist
		index += len(sublist)
	}

	// If we've processed all 'expected' sublists and matched them correctly, return true
	if index != len(got) {
		return fmt.Errorf("Leafs are not equals")
	}
	return nil
}
