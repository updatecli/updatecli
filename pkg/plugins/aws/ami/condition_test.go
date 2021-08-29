package ami

import "testing"

func TestCondition(t *testing.T) {
	// Disable condition testing with running short test
	if testing.Short() {
		return
	}

	for id, d := range dataset {
		got, err := d.ami.Condition("")
		if err != nil {
			t.Errorf("Unexpected error: %q",
				err)
		}

		if got != d.expectedCondition {
			t.Errorf("[%d] Wrong AMI conditional result:\nExpected Result:\t\t%v\nGot:\t\t\t\t\t%v",
				id,
				d.expectedSource,
				got)
		}
	}
}
