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
					Key:  ".name.firstname",
				},
			},
			expectedResult: "John",
		},
		{
			data: XML{
				spec: Spec{
					File: "testdata/data_1.xml",
					Key:  ".breakfast_menu.food.[0].name",
				},
			},
			expectedResult: "Belgian Waffles",
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
