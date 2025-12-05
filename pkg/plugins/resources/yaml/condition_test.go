package yaml

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/text"
)

func ptrInt(i int) *int {
	return &i
}

func Test_Condition(t *testing.T) {
	tests := []struct {
		name             string
		spec             Spec
		files            map[string]file
		inputSourceValue string
		mockedContents   map[string]string
		mockedError      error
		isResultWanted   bool
		isErrorWanted    bool
	}{
		{
			name: "yamlpath Passing Case with multiple document",
			spec: Spec{
				File:   "data.yaml",
				Key:    "$.name",
				Engine: "yamlpath",
			},
			files: map[string]file{
				"data.yaml": {
					originalFilePath: "data.yaml",
					filePath:         "data.yaml",
				},
			},
			inputSourceValue: "John",
			mockedContents: map[string]string{
				"data.yaml": `---
name: John
age: 30
---
name: Doe
age: 25
`,
			},
			isResultWanted: false,
		},
		{
			name: "yamlpath Passing Case with multiple document",
			spec: Spec{
				File:          "data.yaml",
				Key:           "$.name",
				Engine:        "yamlpath",
				DocumentIndex: ptrInt(1),
			},
			files: map[string]file{
				"data.yaml": {
					originalFilePath: "data.yaml",
					filePath:         "data.yaml",
				},
			},
			inputSourceValue: "Doe",
			mockedContents: map[string]string{
				"data.yaml": `---
name: John
age: 30
---
name: Doe
age: 25
`,
			},
			isResultWanted: true,
		},
		{
			name: "Passing Case with multiple document",
			spec: Spec{
				File:          "data.yaml",
				Key:           "$.name",
				Engine:        "go-yaml",
				DocumentIndex: ptrInt(1),
			},
			files: map[string]file{
				"data.yaml": {
					originalFilePath: "data.yaml",
					filePath:         "data.yaml",
				},
			},
			inputSourceValue: "Doe",
			mockedContents: map[string]string{
				"data.yaml": `---
name: John
age: 30
---
name: Doe
age: 25
`,
			},
			isResultWanted: true,
		},
		{
			name: "Passing Case with complex key",
			spec: Spec{
				File:   "test.yaml",
				Key:    "$.annotations['github.owner']",
				Engine: "yamlpath",
			},
			files: map[string]file{
				"test.yaml": {
					originalFilePath: "test.yaml",
					filePath:         "test.yaml",
				},
			},
			inputSourceValue: "olblak",
			mockedContents: map[string]string{
				"test.yaml": `---
annotations:
  github.owner: olblak
  repository: charts
`,
			},
			isResultWanted: true,
		},
		{
			name: "Passing Case",
			spec: Spec{
				File: "test.yaml",
				Key:  "$.github.owner",
			},
			files: map[string]file{
				"test.yaml": {
					originalFilePath: "test.yaml",
					filePath:         "test.yaml",
				},
			},
			inputSourceValue: "olblak",
			mockedContents: map[string]string{
				"test.yaml": `---
github:
  owner: olblak
  repository: charts
`,
			},
			isResultWanted: true,
		},
		{
			name: "Passing Case using yamlpath engine",
			spec: Spec{
				File:   "test.yaml",
				Key:    "$.github.owner",
				Engine: "yamlpath",
			},
			files: map[string]file{
				"test.yaml": {
					originalFilePath: "test.yaml",
					filePath:         "test.yaml",
				},
			},
			inputSourceValue: "olblak",
			mockedContents: map[string]string{
				"test.yaml": `---
github:
  owner: olblak
  repository: charts
`,
			},
			isResultWanted: true,
		},
		{
			name: "Passing case with one 'Files' no input source and only specified value",
			spec: Spec{
				Files: []string{
					"test.yaml",
				},
				Key:   "$.github.owner",
				Value: "olblak",
			},
			files: map[string]file{
				"test.yaml": {
					originalFilePath: "test.yaml",
					filePath:         "test.yaml",
				},
			},
			mockedContents: map[string]string{
				"test.yaml": `---
github:
  owner: olblak
  repository: charts
`,
			},
			isResultWanted: true,
		},
		{
			name: "Passing case with one 'Files' no input source and only specified value using yamlpath engine",
			spec: Spec{
				Files: []string{
					"test.yaml",
				},
				Key:    "$.github.owner",
				Value:  "olblak",
				Engine: "yamlpath",
			},
			files: map[string]file{
				"test.yaml": {
					originalFilePath: "test.yaml",
					filePath:         "test.yaml",
				},
			},
			mockedContents: map[string]string{
				"test.yaml": `---
github:
  owner: olblak
  repository: charts
`,
			},
			isResultWanted: true,
		},
		{
			name: "Validation with more than one 'Files'",
			spec: Spec{
				Files: []string{
					"test.yaml",
					"too-much.yaml",
				},
				Key:   "$.github.owner",
				Value: "olblak",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
				"too-much.yaml": {
					filePath:         "too-much.yaml",
					originalFilePath: "too-much.yaml",
				},
			},
			mockedContents: map[string]string{
				"test.yaml": `---
github:
  owner: olblak
  repository: charts
`,
				"too-much.yaml": `---
github:
  owner: olblak
  repository: charts
`,
			},
			isErrorWanted:  false,
			isResultWanted: true,
		},
		{
			name: "Validation with more than one 'Files' but only one success",
			spec: Spec{
				Files: []string{
					"test.yaml",
					"too-much.yaml",
				},
				Key:   "$.github.owner",
				Value: "olblak",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
				"too-much.yaml": {
					filePath:         "too-much.yaml",
					originalFilePath: "too-much.yaml",
				},
			},
			mockedContents: map[string]string{
				"test.yaml": `---
github:
  owner: olblak
  repository: charts
`,
				"too-much.yaml": `---
github:
  owner: john
  repository: charts
`,
			},
			isErrorWanted:  false,
			isResultWanted: false,
		},
		{
			name: "Validation with more than one 'Files' but only one success 2",
			spec: Spec{
				Files: []string{
					"test.yaml",
					"too-much.yaml",
				},
				Key:   "$.github.owner",
				Value: "olblak",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
				"too-much.yaml": {
					filePath:         "too-much.yaml",
					originalFilePath: "too-much.yaml",
				},
			},
			mockedContents: map[string]string{
				"test.yaml": `---
github:
  owner: olblak
  repository: charts
`,
				"too-much.yaml": `---
github:
  repository: charts
`,
			},
			isErrorWanted:  true,
			isResultWanted: false,
		},
		{
			name: "Failing case with invalid YAML content (dash)",
			spec: Spec{
				Files: []string{
					"test.yaml",
				},
				Key:   "$.github.owner",
				Value: "olblak",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
			},
			mockedContents: map[string]string{
				"test.yaml": `---
github:
  - owner: olblak
  repository: charts
`,
			},
			isErrorWanted: true,
		},
		{
			name: "Failing case with invalid YAML content (tabs)",
			spec: Spec{
				File: "test.yaml",
				Key:  "$.github.owner",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
			},
			inputSourceValue: "olblak",
			mockedContents: map[string]string{
				"test.yaml": `---
github:
	owner: olblak
  	repository: charts
`,
			},
			isErrorWanted: true,
		},
		{
			name: "Passing case with keyonly and input source",
			spec: Spec{
				File:    "test.yaml",
				Key:     "$.github.owner",
				KeyOnly: true,
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml"},
			},
			inputSourceValue: "olblak",
			mockedContents: map[string]string{
				"test.yaml": `---
github:
  owner: olblak
  repository: charts
`,
			},
			isResultWanted: true,
		},
		{
			name: "Passing case with keyonly and input source using yamlpath engine",
			spec: Spec{
				File:    "test.yaml",
				Key:     "$.github.owner",
				KeyOnly: true,
				Engine:  "yamlpath",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml"},
			},
			inputSourceValue: "olblak",
			mockedContents: map[string]string{
				"test.yaml": `---
github:
  owner: olblak
  repository: charts
`,
			},
			isResultWanted: true,
		},
		{
			name: "'No result' case with keyonly and input source",
			spec: Spec{
				File:    "test.yaml",
				Key:     "$.github.country",
				KeyOnly: true,
			},
			files: map[string]file{
				"test.yaml": {
					originalFilePath: "test.yaml",
					filePath:         "test.yaml",
				},
			},
			inputSourceValue: "",
			mockedContents: map[string]string{
				"test.yaml": `---
github:
  owner: olblak
  repository: charts
`,
			},
			isResultWanted: false,
		},
		{
			name: "'No result' case with keyonly and input source using yamlpath engine",
			spec: Spec{
				File:    "test.yaml",
				Key:     "$.github.country",
				KeyOnly: true,
				Engine:  "yamlpath",
			},
			files: map[string]file{
				"test.yaml": {
					originalFilePath: "test.yaml",
					filePath:         "test.yaml",
				},
			},
			inputSourceValue: "",
			mockedContents: map[string]string{
				"test.yaml": `---
github:
  owner: olblak
  repository: charts
`,
			},
			isResultWanted: false,
		},
		{
			name: "Validation error with both keyonly and specified value",
			spec: Spec{
				File:    "test.yaml",
				Key:     "$.github.owner",
				KeyOnly: true,
				Value:   "olblak",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
			},
			inputSourceValue: "",
			mockedContents: map[string]string{
				"test.yaml": `---
github:
  owner: olblak
  repository: charts
`,
			},
			isErrorWanted: true,
		},
		{
			name: "File does not exist",
			spec: Spec{
				File: "not_existing.yaml",
				Key:  "$.github.owner",
			},
			files: map[string]file{
				"not_existing.yaml": {
					filePath:         "not_existing.yaml",
					originalFilePath: "not_existing.yaml",
				},
			},
			mockedError:   fmt.Errorf("no such file or directory"),
			isErrorWanted: true,
		},
		{
			name: "'No result' case (key found but not the correct value)",
			spec: Spec{
				File: "test.yaml",
				Key:  "$.github.owner",
			},
			files: map[string]file{
				"test.yaml": {
					originalFilePath: "test.yaml",
					filePath:         "test.yaml",
				},
			},
			inputSourceValue: "asterix",
			mockedContents: map[string]string{
				"test.yaml": `---
github:
  owner: olblak
  repository: charts
`,
			},
			isResultWanted: false,
		},
		{
			name: "Failing case (key not found)",
			spec: Spec{
				File: "test.yaml",
				Key:  "$.github.admin",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
			},
			inputSourceValue: "asterix",
			mockedContents: map[string]string{
				"test.yaml": `---
github:
  owner: olblak
  repository: charts
`,
			},
			isErrorWanted: true,
		},
		{
			name: "Validation Failure with both source and specified value",
			spec: Spec{
				File:  "test.yaml",
				Key:   "$.github.owner",
				Value: "asterix",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
			},
			inputSourceValue: "olblak",
			isErrorWanted:    true,
		},
		{
			name: "Passing case with no input source and only specified value",
			spec: Spec{
				File:  "test.yaml",
				Key:   "$.github.owner",
				Value: "olblak",
			},
			files: map[string]file{
				"test.yaml": {
					originalFilePath: "test.yaml",
					filePath:         "test.yaml",
				},
			},
			mockedContents: map[string]string{
				"test.yaml": `---
github:
  owner: olblak
  repository: charts
`,
			},
			isResultWanted: true,
		},
		{
			name: "Passing case with one 'Files' input source and only specified value",
			spec: Spec{
				Files: []string{
					"test.yaml",
				},
				Key:    "repos[?(@.repository == 'website')].owner",
				Value:  "updatecli",
				Engine: "yamlpath",
			},
			files: map[string]file{
				"test.yaml": {
					originalFilePath: "test.yaml",
					filePath:         "test.yaml",
				},
			},
			mockedContents: map[string]string{
				"test.yaml": `---
repos:
  - owner: updatecli
    repository: website
  - owner: olblak
    repository: updatecli
`,
			},
			isResultWanted: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockedText := text.MockTextRetriever{
				Contents: tt.mockedContents,
				Err:      tt.mockedError,
			}

			y, err := New(tt.spec)
			y.contentRetriever = &mockedText
			y.files = tt.files

			assert.NoError(t, err)

			gotResult, _, gotErr := y.Condition(tt.inputSourceValue, nil)
			if tt.isErrorWanted {
				assert.Error(t, gotErr)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.isResultWanted, gotResult)
		})
	}
}
