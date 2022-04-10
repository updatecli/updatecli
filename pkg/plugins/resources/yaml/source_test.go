package yaml

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/text"
)

func Test_Source(t *testing.T) {
	tests := []struct {
		name             string
		spec             Spec
		files            map[string]string
		inputSourceValue string
		mockedContents   map[string]string
		mockedError      error
		wantedContents   map[string]string
		isResultWanted   bool
		isErrorWanted    bool
	}{
		{
			name: "Passing Case with 'File'",
			spec: Spec{
				File: "test.yaml",
				Key:  "github.owner",
			},
			files: map[string]string{
				"test.yaml": "",
			},
			inputSourceValue: "olblak",
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
				Key: "github.owner",
			},
			files: map[string]string{
				"test.yaml": "",
			},
			inputSourceValue: "olblak",
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
				Key: "github.owner",
			},
			files: map[string]string{
				"test.yaml":     "",
				"too-much.yaml": "",
			},
			isErrorWanted: true,
		},
		{
			name: "Passing case with one 'Files' (and warning 'Value')",
			spec: Spec{
				Files: []string{
					"test.yaml",
				},
				Key:   "github.owner",
				Value: "olblak",
			},
			files: map[string]string{
				"test.yaml": "",
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
				Key:  "github.owner",
			},
			files: map[string]string{
				"test.yaml": "",
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
			name: "'No result' case with key not found",
			spec: Spec{
				File: "test.yaml",
				Key:  "github.country",
			},
			files: map[string]string{
				"test.yaml": "",
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
			isErrorWanted:  false,
		},
		{
			name: "File does not exist",
			spec: Spec{
				File: "not_existing.yaml",
			},
			files: map[string]string{
				"not_existing.yaml": "",
			},
			mockedError:   fmt.Errorf("no such file or directory"),
			isErrorWanted: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockedText := text.MockTextRetriever{
				Contents: tt.mockedContents,
				Err:      tt.mockedError,
			}
			y := &Yaml{
				spec:             tt.spec,
				contentRetriever: &mockedText,
				files:            tt.files,
			}

			// Looping on the only filePath in 'files'
			for filePath := range y.files {
				source, gotErr := y.Source(filePath)
				if tt.isErrorWanted {
					assert.Error(t, gotErr)
					return
				}

				require.NoError(t, gotErr)
				assert.Equal(t, tt.wantedContents[filePath], source)
			}
		})
	}
}
