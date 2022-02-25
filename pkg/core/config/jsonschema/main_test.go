package jsonschema

import (
	"io/ioutil"
	"os"
	"testing"
)

type Dataset []Data

// Data represent a single json schema test data
type Data struct {
	updatecliManifest string
	expectedErr       error
	expectedResult    bool
}

var (
	TestDataset Dataset = Dataset{
		{
			updatecliManifest: "../../../../e2e/updatecli.d",
			expectedResult:    true,
			expectedErr:       nil,
		},
	}
)

func readFile(file string) ([]byte, error) {
	c, err := os.Open(file)

	if err != nil {
		return nil, err
	}

	defer c.Close()

	content, err := ioutil.ReadAll(c)
	if err != nil {
		return nil, err
	}

	return content, nil
}

// TestValidate tests the function Validate by Using updatecli manifest
// from the e2e directory.
func TestValidate(t *testing.T) {

	failingManifest := []string{}

	for _, d := range TestDataset {
		updatecliManifests, err := getFilesWithSuffix(d.updatecliManifest, ".yaml")
		if err != nil {
			t.Errorf("%s", err)
		}

		for _, manifest := range updatecliManifests {
			t.Logf("Validating file %q\n", manifest)
			m, err := readFile(manifest)
			if err != nil {
				t.Errorf("%s", err)
			}

			result, err := Validate(m)
			if err != nil {
				t.Errorf("Unexpected error for file %q\nError: %s",
					manifest, err)
			}
			if d.expectedResult != result {
				failingManifest = append(failingManifest, manifest)
				t.Errorf("Expecting file %q to be valid", manifest)
			}
		}

	}

	if len(failingManifest) > 0 {
		t.Logf("%d updatecli manifests failed\n", len(failingManifest))
		for _, m := range failingManifest {
			t.Logf("\t%s\n", m)
		}
	}

}
