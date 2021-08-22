package ami

import (
	"strings"
	"testing"
)

type Data struct {
	ami            AMI
	expectedResult string
}

type DataSet []Data

var (
	dataset = DataSet{
		{
			ami: AMI{
				Region: "us-east-2",
				Filters: Filters{
					Name: "jenkins-agent-ubuntu*",
				},
			},
			expectedResult: "ami-0ff3b7aefa91e0935",
		},
		{
			ami: AMI{
				Region: "us-east-2",
				Filters: Filters{
					Name: "jenkins-agent-ubuntu-18-amd64-20210422161407",
				},
			},
			expectedResult: "ami-04099718d202c0132",
		},
		{
			ami: AMI{
				Filters: Filters{
					ImageID: "ami-0477181fce0d41679",
				},
			},
			expectedResult: "ami-0477181fce0d41679",
		},
		{
			ami: AMI{
				Filters: Filters{
					ImageID: "xxx",
				},
			},
			expectedResult: "",
		},
	}
)

func TestSource(t *testing.T) {

	for id, d := range dataset {
		got, err := d.ami.Source("")
		if err != nil {
			t.Errorf("Unexpected error: %q",
				err)
		}

		if strings.Compare(got, d.expectedResult) != 0 {
			t.Errorf("[%d] Wrong AMI ID returned:\nExpected Result:\t\t%q\nGot:\t\t\t\t\t%q",
				id,
				d.expectedResult,
				got)
		}
	}
}
