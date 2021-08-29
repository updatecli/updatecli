package ami

import (
	"strings"
	"testing"
)

func TestSource(t *testing.T) {

	// Disable source testing with running short tes
	if testing.Short() {
		return
	}

	for id, d := range dataset {
		got, err := d.ami.Source("")
		if err != nil {
			t.Errorf("Unexpected error: %q",
				err)
		}

		if strings.Compare(got, d.expectedSource) != 0 {
			t.Errorf("[%d] Wrong AMI ID returned:\nExpected Result:\t\t%q\nGot:\t\t\t\t\t%q",
				id,
				d.expectedSource,
				got)
		}
	}
}
