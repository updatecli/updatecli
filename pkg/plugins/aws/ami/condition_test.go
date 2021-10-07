package ami

import (
	"errors"
	"testing"
)

func TestCondition(t *testing.T) {
	// Disable condition testing with running short test
	if testing.Short() {
		return
	}

	for id, d := range dataset {
		got, err := d.ami.Condition("")

		if !errors.Is(err, d.expectedError) {
			t.Errorf("[%d] Wrong error:\nExpected Error:\t%v\nGot:\t\t%v\n",
				id, d.expectedError, err)
		}

		if got != d.expectedCondition {
			t.Errorf("[%d] Wrong AMI conditional result:\nExpected Result:\t\t%v\nGot:\t\t\t\t\t%v",
				id,
				d.expectedCondition,
				got)
		}
	}

	// Test inject image-id
	ami := AMI{
		Spec: Spec{
			Region:  "eu-west-1",
			Filters: Filters{},
		},
	}
	imageID := "ami-0a9972d9b4dbdabc7"
	exist, err := ami.Condition(imageID)
	if err != nil {
		t.Errorf("Unexpected error: %q",
			err)
	}

	if !exist {
		t.Errorf("[%s] Wrong AMI conditional result:\nExpected Result:\t\t%v\nGot:\t\t\t\t\t%v",
			imageID,
			true,
			exist)
	}
}
