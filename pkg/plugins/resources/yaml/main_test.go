package yaml

import (
	"reflect"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/text"
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
- name: c
  version: 3.5
- name: d
  version: 5.5
image4:
- c
- d
- f
- g
- h
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
			expectedKey:      "[0]image",
			expectedPosition: -1,
			expectedError:    nil,
		},
		{
			key:              "image[x]",
			expectedKey:      "image[x]",
			expectedPosition: -1,
			expectedError:    nil,
		},
		{
			key:              "im[0]age",
			expectedKey:      "im[0]age",
			expectedPosition: -1,
			expectedError:    nil,
		},
		{
			key:              "#image",
			expectedKey:      "#image",
			expectedPosition: -1,
			expectedError:    nil,
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

	dataset1 := []dataSet{
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
			key:                []string{"image3[2]", "version"},
			expectedOldVersion: "3.5",
			expectedValueFound: true,
		},
		{
			key:                []string{"image3[3]", "version"},
			expectedOldVersion: "5.5",
			expectedValueFound: true,
		},
		{
			key:                []string{"image3[4]", "version"},
			expectedOldVersion: "",
			expectedValueFound: false,
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
			key:                []string{"image4[2]"},
			expectedOldVersion: "f",
			expectedValueFound: true,
		},
		{
			key:                []string{"image4[3]"},
			expectedOldVersion: "g",
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
	require.NoError(t, err)

	for _, d := range dataset1 {
		valueFound, oldVersion, _ := replace(&out, d.key, source, 1)

		assert.Equal(t, d.expectedValueFound, valueFound)
		assert.Equal(t, d.expectedOldVersion, oldVersion)
	}

	out2 := yaml.Node{}
	err = yaml.Unmarshal([]byte(data2), &out2)
	require.NoError(t, err)

	for _, d := range dataset2 {
		valueFound, oldVersion, _ := replace(&out2, d.key, source, 1)

		assert.Equal(t, d.expectedValueFound, valueFound)
		assert.Equal(t, d.expectedOldVersion, oldVersion)
	}
}

func TestIndent(t *testing.T) {

	inputData := `
image:
  repository: nginx
image4:
- c
- d
- f
`
	outputData := `image:
    repository: apache
image4:
    - c
    - d
    - f
`
	out := yaml.Node{}

	err := yaml.Unmarshal([]byte(inputData), &out)
	if err != nil {
		logrus.Errorf("err - %s", err)
	}

	replace(&out, []string{"image", "repository"}, "apache", 1)

	raw, err := yaml.Marshal(&out)

	if err != nil {
		logrus.Errorf("err - %s", err)
	}

	if !(reflect.DeepEqual(raw, []byte(outputData))) {
		t.Errorf("Wrong Yaml output\nexpected:\t%#v\n\ngot:\t\t%#v\n", outputData, string(raw))
	}
}

func Test_Validate(t *testing.T) {
	tests := []struct {
		name          string
		spec          Spec
		mockFileExist bool
		isErrorWanted bool
	}{
		{
			name: "Normal case with 'File'",
			spec: Spec{
				File: "/tmp/test.yaml",
				Key:  "foo.bar",
			},
			isErrorWanted: false,
		},
		{
			name: "Normal case with more than one 'Files'",
			spec: Spec{
				Files: []string{
					"/tmp/test.yaml",
					"/tmp/bar.yaml",
				},
				Key: "foo.bar",
			},
			isErrorWanted: false,
		},
		{
			name: "Validation error when both 'File' and 'Files' are empty",
			spec: Spec{
				File:  "",
				Files: []string{},
				Key:   "foo.bar",
			},
			isErrorWanted: true,
		},
		{
			name: "Validation error when 'Key' is empty",
			spec: Spec{
				File: "/tmp/toto.yaml",
				Key:  "",
			},
			isErrorWanted: true,
		},
		{
			name: "Validation error when both 'File' and 'Files' are specified",
			spec: Spec{
				File: "test.yaml",
				Files: []string{
					"bar.yaml",
				},
				Key: "foo.bar",
			},
			isErrorWanted: true,
		},
		{
			name: "Validation error when 'Files' contains duplicates",
			spec: Spec{
				Files: []string{
					"test.yaml",
					"test.yaml",
				},
				Key: "foo.bar",
			},
			isErrorWanted: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yaml := Yaml{
				spec:             tt.spec,
				contentRetriever: &text.MockTextRetriever{},
			}
			gotErr := yaml.spec.Validate()
			if tt.isErrorWanted {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
		})
	}
}
