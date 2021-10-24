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
				File: "testdata/data_0.xml",
				Key:  ".name.firstname",
			},
			expectedResult: false,
		},
		{
			data: XML{
				File:  "testdata/data_0.xml",
				Key:   ".name.firstname",
				Value: "John",
			},
			expectedResult: true,
		},
		{
			data: XML{
				File:  "testdata/data_0.xml",
				Key:   ".name.firstname",
				Value: "wrongValue",
			},
			expectedResult: false,
		},
		//{
		//	data: XML{
		//		File: "testdata/data_0.xml",
		//		Key:  ".name.donotExit",
		//	},
		//	expectedResult: false,
		//	expectedError:  errors.New("could not find value: no value found for selector: .donotExit: map[firstname:John lastname:Doe]"),
		//},
		{
			data: XML{
				File: "testdata/data_1.xml",
				Key:  ".breakfast_menu.food.[0].name",
			},
			expectedResult: false,
		},
		{
			data: XML{
				File:  "testdata/data_1.xml",
				Key:   ".breakfast_menu.food.[0].name",
				Value: "Belgian Waffles",
			},
			expectedResult: true,
		},
		{
			data: XML{
				File:  "testdata/data_1.xml",
				Key:   ".breakfast_menu.food.[0].name",
				Value: "wrongValue",
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
