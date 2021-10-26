package xml

import (
	"errors"
	"os"
	"testing"
)

type TargetDataset []TargetData

type TargetData struct {
	data            XML
	expectedChanged bool
	expectedError   error
}

var (
	targetDataset = TargetDataset{
		{
			data: XML{
				spec: Spec{
					File:  "testdata/data_0.xml",
					Key:   ".name.firstname",
					Value: "Bob",
				},
			},
			expectedChanged: true,
		},
		{
			data: XML{
				spec: Spec{
					File:  "testdata/data_0.xml",
					Key:   ".name.firstname",
					Value: "John",
				},
			},
			expectedChanged: false,
		},
		{
			data: XML{
				spec: Spec{
					File:  "testdata/data_2.xml",
					Key:   ".name.firstname",
					Value: "Bob",
				},
			},
			expectedChanged: true,
		},
		//{
		//	data: XML{
		//		spec: Spec{
		//			File:  "testdata/data_2.xml",
		//			Key:   ".name.donotexist",
		//			Value: "Bob",
		//		},
		//	},
		//	expectedChanged: false,
		//	expectedError: fmt.Errorf("could not find value: %w",
		//		errors.New("no value found for selector: .donotexist: map[firstname:map[#text:John -tag:test] lastname:Doe]")),
		//},
		// Failing due to https://github.com/TomWright/dasel/issues/175
		//{
		//	data: XML{
		//		spec: Spec{
		//			File:  "testdata/data_2.xml",
		//			Key:   ".name.firstname",
		//			Value: "John",
		//		},
		//	},
		//	expectedChanged: false,
		//},
		{
			data: XML{
				spec: Spec{
					File:  "testdata/doNotExist.xml",
					Key:   ".name.firstname",
					Value: "Alice",
				},
			},
			expectedChanged: false,
			expectedError:   os.ErrNotExist,
		},
	}
)

func TestTarget(t *testing.T) {

	for id, s := range targetDataset {
		got, err := s.data.Target("", true)

		if !errors.Is(err, s.expectedError) {
			t.Errorf("Wrong target error for %v:\n\tExpected:\t%s\n\tGot:\t\t%s\n", id, s.expectedError, err)

		}
		if s.expectedChanged != got {
			t.Errorf("Wrong target result for %v:\n\tExpected:\t%t\n\tGot:\t\t%t\n", id, s.expectedChanged, got)
		}
	}
}
