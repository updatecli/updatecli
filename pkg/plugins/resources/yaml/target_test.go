package yaml

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
)

func Test_Target(t *testing.T) {
	tests := []struct {
		name             string
		spec             Spec
		files            map[string]file
		inputSourceValue string
		mockedContents   map[string]string
		mockedError      error
		wantedContents   map[string]string
		wantedResult     bool
		wantedError      bool
		dryRun           bool
	}{
		{
			name: "Passing case with both complex input source and specified value (specified value should be used)",
			spec: Spec{
				File:   "test.yaml",
				Key:    "$.annotations['github.owner']",
				Value:  "obiwankenobi",
				Engine: "yamlpath",
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
## This is a comment that should be preserved
annotations:
  github.owner: olblak
  repository: charts
`,
			},
			// Note: the re-encoded file doesn't contain any separator anymore
			wantedContents: map[string]string{
				"test.yaml": `---
## This is a comment that should be preserved
annotations:
  github.owner: obiwankenobi
  repository: charts
`,
			},
			wantedResult: true,
		},
		{
			name: "Passing case with both input source and specified value (specified value should be used)",
			spec: Spec{
				File:  "test.yaml",
				Key:   "github.owner",
				Value: "obiwankenobi",
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
			// Note: the re-encoded file doesn't contain any separator anymore
			wantedContents: map[string]string{
				"test.yaml": `---
github:
  owner: obiwankenobi
  repository: charts
`,
			},
			wantedResult: true,
		},
		{
			name: "Passing case with 'Files', both input source and specified value (specified value should be used)",
			spec: Spec{
				Files: []string{
					"test.yaml",
					"bar.yaml",
				},
				Key:   "github.owner",
				Value: "obiwankenobi",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
				"bar.yaml": {
					filePath:         "bar.yaml",
					originalFilePath: "bar.yaml",
				},
			},
			inputSourceValue: "olblak",
			mockedContents: map[string]string{
				"test.yaml": `---
github:
  owner: olblak
  repository: charts
`,
				"bar.yaml": `---
github:
  owner: asterix
  repository: charts
`,
			},
			// Note: the updated files don't contain separator anymore
			wantedContents: map[string]string{
				"test.yaml": `---
github:
  owner: obiwankenobi
  repository: charts
`,
				"bar.yaml": `---
github:
  owner: obiwankenobi
  repository: charts
`,
			},
			wantedResult: true,
		},
		{
			name: "Passing case with 'Files' (one already up to date), both input source and specified value (specified value should be used)",
			spec: Spec{
				Files: []string{
					"test.yaml",
					"bar.yaml",
				},
				Key:   "github.owner",
				Value: "obiwankenobi",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
				"bar.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "bar.yaml",
				},
			},
			inputSourceValue: "olblak",
			mockedContents: map[string]string{
				"test.yaml": `---
github:
  owner: olblak
  repository: charts
`,
				"bar.yaml": `---
github:
  owner: obiwankenobi
  repository: charts
`,
			},
			// Note: the updated file doesn't contain separator anymore
			wantedContents: map[string]string{
				"test.yaml": `---
github:
  owner: obiwankenobi
  repository: charts
`,
				"bar.yaml": `---
github:
  owner: obiwankenobi
  repository: charts
`,
			},
			wantedResult: true,
		},
		{
			name: "Validation failure with an https:// URL instead of a file",
			spec: Spec{
				File:  "https://github.com/foo.yaml",
				Key:   "github.owner",
				Value: "obiwankenobi",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
			},
			wantedResult: false,
			wantedError:  true,
		},
		{
			name: "Passing: file already up to date",
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
				"test.yaml": `github:
  owner: olblak
  repository: charts
`,
			},
			wantedContents: map[string]string{
				"test.yaml": `github:
  owner: olblak
  repository: charts
`,
			},
			wantedResult: false,
			wantedError:  false,
		},
		{
			name: "Provided key does not exist",
			spec: Spec{
				File:  "test.yaml",
				Key:   "github.ship",
				Value: "obiwankenobi",
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
			wantedResult: false,
			wantedError:  true,
		},
		{
			name: "Invalid YAML file",
			spec: Spec{
				File:  "test.yaml",
				Key:   "github.ship",
				Value: "obiwankenobi",
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
			wantedResult: false,
			wantedError:  true,
		},
		{
			name: "Passing case with a comment that is added",
			spec: Spec{
				File:    "test.yaml",
				Key:     "github.owner",
				Value:   "obiwankenobi",
				Comment: "comment that should be added",
				Engine:  "yamlpath",
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
`,
			},
			// Note: the re-encoded file doesn't contain any separator anymore
			wantedContents: map[string]string{
				"test.yaml": `---
github:
  owner: obiwankenobi # comment that should be added
`,
			},
			wantedResult: true,
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

			gotResult := result.Target{}
			gotErr := y.Target(tt.inputSourceValue, nil, tt.dryRun, &gotResult)
			if tt.wantedError {
				assert.Error(t, gotErr)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.wantedResult, gotResult.Changed)

			for filePath := range y.files {
				defer os.Remove(filePath)
				assert.Equal(t, tt.wantedContents[filePath], mockedText.Contents[filePath])
			}
		})
	}
}

func Test_TargetFromSCM(t *testing.T) {
	tests := []struct {
		name             string
		spec             Spec
		files            map[string]file
		scm              scm.ScmHandler
		inputSourceValue string
		mockedContents   map[string]string
		mockedError      error
		wantedFiles      []string
		wantedContents   map[string]string
		wantedResult     bool
		wantedError      bool
		dryRun           bool
	}{
		{
			name: "Passing case with both input source and specified value (specified value should be used)",
			spec: Spec{
				File:  "test.yaml",
				Key:   "github.owner",
				Value: "obiwankenobi",
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
			wantedFiles: []string{
				"test.yaml",
			},
			// Note: the re-encoded file doesn't contain any separator anymore
			wantedContents: map[string]string{
				"test.yaml": `---
github:
  owner: obiwankenobi
  repository: charts
`,
			},
			wantedResult: true,
		},
		{
			name: "Passing case with 'Files', both input source and specified value (specified value should be used)",
			spec: Spec{
				Files: []string{
					"test.yaml",
					"bar.yaml",
				},
				Key:   "github.owner",
				Value: "obiwankenobi",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
				"bar.yaml": {
					filePath:         "bar.yaml",
					originalFilePath: "bar.yaml",
				},
			},
			inputSourceValue: "olblak",
			mockedContents: map[string]string{
				"test.yaml": `---
github:
  owner: olblak
  repository: charts
`,
				"bar.yaml": `---
github:
  owner: asterix
  repository: charts
`,
			},
			// returned files are sorted
			wantedFiles: []string{
				"bar.yaml",
				"test.yaml",
			},
			// Note: the updated files don't contain separator anymore
			wantedContents: map[string]string{
				"test.yaml": `---
github:
  owner: obiwankenobi
  repository: charts
`,
				"bar.yaml": `---
github:
  owner: obiwankenobi
  repository: charts
`,
			},
			wantedResult: true,
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

			gotResult := result.Target{}
			gotErr := y.Target(tt.inputSourceValue, tt.scm, tt.dryRun, &gotResult)
			if tt.wantedError {
				assert.Error(t, gotErr)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.wantedResult, gotResult.Changed)
			assert.Equal(t, tt.wantedFiles, gotResult.Files)

			for filePath := range y.files {
				defer os.Remove(filePath)
				assert.Equal(t, tt.wantedContents[filePath], mockedText.Contents[filePath])
			}
		})
	}
}
