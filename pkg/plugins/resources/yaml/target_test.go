package yaml

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/goccy/go-yaml/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
)

func Test_Target(t *testing.T) {
	tests := []struct {
		name              string
		spec              Spec
		files             map[string]file
		inputSourceValue  string
		mockedContents    map[string]string
		mockedError       error
		wantedContents    map[string]string
		wantedResult      bool
		wantedError       bool
		wantedInformation string
		dryRun            bool
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
		{
			name: "Updating an aliased compose image preserves the alias (issue #11053)",
			spec: Spec{
				File:  "test.yaml",
				Key:   "$.services.web.image",
				Value: "nginx:1.30",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
			},
			inputSourceValue: "nginx:1.29",
			mockedContents: map[string]string{
				"test.yaml": `x-image: &image nginx:1.29
services:
  web:
    image: *image
`,
			},
			wantedContents: map[string]string{
				"test.yaml": `x-image: &image nginx:1.30
services:
  web:
    image: *image
`,
			},
			wantedResult: true,
		},
		{
			name: "Updating an aliased compose image preserves the alias with the yamlpath engine (issue #11053)",
			spec: Spec{
				File:   "test.yaml",
				Key:    "$.services.web.image",
				Value:  "nginx:1.30",
				Engine: "yamlpath",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
			},
			inputSourceValue: "nginx:1.29",
			mockedContents: map[string]string{
				"test.yaml": `x-image: &image nginx:1.29
services:
  web:
    image: *image
`,
			},
			wantedContents: map[string]string{
				"test.yaml": `x-image: &image nginx:1.30
services:
  web:
    image: *image
`,
			},
			wantedResult: true,
		},
		{
			name: "Already updated aliased compose image is a no-op (issue #11053)",
			spec: Spec{
				File:  "test.yaml",
				Key:   "$.services.web.image",
				Value: "nginx:1.30",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
			},
			inputSourceValue: "nginx:1.29",
			mockedContents: map[string]string{
				"test.yaml": `x-image: &image nginx:1.30
services:
  web:
    image: *image
`,
			},
			wantedContents: map[string]string{
				"test.yaml": `x-image: &image nginx:1.30
services:
  web:
    image: *image
`,
			},
			wantedResult: false,
		},
		{
			name: "Already updated aliased compose image is a no-op with the yamlpath engine (issue #11053)",
			spec: Spec{
				File:   "test.yaml",
				Key:    "$.services.web.image",
				Value:  "nginx:1.30",
				Engine: "yamlpath",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
			},
			inputSourceValue: "nginx:1.29",
			mockedContents: map[string]string{
				"test.yaml": `x-image: &image nginx:1.30
services:
  web:
    image: *image
`,
			},
			wantedContents: map[string]string{
				"test.yaml": `x-image: &image nginx:1.30
services:
  web:
    image: *image
`,
			},
			wantedResult: false,
		},
		{
			name: "Dry run reports the aliased image value without writing (issue #11053)",
			spec: Spec{
				File:  "test.yaml",
				Key:   "$.services.web.image",
				Value: "nginx:1.30",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
			},
			inputSourceValue: "nginx:1.29",
			mockedContents: map[string]string{
				"test.yaml": `x-image: &image nginx:1.29
services:
  web:
    image: *image
`,
			},
			wantedContents: map[string]string{
				"test.yaml": `x-image: &image nginx:1.29
services:
  web:
    image: *image
`,
			},
			wantedResult:      true,
			wantedInformation: "nginx:1.29",
			dryRun:            true,
		},
		{
			name: "Dry run reports the aliased image value without writing with the yamlpath engine (issue #11053)",
			spec: Spec{
				File:   "test.yaml",
				Key:    "$.services.web.image",
				Value:  "nginx:1.30",
				Engine: "yamlpath",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
			},
			inputSourceValue: "nginx:1.29",
			mockedContents: map[string]string{
				"test.yaml": `x-image: &image nginx:1.29
services:
  web:
    image: *image
`,
			},
			wantedContents: map[string]string{
				"test.yaml": `x-image: &image nginx:1.29
services:
  web:
    image: *image
`,
			},
			wantedResult:      true,
			wantedInformation: "nginx:1.29",
			dryRun:            true,
		},
		{
			name: "Comment is attached to the updated anchor value (issue #11053)",
			spec: Spec{
				File:    "test.yaml",
				Key:     "$.services.web.image",
				Value:   "nginx:1.30",
				Comment: "comment that should be added",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
			},
			inputSourceValue: "nginx:1.29",
			mockedContents: map[string]string{
				"test.yaml": `x-image: &image nginx:1.29
services:
  web:
    image: *image
`,
			},
			wantedContents: map[string]string{
				"test.yaml": `x-image: &image nginx:1.30 # comment that should be added
services:
  web:
    image: *image
`,
			},
			wantedResult: true,
		},
		{
			name: "Comment is attached to the updated anchor value with the yamlpath engine (issue #11053)",
			spec: Spec{
				File:    "test.yaml",
				Key:     "$.services.web.image",
				Value:   "nginx:1.30",
				Comment: "comment that should be added",
				Engine:  "yamlpath",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
			},
			inputSourceValue: "nginx:1.29",
			mockedContents: map[string]string{
				"test.yaml": `x-image: &image nginx:1.29
services:
  web:
    image: *image
`,
			},
			wantedContents: map[string]string{
				"test.yaml": `x-image: &image nginx:1.30 # comment that should be added
services:
  web:
    image: *image
`,
			},
			wantedResult: true,
		},
		{
			name: "Direct update of an anchored key preserves the anchor (issue #11053)",
			spec: Spec{
				File:  "test.yaml",
				Key:   "$.x-image",
				Value: "nginx:1.30",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
			},
			inputSourceValue: "nginx:1.29",
			mockedContents: map[string]string{
				"test.yaml": `x-image: &image nginx:1.29
services:
  web:
    image: *image
`,
			},
			wantedContents: map[string]string{
				"test.yaml": `x-image: &image nginx:1.30
services:
  web:
    image: *image
`,
			},
			wantedResult: true,
		},
		{
			name: "Direct update of an anchored key preserves the anchor with the yamlpath engine (issue #11053)",
			spec: Spec{
				File:   "test.yaml",
				Key:    "$.x-image",
				Value:  "nginx:1.30",
				Engine: "yamlpath",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
			},
			inputSourceValue: "nginx:1.29",
			mockedContents: map[string]string{
				"test.yaml": `x-image: &image nginx:1.29
services:
  web:
    image: *image
`,
			},
			wantedContents: map[string]string{
				"test.yaml": `x-image: &image nginx:1.30
services:
  web:
    image: *image
`,
			},
			wantedResult: true,
		},
		{
			name: "Multiple aliases of one anchor are updated once (issue #11053)",
			spec: Spec{
				File: "test.yaml",
				Keys: []string{
					"$.services.web.image",
					"$.services.api.image",
				},
				Value: "nginx:1.30",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
			},
			inputSourceValue: "nginx:1.29",
			mockedContents: map[string]string{
				"test.yaml": `x-image: &image nginx:1.29
services:
  web:
    image: *image
  api:
    image: *image
`,
			},
			wantedContents: map[string]string{
				"test.yaml": `x-image: &image nginx:1.30
services:
  web:
    image: *image
  api:
    image: *image
`,
			},
			wantedResult: true,
		},
		{
			name: "Multiple aliases of one anchor are updated once with the yamlpath engine (issue #11053)",
			spec: Spec{
				File: "test.yaml",
				Keys: []string{
					"$.services.web.image",
					"$.services.api.image",
				},
				Value:  "nginx:1.30",
				Engine: "yamlpath",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
			},
			inputSourceValue: "nginx:1.29",
			mockedContents: map[string]string{
				"test.yaml": `x-image: &image nginx:1.29
services:
  web:
    image: *image
  api:
    image: *image
`,
			},
			wantedContents: map[string]string{
				"test.yaml": `x-image: &image nginx:1.30
services:
  web:
    image: *image
  api:
    image: *image
`,
			},
			wantedResult: true,
		},
		{
			name: "Wildcard over aliased images updates the anchor once (issue #11053)",
			spec: Spec{
				File:  "test.yaml",
				Key:   "$.services[*].image",
				Value: "nginx:1.30",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
			},
			inputSourceValue: "nginx:1.29",
			mockedContents: map[string]string{
				"test.yaml": `x-image: &image nginx:1.29
services:
  - image: *image
  - image: *image
`,
			},
			wantedContents: map[string]string{
				"test.yaml": `x-image: &image nginx:1.30
services:
  - image: *image
  - image: *image
`,
			},
			wantedResult: true,
		},
		{
			name: "Mixed wildcard updates the aliased anchor and the inline image (issue #11053)",
			spec: Spec{
				File:  "test.yaml",
				Key:   "$.services[*].image",
				Value: "nginx:1.30",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
			},
			inputSourceValue: "nginx:1.29",
			mockedContents: map[string]string{
				"test.yaml": `x-image: &image nginx:1.29
services:
  - image: *image
  - image: nginx:1.28
`,
			},
			wantedContents: map[string]string{
				"test.yaml": `x-image: &image nginx:1.30
services:
  - image: *image
  - image: nginx:1.30
`,
			},
			wantedResult: true,
		},
		{
			name: "Chained aliases update the base anchor (issue #11053)",
			spec: Spec{
				File:  "test.yaml",
				Key:   "$.web.image",
				Value: "nginx:1.30",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
			},
			inputSourceValue: "nginx:1.29",
			mockedContents: map[string]string{
				"test.yaml": `base: &base nginx:1.29
image: &image
  *base
web:
  image: *image
`,
			},
			wantedContents: map[string]string{
				"test.yaml": `base: &base nginx:1.30
image: &image *base
web:
  image: *image
`,
			},
			wantedResult: true,
		},
		{
			name: "Compose merge key updates the anchored image (issue #11053)",
			spec: Spec{
				File:  "test.yaml",
				Key:   "$.services.web.image",
				Value: "nginx:1.30",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
			},
			inputSourceValue: "nginx:1.29",
			mockedContents: map[string]string{
				"test.yaml": `x-common: &common
  image: nginx:1.29
  restart: always
services:
  web:
    <<: *common
    ports:
      - "80:80"
`,
			},
			wantedContents: map[string]string{
				"test.yaml": `x-common: &common
  image: nginx:1.30
  restart: always
services:
  web:
    <<: *common
    ports:
      - "80:80"
`,
			},
			wantedResult: true,
		},
		{
			name: "Mapping alias updates the anchored image (issue #11053)",
			spec: Spec{
				File:  "test.yaml",
				Key:   "$.services.api.image",
				Value: "nginx:1.30",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
			},
			inputSourceValue: "nginx:1.29",
			mockedContents: map[string]string{
				"test.yaml": `services:
  web: &web
    image: nginx:1.29
  api: *web
`,
			},
			wantedContents: map[string]string{
				"test.yaml": `services:
  web: &web
    image: nginx:1.30
  api: *web
`,
			},
			wantedResult: true,
		},
		{
			name: "Append to an aliased sequence updates the anchored sequence (issue #11053)",
			spec: Spec{
				File:          "test.yaml",
				Key:           "$.services.items",
				Value:         "c",
				AppendToArray: true,
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
			},
			inputSourceValue: "c",
			mockedContents: map[string]string{
				"test.yaml": `x-list: &list
  - a
  - b
services:
  items: *list
`,
			},
			wantedContents: map[string]string{
				"test.yaml": `x-list: &list
  - a
  - b
  - c
services:
  items: *list
`,
			},
			wantedResult: true,
		},
		{
			name: "Anchors are scoped per document (issue #11053)",
			spec: Spec{
				File:  "test.yaml",
				Key:   "$.services.web.image",
				Value: "nginx:1.30",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
			},
			inputSourceValue: "nginx:1.29",
			mockedContents: map[string]string{
				"test.yaml": `x-image: &image nginx:1.29
services:
  web:
    image: *image
---
x-image: &image alpine:3
services:
  web:
    image: *image
`,
			},
			wantedContents: map[string]string{
				"test.yaml": `x-image: &image nginx:1.30
services:
  web:
    image: *image
---
x-image: &image nginx:1.30
services:
  web:
    image: *image
`,
			},
			wantedResult: true,
		},
		{
			name: "Dangling alias returns an error (issue #11053)",
			spec: Spec{
				File:  "test.yaml",
				Key:   "$.image",
				Value: "nginx:1.30",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
			},
			inputSourceValue: "nginx:1.29",
			mockedContents: map[string]string{
				"test.yaml": `image: *missing
`,
			},
			wantedError: true,
		},
		{
			name: "Dangling alias returns an error with the yamlpath engine (issue #11053)",
			spec: Spec{
				File:   "test.yaml",
				Key:    "$.image",
				Value:  "nginx:1.30",
				Engine: "yamlpath",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
			},
			inputSourceValue: "nginx:1.29",
			mockedContents: map[string]string{
				"test.yaml": `image: *missing
`,
			},
			wantedError: true,
		},
		{
			name: "Mixed wildcard with inline and merged entry updates both (issue #11053)",
			spec: Spec{
				File:  "test.yaml",
				Key:   "$.services[*].image",
				Value: "nginx:1.30",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
			},
			inputSourceValue: "nginx:1.29",
			mockedContents: map[string]string{
				"test.yaml": `x-common: &common
  image: nginx:1.29
services:
  - image: nginx:1.28
  - <<: *common
`,
			},
			wantedContents: map[string]string{
				"test.yaml": `x-common: &common
  image: nginx:1.30
services:
  - image: nginx:1.30
  - <<: *common
`,
			},
			wantedResult: true,
		},
		{
			name: "Mixed wildcard with inline and mapping alias updates both (issue #11053)",
			spec: Spec{
				File:  "test.yaml",
				Key:   "$.services[*].image",
				Value: "nginx:1.30",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
			},
			inputSourceValue: "nginx:1.29",
			mockedContents: map[string]string{
				"test.yaml": `x-web: &web
  image: nginx:1.29
services:
  - image: nginx:1.28
  - *web
`,
			},
			wantedContents: map[string]string{
				"test.yaml": `x-web: &web
  image: nginx:1.30
services:
  - image: nginx:1.30
  - *web
`,
			},
			wantedResult: true,
		},
		{
			name: "Wildcard with merged entry and genuinely missing entry errors without searchpattern (issue #11053)",
			spec: Spec{
				File:  "test.yaml",
				Key:   "$.services[*].image",
				Value: "nginx:1.30",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
			},
			inputSourceValue: "nginx:1.29",
			mockedContents: map[string]string{
				"test.yaml": `x-common: &common
  image: nginx:1.29
services:
  - <<: *common
  - ports:
      - "80:80"
`,
			},
			wantedError: true,
		},
		{
			name: "Wildcard with duplicate merged entries succeeds without false missing error (issue #11053)",
			spec: Spec{
				File:  "test.yaml",
				Key:   "$.services[*].image",
				Value: "nginx:1.30",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
			},
			inputSourceValue: "nginx:1.29",
			mockedContents: map[string]string{
				"test.yaml": `x-common: &common
  image: nginx:1.29
services:
  - <<: *common
  - <<: *common
`,
			},
			wantedContents: map[string]string{
				"test.yaml": `x-common: &common
  image: nginx:1.30
services:
  - <<: *common
  - <<: *common
`,
			},
			wantedResult: true,
		},
		{
			name: "Backtracking across merge targets finds shared anchor on later step (issue #11053)",
			spec: Spec{
				File:  "test.yaml",
				Key:   "$.services.web.a.b",
				Value: "nginx:1.30",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
			},
			inputSourceValue: "nginx:1.29",
			mockedContents: map[string]string{
				"test.yaml": `x-inner: &inner
  b: nginx:1.29
x-outer: &outer
  a: *inner
services:
  web:
    <<: [*inner, *outer]
`,
			},
			wantedContents: map[string]string{
				"test.yaml": `x-inner: &inner
  b: nginx:1.30
x-outer: &outer
  a: *inner
services:
  web:
    <<: [*inner, *outer]
`,
			},
			wantedResult: true,
		},
		{
			name: "Recursive descent finds nested matching keys through alias (issue #11053)",
			spec: Spec{
				File:  "test.yaml",
				Key:   "$.app..child.target",
				Value: "nginx:1.30",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
			},
			inputSourceValue: "nginx:1.29",
			mockedContents: map[string]string{
				"test.yaml": `x-root: &root
  child:
    child:
      target: nginx:1.29
app: *root
`,
			},
			wantedContents: map[string]string{
				"test.yaml": `x-root: &root
  child:
    child:
      target: nginx:1.30
app: *root
`,
			},
			wantedResult: true,
		},
		{
			name: "Shallow merge precedence does not fall through shadowed key to later merge target (issue #11053)",
			spec: Spec{
				File:  "test.yaml",
				Key:   "$.services.web.env.image",
				Value: "nginx:1.30",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
			},
			inputSourceValue: "nginx:1.29",
			mockedContents: map[string]string{
				"test.yaml": `x-first: &first
  env:
    foo: bar
x-second: &second
  env:
    image: nginx:1.29
services:
  web:
    <<: [*first, *second]
`,
			},
			wantedError: true,
		},
		{
			name: "Merge falls through to later merge target when earlier target lacks the key (issue #11053)",
			spec: Spec{
				File:  "test.yaml",
				Key:   "$.services.web.env.image",
				Value: "nginx:1.30",
			},
			files: map[string]file{
				"test.yaml": {
					filePath:         "test.yaml",
					originalFilePath: "test.yaml",
				},
			},
			inputSourceValue: "nginx:1.29",
			mockedContents: map[string]string{
				"test.yaml": `x-first: &first
  other: bar
x-second: &second
  env:
    image: nginx:1.29
services:
  web:
    <<: [*first, *second]
`,
			},
			wantedContents: map[string]string{
				"test.yaml": `x-first: &first
  other: bar
x-second: &second
  env:
    image: nginx:1.30
services:
  web:
    <<: [*first, *second]
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
			gotErr := y.Target(context.Background(), tt.inputSourceValue, nil, pathresolver.Resolver{}, tt.dryRun, &gotResult)
			if tt.wantedError {
				assert.Error(t, gotErr)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.wantedResult, gotResult.Changed)
			if tt.wantedInformation != "" {
				assert.Equal(t, tt.wantedInformation, gotResult.Information)
			}

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

func Test_MatchPathEdgeCases(t *testing.T) {
	t.Run("Recursive descent matches nested child.child.target", func(t *testing.T) {
		docFile, err := parser.ParseBytes([]byte(`child:
  child:
    target: old
`), parser.ParseComments)
		require.NoError(t, err)
		require.Len(t, docFile.Docs, 1)

		idx := indexDocument(docFile.Docs[0])
		found, missing, err := idx.matchPath(docFile.Docs[0].Body, "$..child.target")
		require.NoError(t, err)
		assert.Equal(t, 0, missing)
		require.Len(t, found, 1)
		assert.Equal(t, "old", found[0].String())
	})

	t.Run("Sequence alias DAG with repeated wildcards does not blow up exponentially", func(t *testing.T) {
		const levels = 25
		var sb strings.Builder
		sb.WriteString("a0: &a0\n  - image: nginx:1.29\n")
		for i := 1; i <= levels; i++ {
			fmt.Fprintf(&sb, "a%d: &a%d [*a%d, *a%d]\n", i, i, i-1, i-1)
		}
		fmt.Fprintf(&sb, "root: *a%d\n", levels)

		key := "$.root" + strings.Repeat("[*]", levels+1) + ".image"

		t.Chdir(t.TempDir())

		mockedText := text.MockTextRetriever{
			Contents: map[string]string{
				"test.yaml": sb.String(),
			},
		}

		y, err := New(Spec{
			File:  "test.yaml",
			Key:   key,
			Value: "nginx:1.30",
		})
		require.NoError(t, err)
		y.contentRetriever = &mockedText
		y.files = map[string]file{
			"test.yaml": {
				filePath:         "test.yaml",
				originalFilePath: "test.yaml",
			},
		}

		gotResult := result.Target{}
		err = y.Target(context.Background(), "nginx:1.29", nil, pathresolver.Resolver{}, false, &gotResult)
		require.NoError(t, err)
		assert.True(t, gotResult.Changed)
		assert.Contains(t, mockedText.Contents["test.yaml"], "image: nginx:1.30")
		assert.Contains(t, mockedText.Contents["test.yaml"], "root: *a25")
	})

	t.Run("Sequence alias DAG with missing leaves reports missing without overflowing", func(t *testing.T) {
		const levels = 35
		var sb strings.Builder
		sb.WriteString("a0: &a0\n  - ports:\n      - 80\n")
		for i := 1; i <= levels; i++ {
			fmt.Fprintf(&sb, "a%d: &a%d [*a%d, *a%d]\n", i, i, i-1, i-1)
		}
		fmt.Fprintf(&sb, "root: *a%d\n", levels)

		docFile, err := parser.ParseBytes([]byte(sb.String()), parser.ParseComments)
		require.NoError(t, err)
		require.Len(t, docFile.Docs, 1)

		idx := indexDocument(docFile.Docs[0])
		key := "$.root" + strings.Repeat("[*]", levels+1) + ".image"
		found, missing, err := idx.matchPath(docFile.Docs[0].Body, key)
		require.NoError(t, err)
		assert.Empty(t, found)
		assert.Greater(t, missing, 0)
	})

	t.Run("Excessive traversal depth returns a bounded error", func(t *testing.T) {
		docFile, err := parser.ParseBytes([]byte("a: &a\n  - *a\nroot: *a\n"), parser.ParseComments)
		require.NoError(t, err)
		require.Len(t, docFile.Docs, 1)

		idx := indexDocument(docFile.Docs[0])
		key := "$.root" + strings.Repeat("[*]", maxMatchDepth+5)
		_, _, err = idx.matchPath(docFile.Docs[0].Body, key)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "maximum depth")
	})

	t.Run("SearchPattern updates both inline and merged wildcard entries", func(t *testing.T) {
		content := `x-common: &common
  image: nginx:1.29
services:
  - image: nginx:1.28
  - <<: *common
`
		wanted := `x-common: &common
  image: nginx:1.30
services:
  - image: nginx:1.30
  - <<: *common
`
		t.Chdir(t.TempDir())
		require.NoError(t, os.WriteFile("test.yaml", []byte(content), 0600))

		mockedText := text.MockTextRetriever{
			Contents: map[string]string{"test.yaml": content},
		}
		y, err := New(Spec{
			File:          "test.yaml",
			Key:           "$.services[*].image",
			Value:         "nginx:1.30",
			SearchPattern: true,
		})
		require.NoError(t, err)
		y.contentRetriever = &mockedText

		gotResult := result.Target{}
		require.NoError(t, y.Target(context.Background(), "nginx:1.29", nil, pathresolver.Resolver{}, false, &gotResult))
		assert.True(t, gotResult.Changed)
		assert.Equal(t, result.ATTENTION, gotResult.Result)
		assert.Equal(t, wanted, mockedText.Contents["test.yaml"])
	})

	t.Run("SearchPattern updates inline and merged entries while skipping genuinely missing wildcard entry", func(t *testing.T) {
		content := `x-common: &common
  image: nginx:1.29
services:
  - image: nginx:1.28
  - <<: *common
  - ports:
      - "80:80"
`
		wanted := `x-common: &common
  image: nginx:1.30
services:
  - image: nginx:1.30
  - <<: *common
  - ports:
      - "80:80"
`
		t.Chdir(t.TempDir())
		require.NoError(t, os.WriteFile("test.yaml", []byte(content), 0600))

		mockedText := text.MockTextRetriever{
			Contents: map[string]string{"test.yaml": content},
		}
		y, err := New(Spec{
			File:          "test.yaml",
			Key:           "$.services[*].image",
			Value:         "nginx:1.30",
			SearchPattern: true,
		})
		require.NoError(t, err)
		y.contentRetriever = &mockedText

		gotResult := result.Target{}
		require.NoError(t, y.Target(context.Background(), "nginx:1.29", nil, pathresolver.Resolver{}, false, &gotResult))
		assert.True(t, gotResult.Changed)
		assert.Equal(t, result.ATTENTION, gotResult.Result)
		assert.Equal(t, wanted, mockedText.Contents["test.yaml"])
	})

	t.Run("SearchPattern skips shadowed merge key without mutating later merge target", func(t *testing.T) {
		content := `x-first: &first
  env:
    foo: bar
x-second: &second
  env:
    image: nginx:1.29
services:
  web:
    <<: [*first, *second]
`
		t.Chdir(t.TempDir())
		require.NoError(t, os.WriteFile("test.yaml", []byte(content), 0600))

		mockedText := text.MockTextRetriever{
			Contents: map[string]string{"test.yaml": content},
		}
		y, err := New(Spec{
			File:          "test.yaml",
			Key:           "$.services.web.env.image",
			Value:         "nginx:1.30",
			SearchPattern: true,
		})
		require.NoError(t, err)
		y.contentRetriever = &mockedText

		gotResult := result.Target{}
		require.NoError(t, y.Target(context.Background(), "nginx:1.29", nil, pathresolver.Resolver{}, false, &gotResult))
		assert.False(t, gotResult.Changed)
		assert.Equal(t, result.SKIPPED, gotResult.Result)
		assert.Equal(t, content, mockedText.Contents["test.yaml"])
	})

	t.Run("Recursive descent updates both inline and merged matches", func(t *testing.T) {
		content := `x-common: &common
  image: nginx:1.29
services:
  web:
    image: nginx:1.28
  api:
    <<: *common
`
		wanted := `x-common: &common
  image: nginx:1.30
services:
  web:
    image: nginx:1.30
  api:
    <<: *common
`
		t.Chdir(t.TempDir())
		require.NoError(t, os.WriteFile("test.yaml", []byte(content), 0600))

		mockedText := text.MockTextRetriever{
			Contents: map[string]string{"test.yaml": content},
		}
		y, err := New(Spec{
			File:  "test.yaml",
			Key:   "$.services..image",
			Value: "nginx:1.30",
		})
		require.NoError(t, err)
		y.contentRetriever = &mockedText

		gotResult := result.Target{}
		require.NoError(t, y.Target(context.Background(), "nginx:1.29", nil, pathresolver.Resolver{}, false, &gotResult))
		assert.True(t, gotResult.Changed)
		assert.Equal(t, result.ATTENTION, gotResult.Result)
		assert.Equal(t, wanted, mockedText.Contents["test.yaml"])
	})

	t.Run("Recursive descent updates both inline and mapping-aliased matches", func(t *testing.T) {
		content := `x-api: &api
  image: nginx:1.29
services:
  web:
    image: nginx:1.28
  api: *api
`
		wanted := `x-api: &api
  image: nginx:1.30
services:
  web:
    image: nginx:1.30
  api: *api
`
		t.Chdir(t.TempDir())
		require.NoError(t, os.WriteFile("test.yaml", []byte(content), 0600))

		mockedText := text.MockTextRetriever{
			Contents: map[string]string{"test.yaml": content},
		}
		y, err := New(Spec{
			File:  "test.yaml",
			Key:   "$.services..image",
			Value: "nginx:1.30",
		})
		require.NoError(t, err)
		y.contentRetriever = &mockedText

		gotResult := result.Target{}
		require.NoError(t, y.Target(context.Background(), "nginx:1.29", nil, pathresolver.Resolver{}, false, &gotResult))
		assert.True(t, gotResult.Changed)
		assert.Equal(t, result.ATTENTION, gotResult.Result)
		assert.Equal(t, wanted, mockedText.Contents["test.yaml"])
	})

	t.Run("Nested wildcards traversing an empty aliased sequence do not report missing key", func(t *testing.T) {
		content := `x-empty: &empty []
a:
  - b:
      - c: *empty
  - b:
      - c:
          - d: nginx:1.29
`
		wanted := `x-empty: &empty []
a:
  - b:
      - c: *empty
  - b:
      - c:
          - d: nginx:1.30
`
		t.Chdir(t.TempDir())
		require.NoError(t, os.WriteFile("test.yaml", []byte(content), 0600))

		mockedText := text.MockTextRetriever{
			Contents: map[string]string{"test.yaml": content},
		}
		y, err := New(Spec{
			File:  "test.yaml",
			Key:   "$.a[*].b[*].c[*].d",
			Value: "nginx:1.30",
		})
		require.NoError(t, err)
		y.contentRetriever = &mockedText

		gotResult := result.Target{}
		require.NoError(t, y.Target(context.Background(), "nginx:1.29", nil, pathresolver.Resolver{}, false, &gotResult))
		assert.True(t, gotResult.Changed)
		assert.Equal(t, result.ATTENTION, gotResult.Result)
		assert.Equal(t, wanted, mockedText.Contents["test.yaml"])
	})

	t.Run("Nested wildcard with empty trailing aliased sequence does not report missing key", func(t *testing.T) {
		content := `x-empty: &empty []
groups:
  - items: *empty
  - items:
      - nginx:1.29
`
		wanted := `x-empty: &empty []
groups:
  - items: *empty
  - items:
      - nginx:1.30
`
		t.Chdir(t.TempDir())
		require.NoError(t, os.WriteFile("test.yaml", []byte(content), 0600))

		mockedText := text.MockTextRetriever{
			Contents: map[string]string{"test.yaml": content},
		}
		y, err := New(Spec{
			File:  "test.yaml",
			Key:   "$.groups[*].items[*]",
			Value: "nginx:1.30",
		})
		require.NoError(t, err)
		y.contentRetriever = &mockedText

		gotResult := result.Target{}
		require.NoError(t, y.Target(context.Background(), "nginx:1.29", nil, pathresolver.Resolver{}, false, &gotResult))
		assert.True(t, gotResult.Changed)
		assert.Equal(t, result.ATTENTION, gotResult.Result)
		assert.Equal(t, wanted, mockedText.Contents["test.yaml"])
	})

	t.Run("Nested wildcards still report genuinely missing key alongside empty aliased sequence", func(t *testing.T) {
		content := `x-empty: &empty []
a:
  - b:
      - c: *empty
      - other: value
  - b:
      - c:
          - d: nginx:1.29
`
		t.Chdir(t.TempDir())
		require.NoError(t, os.WriteFile("test.yaml", []byte(content), 0600))

		mockedText := text.MockTextRetriever{
			Contents: map[string]string{"test.yaml": content},
		}
		y, err := New(Spec{
			File:  "test.yaml",
			Key:   "$.a[*].b[*].c[*].d",
			Value: "nginx:1.30",
		})
		require.NoError(t, err)
		y.contentRetriever = &mockedText

		gotResult := result.Target{}
		err = y.Target(context.Background(), "nginx:1.29", nil, pathresolver.Resolver{}, false, &gotResult)
		require.Error(t, err)
	})
}
