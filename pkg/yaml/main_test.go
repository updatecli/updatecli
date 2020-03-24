package yaml

import (
	"fmt"
	"testing"

	"gopkg.in/yaml.v3"
)

var data = `
image:
  repository: nginx
  tag: 1.17
database:
  image:
    repository: postgresql
    tag: 10
image2: none
`

var source = "1.17"

func TestReplace(t *testing.T) {

	type dataSet struct {
		key                []string
		expectedOldVersion string
		expectedValueFound bool
	}

	dataset := []dataSet{
		{
			key:                []string{"image", "tag"},
			expectedOldVersion: "1.17",
			expectedValueFound: true,
		},
		{
			key:                []string{"database", "image", "tag"},
			expectedOldVersion: "10",
			expectedValueFound: true,
		},
		{
			key:                []string{"image2"},
			expectedOldVersion: "none",
			expectedValueFound: true,
		},
	}

	out := yaml.Node{}

	err := yaml.Unmarshal([]byte(data), &out)
	if err != nil {
		fmt.Println(err)
	}

	for _, d := range dataset {
		valueFound, oldVersion, _ := replace(&out, d.key, source, 1)

		if valueFound != d.expectedValueFound {
			t.Errorf("Value not found for key %v! expected %v, got %v", d.key, d.expectedValueFound, valueFound)
		}

		if oldVersion != d.expectedOldVersion {
			t.Errorf("Old Version mismatch for key %v! expected %v, got %v", d.key, d.expectedOldVersion, oldVersion)
		}
	}
}
