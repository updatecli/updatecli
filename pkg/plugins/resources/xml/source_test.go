package xml

import "testing"

type SourceDataset []SourceData

type SourceData struct {
	data           XML
	expectedResult string
}

var (
	sourceDataset = SourceDataset{
		{
			data: XML{
				spec: Spec{
					File: "testdata/data_0.xml",
					Path: "//name/firstname",
				},
			},
			expectedResult: "John",
		},
		{
			data: XML{
				spec: Spec{
					File: "testdata/data_1.xml",
					Path: "doNotExist",
				},
			},
			expectedResult: "",
		},
		{
			data: XML{
				spec: Spec{
					File: "testdata/data_1.xml",
					Path: "//breakfast_menu[0]/name",
				},
			},
			expectedResult: "",
		},
	}
)

func TestSource(t *testing.T) {

	for _, s := range sourceDataset {
		got, err := s.data.Source("")
		if err != nil {
			t.Errorf("error: %s", err.Error())
		}
		if s.expectedResult != got {
			t.Errorf("Wrong source result:\n\tExpected:\t%q\n\tGot:\t%q", s.expectedResult, got)
		}
	}
}
