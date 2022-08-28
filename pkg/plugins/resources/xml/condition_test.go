package xml

import (
	"errors"
	"testing"
)

type ConditionDataset []ConditionData

type ConditionData struct {
	data           XML
	expectedResult bool
	expectedError  error
}

var (
	conditionDataset = ConditionDataset{
		{
			data: XML{
				spec: Spec{
					File: "testdata/data_0.xml",
					Path: "/name/firstname",
				},
			},
			expectedResult: false,
		},
		{
			data: XML{
				spec: Spec{
					File:  "testdata/data_0.xml",
					Path:  "/name/firstname",
					Value: "John",
				},
			},
			expectedResult: true,
		},
		{
			data: XML{
				spec: Spec{
					File:  "testdata/data_0.xml",
					Path:  "/name/firstname",
					Value: "wrongValue",
				},
			},
			expectedResult: false,
		},
		{
			data: XML{
				spec: Spec{
					File: "testdata/data_0.xml",
					Path: ".name.donotExit",
				},
			},
			expectedResult: false,
		},
		{
			data: XML{
				spec: Spec{
					File: "testdata/data_1.xml",
					Path: "/breakfast_menu/food[0]/name",
				},
			},
			expectedResult: false,
		},
		{
			data: XML{
				spec: Spec{
					File:  "testdata/data_1.xml",
					Path:  "/breakfast_menu/food[0]/name",
					Value: "Belgian Waffles",
				},
			},
			expectedResult: true,
		},
		{
			data: XML{
				spec: Spec{
					File:  "testdata/data_1.xml",
					Path:  "/breakfast_menu.food[0]/name",
					Value: "wrongValue",
				},
			},
			expectedResult: false,
		},
	}
)

func TestCondition(t *testing.T) {

	for id, c := range conditionDataset {
		got, err := c.data.Condition("")
		if !errors.Is(c.expectedError, err) {
			t.Errorf("Wrong error for %v:\nGot:\n\t%q\nExpected:\n\t%q\n",
				id, err.Error(), c.expectedError.Error())
		}
		if c.expectedResult != got {
			t.Errorf("Wrong source result for %v:\n\tExpected:\t%t\n\tGot:\t%t",
				id, c.expectedResult, got)
		}
	}
}
