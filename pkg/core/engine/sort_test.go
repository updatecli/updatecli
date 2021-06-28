package engine

import (
	"strings"
	"testing"

	"github.com/olblak/updateCli/pkg/core/config"
	"github.com/olblak/updateCli/pkg/core/engine/condition"
	"github.com/olblak/updateCli/pkg/core/engine/source"
	"github.com/olblak/updateCli/pkg/core/engine/target"
)

type SortedKeysData struct {
	Conf                     config.Config
	ExpectedSourcesResult    []string
	ExpectedConditionsResult []string
	ExpectedTargetsResult    []string
	ExpectedSourcesErr       error
	ExpectedConditionsErr    error
	ExpectedTargetsErr       error
}

type SortedKeysDataSet []SortedKeysData

var (
	sortedKeysDataset = SortedKeysDataSet{
		{
			Conf: config.Config{
				Sources: map[string]source.Source{
					"1": {
						DependsOn: []string{
							"2",
							"3",
						},
					},
					"2": {
						DependsOn: []string{
							"3",
						},
					},
					"3": {},
				},
				Conditions: map[string]condition.Condition{
					"1": {
						DependsOn: []string{
							"2",
						},
					},
					"2": {
						DependsOn: []string{
							"3",
						},
					},
					"3": {},
				},
				Targets: map[string]target.Target{
					"1": {
						DependsOn: []string{
							"2",
						},
					},
					"2": {
						DependsOn: []string{
							"3",
						},
					},
					"3": {},
				},
			},
			ExpectedSourcesResult: []string{
				"3", "2", "1",
			},
			ExpectedConditionsResult: []string{
				"3", "2", "1",
			},
			ExpectedTargetsResult: []string{
				"3", "2", "1",
			},
		},
		{
			Conf: config.Config{
				Sources: map[string]source.Source{
					"1": {
						DependsOn: []string{
							"3",
						},
					},
					"2": {
						DependsOn: []string{
							"4",
						},
					},
					"3": {
						DependsOn: []string{
							"2",
						},
					},
					"4": {},
				},
				Conditions: map[string]condition.Condition{
					"1": {
						DependsOn: []string{
							"3",
						},
					},
					"2": {
						DependsOn: []string{
							"4",
						},
					},
					"3": {
						DependsOn: []string{
							"2",
						},
					},
					"4": {},
				},
				Targets: map[string]target.Target{
					"1": {
						DependsOn: []string{
							"3",
						},
					},
					"2": {
						DependsOn: []string{
							"4",
						},
					},
					"3": {
						DependsOn: []string{
							"2",
						},
					},
					"4": {},
				},
			},
			ExpectedSourcesResult: []string{
				"4", "2", "3", "1",
			},
			ExpectedConditionsResult: []string{
				"4", "2", "3", "1",
			},
			ExpectedTargetsResult: []string{
				"4", "2", "3", "1",
			},
		},
		{
			Conf: config.Config{
				Sources: map[string]source.Source{
					"1": {
						DependsOn: []string{
							"2",
						},
					},
				},
				Conditions: map[string]condition.Condition{
					"2": {
						DependsOn: []string{
							"3",
						},
					},
				},
				Targets: map[string]target.Target{
					"3": {
						DependsOn: []string{
							"4",
						},
					},
				},
			},
			ExpectedSourcesResult:    []string{},
			ExpectedConditionsResult: []string{},
			ExpectedTargetsResult:    []string{},
			ExpectedSourcesErr:       ErrNotValidDependsOn,
			ExpectedConditionsErr:    ErrNotValidDependsOn,
			ExpectedTargetsErr:       ErrNotValidDependsOn,
		},
	}
)

