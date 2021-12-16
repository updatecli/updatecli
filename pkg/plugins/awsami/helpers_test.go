package awsami

import (
	"strings"
	"testing"
)

func TestGetLatestAmiID(t *testing.T) {

	for id, d := range dataset {
		d.ami.apiClient = mockDescribeImagesOutput{
			Resp: d.mockedResponse,
		}
		got, err := d.ami.getLatestAmiID()
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
