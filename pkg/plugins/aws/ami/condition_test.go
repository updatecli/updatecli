package ami

import (
	"testing"
)

type ConditionData struct {
	ami            AMI
	expectedResult bool
}

type ConditionDataSet []ConditionData

var (
	conditionDataset = ConditionDataSet{
		{
			ami: AMI{
				Region: "us-east-2",
				Filters: Filters{
					{
						Name:   "name",
						Values: "jenkins-agent-ubuntu*",
					},
				},
			},
			expectedResult: true,
		},
		{
			ami: AMI{
				Region: "us-east-2",
				Filters: Filters{
					{
						Name:   "name",
						Values: "jenkins-agent-ubuntu-18-amd64-20210422161407",
					},
				},
			},
			expectedResult: true,
		},
		{
			ami: AMI{
				Filters: Filters{
					{
						Name:   "image-id",
						Values: "ami-0477181fce0d41679",
					},
				},
			},
			expectedResult: true,
		},
		{
			ami: AMI{
				Filters: Filters{
					{
						Name:   "image-id",
						Values: "xxx",
					},
				},
			},
			expectedResult: false,
		},
	}
)

func TestCondition(t *testing.T) {

	for id, d := range conditionDataset {
		got, err := d.ami.Condition("")
		if err != nil {
			t.Errorf("Unexpected error: %q",
				err)
		}

		if got != d.expectedResult {
			t.Errorf("[%d] Wrong AMI ID returned:\nExpected Result:\t\t%v\nGot:\t\t\t\t\t%v",
				id,
				d.expectedResult,
				got)
		}
	}
}
