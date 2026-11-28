package yaml

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
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
			name: "Yamlpath passing case with multiple documents in a file, both input source and specified value (specified value should be used)",
			spec: Spec{
				Files: []string{
					"test.yaml",
				},
				Key:           "github.owner",
				Value:         "obiwankenobi",
				Comment:       "comment that should be added",
				Engine:        "yamlpath",
				DocumentIndex: ptrInt(1),
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
---
github:
  owner: asterix
  repository: charts
`,
			},
			wantedContents: map[string]string{
				"test.yaml": `---
github:
  owner: olblak
  repository: charts
---
github:
  owner: obiwankenobi # comment that should be added
  repository: charts
`,
			},
			wantedResult: true,
		},
		{
			name: "Passing case with multiple documents and no document index in a file, both input source and specified value (specified value should be used)",
			spec: Spec{
				Files: []string{
					"test.yaml",
				},
				Key:     "github.owner",
				Value:   "obiwankenobi",
				Comment: "comment that should be added",
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
---
github:
  owner: asterix
  repository: charts
`,
			},
			wantedContents: map[string]string{
				"test.yaml": `---
github:
  owner: obiwankenobi # comment that should be added
  repository: charts
---
github:
  owner: obiwankenobi # comment that should be added
  repository: charts
`,
			},
			wantedResult: true,
		},
		{
			name: "Passing case with multiple documents in a file, both input source and specified value (specified value should be used)",
			spec: Spec{
				Files: []string{
					"test.yaml",
				},
				Key:           "github.owner",
				Value:         "obiwankenobi",
				Comment:       "comment that should be added",
				DocumentIndex: ptrInt(1),
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
---
github:
  owner: asterix
  repository: charts
`,
			},
			wantedContents: map[string]string{
				"test.yaml": `---
github:
  owner: olblak
  repository: charts
---
github:
  owner: obiwankenobi # comment that should be added
  repository: charts
`,
			},
			wantedResult: true,
		},
		{
			name: "Passing case with 'Files' (one already up to date), both input source and specified value (specified value should be used) and updated comment",
			spec: Spec{
				Files: []string{
					"test.yaml",
					"bar.yaml",
				},
				Key:     "github.owner",
				Value:   "obiwankenobi",
				Comment: "comment that should be added",
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
			wantedContents: map[string]string{
				"test.yaml": `---
github:
  owner: obiwankenobi # comment that should be added
  repository: charts
`,
				// Please note that the comment shouldn't be added as the value wasn't changed
				"bar.yaml": `---
github:
  owner: obiwankenobi
  repository: charts
`,
			},
			wantedResult: true,
		},
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
		{
			name: "Passing case with a string that looks like a number",
			spec: Spec{
				File:  "melange.yaml",
				Key:   "package.version",
				Value: "0.8",
			},
			files: map[string]file{
				"melange.yaml": {
					filePath:         "melange.yaml",
					originalFilePath: "melange.yaml",
				},
			},
			inputSourceValue: "0.8",
			mockedContents: map[string]string{
				"melange.yaml": `---
package:
  version: "0.8"
`,
			},
			wantedContents: map[string]string{
				"melange.yaml": `---
package:
  version: "0.8"
`,
			},
			wantedResult: false,
		},
		{
			name: "Passing case with a string that starts with a @",
			spec: Spec{
				File:  "apko.yaml",
				Key:   "contents.repositories",
				Value: "@local foo/0.1.2/packages",
			},
			files: map[string]file{
				"apko.yaml": {
					filePath:         "apko.yaml",
					originalFilePath: "apko.yaml",
				},
			},
			inputSourceValue: "@local foo/0.1.2/packages",
			mockedContents: map[string]string{
				"apko.yaml": `---
contents:
  repositories: "@local foo/0.1.2/packages"
`,
			},
			wantedContents: map[string]string{
				"apko.yaml": `---
contents:
  repositories: "@local foo/0.1.2/packages"
`,
			},
			wantedResult: false,
		},
		{
			name: "Passing case with multiple keys updating different paths",
			spec: Spec{
				File: "test.yaml",
				Keys: []string{
					"$.image.tag",
					"$.sidecar.tag",
				},
				Value: "v1.2.3",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
			},
			inputSourceValue: "v1.0.0",
			mockedContents: map[string]string{
				"test.yaml": `---
image:
  tag: v1.0.0
sidecar:
  tag: v1.1.0
`,
			},
			wantedContents: map[string]string{
				"test.yaml": `---
image:
  tag: v1.2.3
sidecar:
  tag: v1.2.3
`,
			},
			wantedResult: true,
		},
		{
			name: "Passing case with multiple keys, some already up to date",
			spec: Spec{
				File: "test.yaml",
				Keys: []string{
					"$.app.version",
					"$.app.tag",
					"$.database.version",
				},
				Value: "v2.0.0",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
			},
			inputSourceValue: "v1.0.0",
			mockedContents: map[string]string{
				"test.yaml": `---
app:
  version: v1.5.0
  tag: v2.0.0
database:
  version: v1.8.0
`,
			},
			wantedContents: map[string]string{
				"test.yaml": `---
app:
  version: v2.0.0
  tag: v2.0.0
database:
  version: v2.0.0
`,
			},
			wantedResult: true,
		},
		{
			name: "Passing case with multiple keys using yamlpath engine",
			spec: Spec{
				File: "test.yaml",
				Keys: []string{
					"$.services[0].image",
					"$.services[1].image",
				},
				Value:  "nginx:1.21",
				Engine: "yamlpath",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
			},
			inputSourceValue: "nginx:1.20",
			mockedContents: map[string]string{
				"test.yaml": `---
services:
  - name: web
    image: nginx:1.20
  - name: api
    image: nginx:1.19
`,
			},
			wantedContents: map[string]string{
				"test.yaml": `---
services:
  - name: web
    image: nginx:1.21
  - name: api
    image: nginx:1.21
`,
			},
			wantedResult: true,
		},
		{
			name: "Folded scalar with identical value should not be rewritten (issue #8295)",
			spec: Spec{
				File:  "test.yaml",
				Key:   "$.jobs.build.steps[0].run",
				Value: "echo ${{ github.sha }}\necho done",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
			},
			inputSourceValue: "echo ${{ github.sha }}\necho done",
			mockedContents: map[string]string{
				"test.yaml": `jobs:
  build:
    steps:
      - run: >-
          echo ${{ github.sha }}

          echo done
`,
			},
			wantedContents: map[string]string{
				"test.yaml": `jobs:
  build:
    steps:
      - run: >-
          echo ${{ github.sha }}

          echo done
`,
			},
			wantedResult: false,
		},
		{
			name: "Folded scalar with changed value should be rewritten (issue #8295)",
			spec: Spec{
				File:  "test.yaml",
				Key:   "$.jobs.build.steps[0].run",
				Value: "echo updated\necho done",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
			},
			inputSourceValue: "echo updated\necho done",
			mockedContents: map[string]string{
				"test.yaml": `jobs:
  build:
    steps:
      - run: >-
          echo ${{ github.sha }}

          echo done
`,
			},
			wantedContents: map[string]string{
				"test.yaml": `jobs:
  build:
    steps:
      - run: |-
               echo updated
               echo done
`,
			},
			wantedResult: true,
		},
		{
			name: "Folded scalar sibling of an updated key should be preserved (issue #8295)",
			spec: Spec{
				File:    "test.yaml",
				Key:     "$.jobs.build.steps[0].uses",
				Comment: "v2.0.0",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
			},
			inputSourceValue: "actions/checkout@def456",
			mockedContents: map[string]string{
				"test.yaml": `jobs:
  build:
    environment: >-
      ${{
        github.ref == 'refs/heads/main' && 'dev' ||
        'unknown'
      }}
    steps:
      - uses: actions/checkout@abc123 # v1.0.0
`,
			},
			wantedContents: map[string]string{
				"test.yaml": `jobs:
  build:
    environment: >-
      ${{
        github.ref == 'refs/heads/main' && 'dev' ||
        'unknown'
      }}
    steps:
      - uses: actions/checkout@def456 # v2.0.0
`,
			},
			wantedResult: true,
		},
		{
			name: "Error case with multiple keys, one key not found",
			spec: Spec{
				File: "test.yaml",
				Keys: []string{
					"$.existing.key",
					"$.nonexistent.key",
				},
				Value: "newvalue",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
			},
			inputSourceValue: "oldvalue",
			mockedContents: map[string]string{
				"test.yaml": `---
existing:
  key: oldvalue
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

			y, err := New(tt.spec)
			y.contentRetriever = &mockedText
			y.files = tt.files

			assert.NoError(t, err)

			gotResult := result.Target{}
			gotErr := y.Target(context.Background(), tt.inputSourceValue, nil, pathresolver.Resolver{}, tt.dryRun, &gotResult)
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
		{
			name: "Passing case with 'Files' using file protocol, both input source and specified value (specified value should be used)",
			spec: Spec{
				Files: []string{
					"file://test.yaml",
					"file://bar.yaml",
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
			gotErr := y.Target(context.Background(), tt.inputSourceValue, tt.scm, pathresolver.New(tt.scm, ""), tt.dryRun, &gotResult)
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

// Test_TargetAnchorsAliases covers issue #11053: updating a key whose value is an
// *alias, or an &anchor, rewrites the anchored value and keeps both in place.
func Test_TargetAnchorsAliases(t *testing.T) {
	const compose = `x-image: &image nginx:1.29
services:
  web:
    image: *image
`
	const composeUpdated = `x-image: &image nginx:1.30
services:
  web:
    image: *image
`
	bothEngines := []string{"", "yamlpath"}

	tests := []struct {
		name          string
		engines       []string
		keys          []string
		value         string
		comment       string
		appendToArray bool
		createMissing bool
		documentIndex *int
		dryRun        bool
		content       string
		want          string
		wantChanged   bool
		wantInfo      string
		wantErr       bool
	}{
		{
			name:        "alias update rewrites the anchored value",
			engines:     bothEngines,
			keys:        []string{"$.services.web.image"},
			content:     compose,
			want:        composeUpdated,
			wantChanged: true,
			wantInfo:    "nginx:1.29",
		},
		{
			name:        "direct anchor update keeps the anchor",
			engines:     bothEngines,
			keys:        []string{"$.x-image"},
			content:     compose,
			want:        composeUpdated,
			wantChanged: true,
		},
		{
			name:    "several aliases of one anchor update it once",
			engines: bothEngines,
			keys:    []string{"$.services.web.image", "$.services.api.image"},
			content: `x-image: &image nginx:1.29
services:
  web:
    image: *image
  api:
    image: *image
`,
			want: `x-image: &image nginx:1.30
services:
  web:
    image: *image
  api:
    image: *image
`,
			wantChanged: true,
		},
		{
			name:    "wildcard over an alias and an inline value updates both",
			engines: []string{""},
			keys:    []string{"$.services[*].image"},
			content: `x-image: &image nginx:1.29
services:
  - image: *image
  - image: nginx:1.28
  - image: *image
`,
			want: `x-image: &image nginx:1.30
services:
  - image: *image
  - image: nginx:1.30
  - image: *image
`,
			wantChanged: true,
		},
		{
			name:     "alias already holding the value is a no-op",
			engines:  bothEngines,
			keys:     []string{"$.services.web.image"},
			content:  composeUpdated,
			want:     composeUpdated,
			wantInfo: "nginx:1.30",
		},
		{
			name:        "dry run reports the change without writing",
			engines:     bothEngines,
			keys:        []string{"$.services.web.image"},
			dryRun:      true,
			content:     compose,
			want:        compose,
			wantChanged: true,
			wantInfo:    "nginx:1.29",
		},
		{
			name:    "comments are kept and the new comment goes on the anchored value",
			engines: bothEngines,
			keys:    []string{"$.services.web.image"},
			comment: "updated",
			content: `# shared image
x-image: &image nginx:1.29
services:
  web:
    # web image
    image: *image
`,
			want: `# shared image
x-image: &image nginx:1.30 # updated
services:
  web:
    # web image
    image: *image
`,
			wantChanged: true,
		},
		{
			name:          "an alias binds to the anchor of its own document",
			engines:       bothEngines,
			keys:          []string{"$.services.web.image"},
			documentIndex: ptrInt(1),
			content: `x-image: &image nginx:1.29
services:
  web:
    image: *image
---
x-image: &image nginx:1.28
services:
  web:
    image: *image
`,
			want: `x-image: &image nginx:1.29
services:
  web:
    image: *image
---
x-image: &image nginx:1.30
services:
  web:
    image: *image
`,
			wantChanged: true,
		},
		{
			name:    "an alias binds to the closest preceding anchor",
			engines: []string{""},
			keys:    []string{"$.services.web.image"},
			content: `old: &image nginx:1.27
x-image: &image nginx:1.29
services:
  web:
    image: *image
`,
			want: `old: &image nginx:1.27
x-image: &image nginx:1.30
services:
  web:
    image: *image
`,
			wantChanged: true,
		},
		{
			name:    "dangling alias is an error",
			engines: bothEngines,
			keys:    []string{"$.image"},
			content: "image: *missing\n",
			wantErr: true,
		},
		{
			name:          "append to an aliased sequence updates the anchored sequence",
			engines:       []string{""},
			keys:          []string{"$.services.items"},
			value:         "c",
			appendToArray: true,
			content: `x-list: &list
  - a
  - b
services:
  items: *list
`,
			want: `x-list: &list
  - a
  - b
  - c
services:
  items: *list
`,
			wantChanged: true,
		},
		{
			name:          "append skips a value already held through an alias entry",
			engines:       []string{""},
			keys:          []string{"$.selected"},
			value:         "c",
			appendToArray: true,
			content: `item: &item c
selected: [*item]
`,
			want: `item: &item c
selected: [*item]
`,
		},
		{
			name:          "createmissingkey under a merge key adds a local override",
			engines:       []string{""},
			keys:          []string{"$.svc.image"},
			createMissing: true,
			content: `x-base: &base
  image: nginx:1.29
svc:
  <<: *base
`,
			want: `x-base: &base
  image: nginx:1.29
svc:
  <<: *base
  image: nginx:1.30
`,
			wantChanged: true,
		},
	}

	for _, tt := range tests {
		for _, engine := range tt.engines {
			name := tt.name
			if engine != "" {
				name += " (" + engine + ")"
			}
			t.Run(name, func(t *testing.T) {
				value := tt.value
				if value == "" {
					value = "nginx:1.30"
				}
				spec := Spec{
					File:          "test.yaml",
					Value:         value,
					Comment:       tt.comment,
					Engine:        engine,
					AppendToArray: tt.appendToArray,
					DocumentIndex: tt.documentIndex,

					CreateMissingKey: tt.createMissing,
				}
				if len(tt.keys) == 1 {
					spec.Key = tt.keys[0]
				} else {
					spec.Keys = tt.keys
				}

				mockedText := text.MockTextRetriever{Contents: map[string]string{"test.yaml": tt.content}}
				y, err := New(spec)
				require.NoError(t, err)
				y.contentRetriever = &mockedText
				y.files = map[string]file{"test.yaml": {filePath: "test.yaml", originalFilePath: "test.yaml"}}
				defer os.Remove("test.yaml")

				gotResult := result.Target{}
				gotErr := y.Target(context.Background(), "", nil, pathresolver.Resolver{}, tt.dryRun, &gotResult)
				if tt.wantErr {
					assert.Error(t, gotErr)
					return
				}

				require.NoError(t, gotErr)
				assert.Equal(t, tt.wantChanged, gotResult.Changed)
				assert.Equal(t, tt.want, mockedText.Contents["test.yaml"])
				if tt.wantInfo != "" {
					assert.Equal(t, tt.wantInfo, gotResult.Information)
				}
			})
		}
	}
}
