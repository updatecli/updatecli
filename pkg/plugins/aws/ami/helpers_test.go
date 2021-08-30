package ami

import (
	"strings"
	"testing"
)

func TestGetLatestAmiID(t *testing.T) {

	for id, d := range dataset {
		got, err := d.ami.getLatestAmiID(
			mockDescribeImagesOutput{
				Resp: d.resp})
		if err != nil {
			t.Errorf("Unexpected error: %q",
				err)
		}

		if strings.Compare(got, d.expectedGetAMI) != 0 {

			t.Errorf("[%d] Wrong AMI ID returned:\nExpected Result:\t\t%q\nGot:\t\t\t\t\t%q",
				id,
				d.expectedSource,
				got)
		}
	}
}
