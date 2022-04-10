package yaml

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/text"
)

func Test_Condition(t *testing.T) {
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
			name: "Passing Case",
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
				"test.yaml": `---
github:
  owner: olblak
  repository: charts
`,
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
				Key:   "github.owner",
				Value: "olblak",
			},
			files: map[string]string{
				"test.yaml":     "",
				"too-much.yaml": "",
			},
			isErrorWanted: true,
		},
		{
			name: "Failing case with invalid YAML content (dash)",
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
			name: "Passing case with keyonly and input source",
			spec: Spec{
				File:    "test.yaml",
				Key:     "github.owner",
				KeyOnly: true,
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
				Key:     "github.country",
				KeyOnly: true,
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
			wantedContents: map[string]string{
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
				Key:     "github.owner",
				KeyOnly: true,
				Value:   "olblak",
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
			isErrorWanted: true,
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
		{
			name: "'No result' case (key found but not the correct value)",
			spec: Spec{
				File: "test.yaml",
				Key:  "github.owner",
			},
			files: map[string]string{
				"test.yaml": "",
			},
			inputSourceValue: "asterix",
			mockedContents: map[string]string{
				"test.yaml": `---
github:
  owner: olblak
  repository: charts
`,
			},
			wantedContents: map[string]string{
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
				Key:  "github.admin",
			},
			files: map[string]string{
				"test.yaml": "",
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
				Key:   "github.owner",
				Value: "asterix",
			},
			files: map[string]string{
				"test.yaml": "",
			},
			inputSourceValue: "olblak",
			isErrorWanted:    true,
		},
		{
			name: "Passing case with no input source and only specified value",
			spec: Spec{
				File:  "test.yaml",
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
				"test.yaml": `---
github:
  owner: olblak
  repository: charts
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
			y := &Yaml{
				spec:             tt.spec,
				contentRetriever: &mockedText,
				files:            tt.files,
			}
			gotResult, gotErr := y.Condition(tt.inputSourceValue)
			if tt.isErrorWanted {
				assert.Error(t, gotErr)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.isResultWanted, gotResult)
			for filePath := range tt.files {
				assert.Equal(t, tt.wantedContents[filePath], mockedText.Contents[filePath])
			}
		})
	}
}

func Test_ConditionFromSCM(t *testing.T) {
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
		scm              scm.ScmHandler
	}{
		{
			name: "Passing case with no input source and only specified value",
			spec: Spec{
				File:  "test.yaml",
				Key:   "github.owner",
				Value: "olblak",
			},
			files: map[string]string{
				"/tmp/test.yaml": "",
			},
			scm: &scm.MockScm{
				WorkingDir: "/tmp",
			},
			mockedContents: map[string]string{
				"/tmp/test.yaml": `---
github:
  owner: olblak
  repository: charts
`,
			},
			wantedContents: map[string]string{
				"/tmp/test.yaml": `---
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
				Key:   "github.owner",
				Value: "olblak",
			},
			files: map[string]string{
				"/tmp/test.yaml": "",
			},
			scm: &scm.MockScm{
				WorkingDir: "/tmp",
			},
			mockedContents: map[string]string{
				"/tmp/test.yaml": `---
github:
  owner: olblak
  repository: charts
`,
			},
			wantedContents: map[string]string{
				"/tmp/test.yaml": `---
github:
  owner: olblak
  repository: charts
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
			y := &Yaml{
				spec:             tt.spec,
				contentRetriever: &mockedText,
				files:            tt.files,
			}
			gotResult, gotErr := y.ConditionFromSCM(tt.inputSourceValue, tt.scm)
			if tt.isErrorWanted {
				assert.Error(t, gotErr)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.isResultWanted, gotResult)
			for filePath := range tt.files {
				assert.Equal(t, tt.wantedContents[filePath], mockedText.Contents[filePath])
			}
		})
	}
}
