package yaml

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/text"
)

func Test_Target(t *testing.T) {
	tests := []struct {
		name             string
		spec             Spec
		files            map[string]string
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
				File:  "test.yaml",
				Key:   "annotations.github\\.owner",
				Value: "obiwankenobi",
			},
			files: map[string]string{
				"test.yaml": "",
			},
			inputSourceValue: "olblak",
			mockedContents: map[string]string{
				"test.yaml": `---
annotations:
    github.owner: olblak
    repository: charts
`,
			},
			// Note: the re-encoded file doesn't contain any separator anymore
			wantedContents: map[string]string{
				"test.yaml": `annotations:
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
			// Note: the re-encoded file doesn't contain any separator anymore
			wantedContents: map[string]string{
				"test.yaml": `github:
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
			files: map[string]string{
				"test.yaml": "",
				"bar.yaml":  "",
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
				"test.yaml": `github:
    owner: obiwankenobi
    repository: charts
`,
				"bar.yaml": `github:
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
			files: map[string]string{
				"test.yaml": "",
				"bar.yaml":  "",
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
				"test.yaml": `github:
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
			files: map[string]string{
				"test.yaml": "",
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
			files: map[string]string{
				"test.yaml": "",
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
			wantedResult: false,
			wantedError:  true,
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
			gotResult, gotErr := y.Target(tt.inputSourceValue, tt.dryRun)
			if tt.wantedError {
				assert.Error(t, gotErr)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.wantedResult, gotResult)

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
		files            map[string]string
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
			files: map[string]string{
				"/tmp/test.yaml": "",
			},
			inputSourceValue: "olblak",
			mockedContents: map[string]string{
				"/tmp/test.yaml": `---
github:
  owner: olblak
  repository: charts
`,
			},
			wantedFiles: []string{
				"/tmp/test.yaml",
			},
			// Note: the re-encoded file doesn't contain any separator anymore
			wantedContents: map[string]string{
				"/tmp/test.yaml": `github:
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
			files: map[string]string{
				"/tmp/test.yaml": "",
				"/tmp/bar.yaml":  "",
			},
			inputSourceValue: "olblak",
			mockedContents: map[string]string{
				"/tmp/test.yaml": `---
github:
  owner: olblak
  repository: charts
`,
				"/tmp/bar.yaml": `---
github:
  owner: asterix
  repository: charts
`,
			},
			// returned files are sorted
			wantedFiles: []string{
				"/tmp/bar.yaml",
				"/tmp/test.yaml",
			},
			// Note: the updated files don't contain separator anymore
			wantedContents: map[string]string{
				"/tmp/test.yaml": `github:
    owner: obiwankenobi
    repository: charts
`,
				"/tmp/bar.yaml": `github:
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
			y := &Yaml{
				spec:             tt.spec,
				contentRetriever: &mockedText,
				files:            tt.files,
			}
			gotResult, gotFiles, _, gotErr := y.TargetFromSCM(tt.inputSourceValue, tt.scm, tt.dryRun)
			if tt.wantedError {
				assert.Error(t, gotErr)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.wantedResult, gotResult)
			assert.Equal(t, tt.wantedFiles, gotFiles)

			for filePath := range y.files {
				defer os.Remove(filePath)
				assert.Equal(t, tt.wantedContents[filePath], mockedText.Contents[filePath])
			}
		})
	}
}
