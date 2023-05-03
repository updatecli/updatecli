package awsami

import (
	"errors"
	"strings"
	"testing"

	"github.com/updatecli/updatecli/pkg/core/result"
)

func TestSource(t *testing.T) {

	// Disable source testing with running short tes
	if testing.Short() {
		return
	}

	for id, d := range dataset {
		// Do not run test for source if the spec is not valid
		if d.expectedError == ErrSpecNotValid {
			return
		}

		d.ami.apiClient = mockDescribeImagesOutput{
			Resp: d.mockedResponse,
		}

		gotResult := result.Source{}

		err := d.ami.Source("", &gotResult)

		if !errors.Is(err, d.expectedError) {
			t.Errorf("[%d] Wrong error:\nExpected Error:\t%v\nGot:\t\t%v\n",
				id, d.expectedError, err)
		}

		if strings.Compare(gotResult.Information, d.expectedSource) != 0 {
			t.Errorf("[%d] Wrong AMI ID returned:\nExpected Result:\t\t%q\nGot:\t\t\t\t\t%q",
				id,
				d.expectedSource,
				gotResult.Information)
		}
	}
}