func TestSortedSourcesKeys(t *testing.T) {

	for _, data := range sortedKeysDataset {
		// Test Source
		gotSortedSourcesKeys, err := SortedSourcesKeys(&data.Conf.Sources)
		if err != nil && data.ExpectedSourcesErr != nil {
			if strings.Compare(err.Error(), data.ExpectedSourcesErr.Error()) != 0 {
				t.Errorf("Unexpected error:\nExpected:\t\t%q\nGot:\t\t\t%q",
					data.ExpectedSourcesErr,
					err.Error())

			}

		} else if err != nil && data.ExpectedSourcesErr == nil {
			t.Errorf("Unexpected error:\nExpected:\t\tnil\nGot:\t\t\t%q",
				err.Error())

		} else if err == nil && data.ExpectedSourcesErr != nil {
			t.Errorf("Unexpected error:\nExpected:\t\t%q\nGot:\t\t\tnil",
				data.ExpectedSourcesErr)
		}

		for i := range gotSortedSourcesKeys {
			if len(data.ExpectedSourcesResult) < len(gotSortedSourcesKeys) {
				t.Errorf("Unexpected result length:\n\tExpected:\t%d\n\tGot:\t\t%d",
					len(data.ExpectedSourcesResult),
					len(gotSortedSourcesKeys))
				break

			}
			if strings.Compare(gotSortedSourcesKeys[i], data.ExpectedSourcesResult[i]) != 0 {
				t.Errorf("Unexpected result:\n\tExpected:\t%q\n\tGot:\t\t%q",
					data.ExpectedSourcesResult,
					gotSortedSourcesKeys)
			}
		}

	}

}

func TestSortedConditionsKeys(t *testing.T) {

	for _, data := range sortedKeysDataset {
		// Test Source
		gotSortedConditionsKeys, err := SortedConditionsKeys(&data.Conf.Conditions)

		if err != nil && data.ExpectedConditionsErr != nil {
			if strings.Compare(err.Error(), data.ExpectedConditionsErr.Error()) != 0 {
				t.Errorf("Unexpected error:\nExpected:\t\t%q\nGot:\t\t\t%q",
					data.ExpectedConditionsErr,
					err.Error())
			}

		} else if err != nil && data.ExpectedConditionsErr == nil {
			t.Errorf("Unexpected error:\nExpected:\t\tnil\nGot:\t\t\t%q",
				err.Error())

		} else if err == nil && data.ExpectedConditionsErr != nil {
			t.Errorf("Unexpected error:\nExpected:\t\t%q\nGot:\t\t\tnil",
				data.ExpectedConditionsErr)
		}

		for i := range gotSortedConditionsKeys {
			if len(data.ExpectedConditionsResult) < len(gotSortedConditionsKeys) {
				t.Errorf("Unexpected result length:\n\tExpected:\t%d\n\tGot:\t\t%d",
					len(data.ExpectedConditionsResult),
					len(gotSortedConditionsKeys))
				break

			}
			if strings.Compare(gotSortedConditionsKeys[i], data.ExpectedConditionsResult[i]) != 0 {
				t.Errorf("Unexpected result:\n\tExpected:\t%q\n\tGot:\t\t%q",
					data.ExpectedConditionsResult,
					gotSortedConditionsKeys)
			}
		}

	}

}

func TestSortedTargetsKeys(t *testing.T) {

	for _, data := range sortedKeysDataset {
		// Test Source
		gotSortedTargetsKeys, err := SortedTargetsKeys(&data.Conf.Targets)

		if err != nil && data.ExpectedTargetsErr != nil {
			if strings.Compare(err.Error(), data.ExpectedTargetsErr.Error()) != 0 {
				t.Errorf("Unexpected error:\nExpected:\t\t%q\nGot:\t\t\t%q",
					data.ExpectedTargetsErr,
					err.Error())
			}

		} else if err != nil && data.ExpectedTargetsErr == nil {
			t.Errorf("Unexpected error:\nExpected:\t\tnil\nGot:\t\t\t%q",
				err.Error())

		} else if err == nil && data.ExpectedTargetsErr != nil {
			t.Errorf("Unexpected error:\nExpected:\t\t%q\nGot:\t\t\tnil",
				data.ExpectedTargetsErr)
		}

		for i := range gotSortedTargetsKeys {
			if len(data.ExpectedTargetsResult) < len(gotSortedTargetsKeys) {
				t.Errorf("Unexpected result length:\n\tExpected:\t%d\n\tGot:\t\t%d",
					len(data.ExpectedTargetsResult),
					len(gotSortedTargetsKeys))
				break

			}
			if strings.Compare(gotSortedTargetsKeys[i], data.ExpectedTargetsResult[i]) != 0 {
				t.Errorf("Unexpected result:\n\tExpected:\t%q\n\tGot:\t\t%q",
					data.ExpectedTargetsResult,
					gotSortedTargetsKeys)
			}
		}

	}

}
