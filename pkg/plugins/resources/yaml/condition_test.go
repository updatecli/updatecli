package yaml

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
)

func Test_Condition(t *testing.T) {
	tests := []struct {
		name             string
		spec             Spec
		files            map[string]file
		inputSourceValue string
		mockedContents   map[string]string
		mockedError      error
		wantedContents   map[string]string
		isResultWanted   bool
		isErrorWanted    bool
	}{
		{
			name: "Passing Case with complex key",
			spec: Spec{
				File: "test.yaml",
				Key:  "annotations.'github.owner'",
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
			wantedContents: map[string]string{
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
				Key:  "github.owner",
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
			name: "Failing case with invalid YAML content (dash)",
			spec: Spec{
				Files: []string{
					"test.yaml",
				},
				Key:   "github.owner",
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
				Key:  "github.owner",
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
				Key:     "github.owner",
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
				Key:  "github.owner",
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
				Key:   "github.owner",
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
				Key:   "github.owner",
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
			gotResult := result.Condition{}
			gotErr := y.Condition(tt.inputSourceValue, nil, &gotResult)
			if tt.isErrorWanted {
				assert.Error(t, gotErr)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.isResultWanted, gotResult.Pass)
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
		files            map[string]file
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
			files: map[string]file{
				"/tmp/test.yaml": {
					filePath:         "/tmp/test.yaml",
					originalFilePath: "/tmp/test.yaml",
				},
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
			files: map[string]file{
				"/tmp/test.yaml": {
					filePath:         "/tmp/test.yaml",
					originalFilePath: "/tmp/test.yaml",
				},
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
			gotResult := result.Condition{}
			gotErr := y.Condition(tt.inputSourceValue, tt.scm, &gotResult)
			if tt.isErrorWanted {
				assert.Error(t, gotErr)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.isResultWanted, gotResult.Pass)
			for filePath := range tt.files {
				assert.Equal(t, tt.wantedContents[filePath], mockedText.Contents[filePath])
			}
		})
	}
}
