package yaml

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
)

func Test_Source(t *testing.T) {
	tests := []struct {
		name           string
		spec           Spec
		files          map[string]file
		mockedContents map[string]string
		mockedError    error
		wantedContents map[string]string
		isResultWanted bool
		isErrorWanted  bool
	}{
		{
			name: "Passing Case with 'File' and complex key",
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
			mockedContents: map[string]string{
				"test.yaml": `---
annotations:
  github.owner: olblak
  repository: charts
`,
			},
			wantedContents: map[string]string{
				"test.yaml": "olblak",
			},
			isResultWanted: true,
		},
		{
			name: "Passing Case with 'File'",
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
			mockedContents: map[string]string{
				"test.yaml": `---
github:
  owner: olblak
  repository: charts
`,
			},
			wantedContents: map[string]string{
				"test.yaml": "olblak",
			},
			isResultWanted: true,
		},
		{
			name: "Passing Case with 'Files'",
			spec: Spec{
				Files: []string{
					"test.yaml",
				},
				Key: "$.github.owner",
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
  owner: olblak
  repository: charts
`,
			},
			wantedContents: map[string]string{
				"test.yaml": "olblak",
			},
			isResultWanted: true,
		},
		{
			name: "Validation error with more than one 'Files'",
			spec: Spec{
				Files: []string{
					"test.yaml",
					"too-much.yaml",
				},
				Key: "$.github.owner",
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
			isErrorWanted: true,
		},
		{
			name: "Passing case with one 'Files' (and warning 'Value')",
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
					originalFilePath: "test.yaml"},
			},
			mockedContents: map[string]string{
				"test.yaml": `---
github:
  owner: olblak
  repository: charts
`,
			},
			wantedContents: map[string]string{
				"test.yaml": "olblak",
			},
			isResultWanted: true,
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
			name: "'No result' case with key not found",
			spec: Spec{
				File: "test.yaml",
				Key:  "$.github.country",
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
  owner: olblak
  repository: charts
`,
			},
			isResultWanted: false,
			isErrorWanted:  true,
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
			name: "Passing Case with 'Files' array entry",
			spec: Spec{
				Files: []string{
					"test.yaml",
				},
				Key:    "$.repos[?(@.repository == 'website')].owner",
				Engine: "yamlpath",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
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
			wantedContents: map[string]string{
				"test.yaml": "updatecli",
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

			// Looping on the only filePath in 'files'
			for filePath := range y.files {
				gotResult := result.Source{}

				gotErr := y.Source("", &gotResult)
				if tt.isErrorWanted {
					assert.Error(t, gotErr)
					return
				}

				require.NoError(t, gotErr)
				assert.Equal(t, tt.wantedContents[filePath], gotResult.Information)
			}
		})
	}
}
