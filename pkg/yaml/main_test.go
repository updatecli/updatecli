package yaml

import (
	"errors"
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
image3:
- name: a
  version: 1.0
- name: b
  version: 1.5
image4:
- c
- d
- f
image5::tag: 1.17
image6::tags: 
- 1.17
- 1.18
image7@backend:
  - repository: golang
  - repository: nodejs
`
var data2 = `
- image:
    repository: nginx
    tag: 1.17
- database:
    image:
      repository: postgresql
      tag: 10
- image2: none
- image3:
  - name: a
    version: 1.0
  - name: b
    version: 1.5
- image4:
  - c
  - d
  - f
- 5
  `

var source = "1.17"

func TestGetPositionKeyValue(t *testing.T) {
	type dataSet struct {
		key              string
		expectedKey      string
		expectedPosition int
		expectedError    error
	}

	dataset := []dataSet{
		{
			key:              "image",
			expectedKey:      "image",
			expectedPosition: -1,
			expectedError:    nil,
		},
		{
			key:              "image[0]",
			expectedKey:      "image",
			expectedPosition: 0,
			expectedError:    nil,
		},
		{
			key:              "[0]image",
			expectedKey:      "",
			expectedPosition: -1,
			expectedError:    errors.New("Error: key '[0]image' cannot contains yaml special characters"),
		},
		{
			key:              "image[x]",
			expectedKey:      "",
			expectedPosition: -1,
			expectedError:    fmt.Errorf("Error: key 'image[x]' cannot contains yaml special characters"),
		},
		{
			key:              "im[0]age",
			expectedKey:      "",
			expectedPosition: -1,
			expectedError:    fmt.Errorf("Error: key 'im[0]age' cannot contains yaml special characters"),
		},
		{
			key:              "#image",
			expectedKey:      "",
			expectedPosition: -1,
			expectedError:    fmt.Errorf("Error: key '#image' cannot contains yaml special characters"),
		},
	}

	for _, d := range dataset {
		gotKey, gotPosition, gotError := getPositionKeyValue(d.key)
		if gotError == nil && d.expectedError == nil {
			if gotKey != d.expectedKey || gotPosition != d.expectedPosition {
				t.Errorf("Returned getPositionKeyValue is wrong for key '%s'"+
					"key expected %s, got %s "+
					"position expected %d, got %d "+
					d.key,
					d.expectedKey, gotKey,
					d.expectedPosition, gotPosition)
			}

		} else if gotError != nil && d.expectedError != nil {
			if gotKey != d.expectedKey || gotPosition != d.expectedPosition || gotError.Error() != d.expectedError.Error() {
				t.Errorf("Returned getPositionKeyValue value is wrong for 'key' %s"+
					"key expected %s, got %s "+
					"position expected %d, got %d "+
					"Error expected %v, got %v ",
					d.key,
					d.expectedKey, gotKey,
					d.expectedPosition, gotPosition,
					d.expectedError, gotError)
			}
		} else {
			t.Errorf("Returned getPositionKeyValue value is wrong for 'key' %s"+
				"key expected %s, got %s "+
				"position expected %d, got %d "+
				"Error expected %v, got %v ",
				d.key,
				d.expectedKey, gotKey,
				d.expectedPosition, gotPosition,
				d.expectedError, gotError)
		}
	}
}

func TestIsPositionKey(t *testing.T) {
	type dataSet struct {
		key      string
		expected bool
	}

	dataset := []dataSet{
		{
			key:      "image",
			expected: false,
		},
		{
			key:      "image[0]",
			expected: true,
		},
		{
			key:      "image&tags[0]",
			expected: true,
		},
		{
			key:      "image&tags\\[0\\]",
			expected: false,
		},
		{
			key:      "[0]image",
			expected: false,
		},
		{
			key:      "[0]image::tag",
			expected: false,
		},
		{
			key:      "image[x]",
			expected: false,
		},
		{
			key:      "im[0]age",
			expected: false,
		},
		{
			key:      "image7@backend[1]",
			expected: true,
		},
	}

	for _, d := range dataset {
		got := isPositionKey(d.key)
		if got != d.expected {
			t.Errorf("isPositionKey is wrong for key %s! expected %v, got %v", d.key, d.expected, got)
		}
	}
}

func TestReplace(t *testing.T) {

	type dataSet struct {
		key                []string
		expectedOldVersion string
		expectedValueFound bool
	}

	//https://github.com/go-yaml/yaml/issues/599

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
		{
			key:                []string{"image3[0]", "version"},
			expectedOldVersion: "1.0",
			expectedValueFound: true,
		},
		{
			key:                []string{"image3[1]", "version"},
			expectedOldVersion: "1.5",
			expectedValueFound: true,
		},
		{
			key:                []string{"image5::tag"},
			expectedOldVersion: "1.17",
			expectedValueFound: true,
		},
		{
			key:                []string{"image6::tags[0]"},
			expectedOldVersion: "1.17",
			expectedValueFound: true,
		},
		{
			key:                []string{"image4[0]"},
			expectedOldVersion: "c",
			expectedValueFound: true,
		},
		{
			key:                []string{"image4[1]"},
			expectedOldVersion: "d",
			expectedValueFound: true,
		},
		{
			key:                []string{"image4[10]"},
			expectedOldVersion: "",
			expectedValueFound: false,
		},
		{
			key:                []string{"image7@backend[0]", "repository"},
			expectedOldVersion: "golang",
			expectedValueFound: true,
		},
	}

	dataset2 := []dataSet{
		{
			key:                []string{"[0]", "image", "tag"},
			expectedOldVersion: "1.17",
			expectedValueFound: true,
		},
		{
			key:                []string{"[1]", "database", "image", "tag"},
			expectedOldVersion: "10",
			expectedValueFound: true,
		},
		{
			key:                []string{"[5]"},
			expectedOldVersion: "5",
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

	out2 := yaml.Node{}
	err = yaml.Unmarshal([]byte(data2), &out2)
	if err != nil {
		fmt.Println(err)
	}

	for _, d := range dataset2 {
		valueFound, oldVersion, _ := replace(&out2, d.key, source, 1)

		if valueFound != d.expectedValueFound {
			t.Errorf("Value not found for key %v! expected %v, got %v", d.key, d.expectedValueFound, valueFound)
		}

		if oldVersion != d.expectedOldVersion {
			t.Errorf("Old Version mismatch for key %v! expected %v, got %v", d.key, d.expectedOldVersion, oldVersion)
		}
	}
}
