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
}

type SortedKeysDataSet []SortedKeysData

var (
	sortedKeysDataset = SortedKeysDataSet{
		{
			Conf: config.Config{
				Sources: map[string]source.Source{
					"1": {
						DependsOn: "2",
					},
					"2": {
						DependsOn: "3",
					},
					"3": {},
				},
				Conditions: map[string]condition.Condition{
					"1": {
						DependsOn: "2",
					},
					"2": {
						DependsOn: "3",
					},
					"3": {},
				},
				Targets: map[string]target.Target{
					"1": {
						DependsOn: "2",
					},
					"2": {
						DependsOn: "3",
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
						DependsOn: "3",
					},
					"2": {
						DependsOn: "4",
					},
					"3": {
						DependsOn: "2",
					},
					"4": {},
				},
				Conditions: map[string]condition.Condition{
					"1": {
						DependsOn: "3",
					},
					"2": {
						DependsOn: "4",
					},
					"3": {
						DependsOn: "2",
					},
					"4": {},
				},
				Targets: map[string]target.Target{
					"1": {
						DependsOn: "3",
					},
					"2": {
						DependsOn: "4",
					},
					"3": {
						DependsOn: "2",
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
	}
)

func TestSortedSourcesKeys(t *testing.T) {

	for _, data := range sortedKeysDataset {
		// Test Source
		gotSortedSourcesKeys, err := SortedSourcesKeys(&data.Conf.Sources)
		if err != nil {
			t.Errorf("Unexpected error: %q", err.Error())
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
		if err != nil {
			t.Errorf("Unexpected error: %q", err.Error())
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
		if err != nil {
			t.Errorf("Unexpected error: %q", err.Error())
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
