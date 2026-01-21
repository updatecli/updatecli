package file

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
)

func TestFile_TargetMultiples(t *testing.T) {
	tests := []struct {
		name             string
		spec             Spec
		files            map[string]fileMetadata
		inputSourceValue string
		mockedContents   map[string]string
		mockedError      error
		wantedContents   map[string]string
		wantedResult     bool
		wantedErr        bool
		dryRun           bool
	}{
		{
			name:             "(File) Replace content with matchPattern and ReplacePattern",
			inputSourceValue: "3.9.0",
			spec: Spec{
				File:           "foo.txt",
				MatchPattern:   "maven_(.*)=.*",
				ReplacePattern: "maven_$1= 3.9.0",
			},
			files: map[string]fileMetadata{
				"foo.txt": {
					originalPath: "foo.txt",
					path:         "foo.txt",
				},
			},
			mockedContents: map[string]string{
				"foo.txt": `maven_version = "3.8.2"
				git_version = "2.33.1"
				jdk11_version = "11.0.12+7"
				jdk17_version = "17+35"
				jdk8_version = "8u302-b08"
				maven_major_release = "3"
				git_lfs_version = "3.0.1"
				compose_version = "1.29.2"`,
			},
			wantedContents: map[string]string{
				"foo.txt": `maven_version = 3.9.0
				git_version = "2.33.1"
				jdk11_version = "11.0.12+7"
				jdk17_version = "17+35"
				jdk8_version = "8u302-b08"
				maven_major_release = 3.9.0
				git_lfs_version = "3.0.1"
				compose_version = "1.29.2"`,
			},
			wantedResult: true,
		},
		{
			name:             "Replace content with matchPattern and ReplacePattern",
			inputSourceValue: "3.9.0",
			spec: Spec{
				Files: []string{
					"foo.txt",
					"bar.txt",
				},
				MatchPattern:   "maven_(.*)=.*",
				ReplacePattern: "maven_$1= 3.9.0",
			},
			files: map[string]fileMetadata{
				"foo.txt": {
					originalPath: "foo.txt",
					path:         "foo.txt",
				},
				"bar.txt": {
					originalPath: "bar.txt",
					path:         "bar.txt",
				},
			},
			mockedContents: map[string]string{
				"foo.txt": `maven_version = "3.8.2"
				git_version = "2.33.1"
				jdk11_version = "11.0.12+7"
				jdk17_version = "17+35"
				jdk8_version = "8u302-b08"
				maven_major_release = "3"
				git_lfs_version = "3.0.1"
				compose_version = "1.29.2"`,

				"bar.txt": `another_param=9.8.7
				maven_major_release = "2"
				maven_version = "3.1.2"
				some_stuff = "11.9.1"`,
			},
			wantedContents: map[string]string{
				"foo.txt": `maven_version = 3.9.0
				git_version = "2.33.1"
				jdk11_version = "11.0.12+7"
				jdk17_version = "17+35"
				jdk8_version = "8u302-b08"
				maven_major_release = 3.9.0
				git_lfs_version = "3.0.1"
				compose_version = "1.29.2"`,

				"bar.txt": `another_param=9.8.7
				maven_major_release = 3.9.0
				maven_version = 3.9.0
				some_stuff = "11.9.1"`,
			},
			wantedResult: true,
		},
		{
			name: "(File) Passing case with both input source and specified content but no line (specified content should be used)",
			spec: Spec{
				File:    "foo.txt",
				Content: "Be happy",
			},
			files: map[string]fileMetadata{
				"foo.txt": {
					originalPath: "foo.txt",
					path:         "foo.txt",
				},
			},
			inputSourceValue: "current_version=1.2.3",
			mockedContents: map[string]string{
				"foo.txt": "Hello World",
			},
			wantedContents: map[string]string{
				"foo.txt": "Be happy",
			},
			wantedResult: true,
		},
		{
			name: "Passing case with both input source and specified content but no line (specified content should be used)",
			spec: Spec{
				Files: []string{
					"foo.txt",
					"bar.txt",
				},
				Content: "Be happy",
			},
			files: map[string]fileMetadata{
				"foo.txt": {
					originalPath: "foo.txt",
					path:         "foo.txt",
				},
				"bar.txt": {
					originalPath: "bar.txt",
					path:         "bar.txt",
				},
			},
			inputSourceValue: "current_version=1.2.3",
			mockedContents: map[string]string{
				"foo.txt": "Hello World",
				"bar.txt": "Another content",
			},
			wantedContents: map[string]string{
				"foo.txt": "Be happy",
				"bar.txt": "Be happy",
			},
			wantedResult: true,
		},
		{
			name: "(File) Passing case with an updated line from provided content",
			spec: Spec{
				File:    "foo.txt",
				Content: "Hello World",
				Line:    2,
			},
			files: map[string]fileMetadata{
				"foo.txt": {
					originalPath: "foo.txt",
					path:         "foo.txt",
				},
			},
			mockedContents: map[string]string{
				"foo.txt": "Title\r\nGood Bye\r\nThe end",
			},
			inputSourceValue: "current_version=1.2.3",
			wantedContents: map[string]string{
				"foo.txt": "Title\r\nHello World\r\nThe end",
			},
			wantedResult: true,
		},
		{
			name: "Passing case with an updated line from provided content",
			spec: Spec{
				Files: []string{
					"foo.txt",
					"bar.txt",
				},
				Content: "Hello World",
				Line:    2,
			},
			files: map[string]fileMetadata{
				"foo.txt": {
					originalPath: "foo.txt",
					path:         "foo.txt",
				},
				"bar.txt": {
					originalPath: "bar.txt",
					path:         "bar.txt",
				},
			},
			mockedContents: map[string]string{
				"foo.txt": "Title\r\nGood Bye\r\nThe end",
				"bar.txt": "Be happy\nDon't worry",
			},
			inputSourceValue: "current_version=1.2.3",
			wantedContents: map[string]string{
				"foo.txt": "Title\r\nHello World\r\nThe end",
				"bar.txt": "Be happy\nHello World",
			},
			wantedResult: true,
		},
		{
			name: "(File) Validation failure with an https:// URL instead of a file",
			spec: Spec{
				File: "https://github.com/foo.txt",
			},
			files: map[string]fileMetadata{
				"https://github.com/foo.txt": {
					originalPath: "https://github.com/foo.txt",
					path:         "https://github.com/foo.txt",
				},
			},
			wantedResult: false,
			wantedErr:    true,
		},
		{
			name: "Validation failure with an https:// URL instead of a file",
			spec: Spec{
				Files: []string{
					"foo.txt",
					"https://github.com/bar.txt",
				},
			},
			files: map[string]fileMetadata{
				"foo.txt": {
					originalPath: "foo.txt",
					path:         "foo.txt",
				},
				"https://github.com/bar.txt": {
					originalPath: "https://github.com/bar.txt",
					path:         "https://github.com/bar.txt",
				},
			},
			wantedResult: false,
			wantedErr:    true,
		},
		{
			name: "Validation failure with both line and forcecreate specified",
			spec: Spec{
				Files: []string{
					"foo.txt",
				},
				ForceCreate: true,
				Line:        2,
			},
			files: map[string]fileMetadata{
				"foo.txt": {
					originalPath: "foo.txt",
					path:         "foo.txt",
				},
			},
			wantedResult: false,
			wantedErr:    true,
		},
		{
			name: "Validation failure with invalid regexp for MatchPattern",
			spec: Spec{
				Files: []string{
					"foo.txt",
				},
				MatchPattern: "(d+:1",
			},
			files: map[string]fileMetadata{
				"foo.txt": {
					originalPath: "foo.txt",
					path:         "foo.txt",
				},
			},
			wantedResult: false,
			wantedErr:    true,
		},
		{
			name: "Error with file not existing (with line)",
			spec: Spec{
				Files: []string{
					"not_existing.txt",
				},
				Line: 3,
			},
			files: map[string]fileMetadata{
				"not_existing.txt": {
					originalPath: "not_existing.txt",
					path:         "not_existing.txt",
				},
			},
			wantedResult: false,
			wantedErr:    true,
		},
		{
			name: "Error with file not existing (with content)",
			spec: Spec{
				Files: []string{
					"not_existing.txt",
				},
				Content: "Hello World",
			},
			files: map[string]fileMetadata{
				"not_existing.txt": {
					originalPath: "not_existing.txt",
					path:         "not_existing.txt",
				},
			},
			wantedResult: false,
			wantedErr:    true,
		},
		{
			name: "Error while reading the line in file",
			spec: Spec{
				Files: []string{
					"foo.txt",
				},
				Line: 3,
			},
			files: map[string]fileMetadata{
				"foo.txt": {
					originalPath: "foo.txt",
					path:         "foo.txt",
				},
			},
			mockedContents: map[string]string{
				"foo.txt": "Be happy",
			},
			mockedError:  fmt.Errorf("I/O error: file system too slow"),
			wantedResult: false,
			wantedErr:    true,
		},
		{
			name: "Error while reading a full file",
			spec: Spec{
				Files: []string{
					"foo.txt",
				},
				Content: "Hello World",
			},
			files: map[string]fileMetadata{
				"foo.txt": {
					originalPath: "foo.txt",
					path:         "foo.txt",
				},
			},
			mockedContents: map[string]string{
				"foo.txt": "Be happy",
			},
			mockedError: fmt.Errorf("I/O error: file system too slow"),
			wantedContents: map[string]string{
				"foo.txt": "current_version=1.2.3",
			},
			wantedResult: false,
			wantedErr:    true,
		},
		{
			name: "Line in files not updated",
			spec: Spec{
				Files: []string{
					"foo.txt",
					"bar.txt",
				},
				Line: 3,
			},
			files: map[string]fileMetadata{
				"foo.txt": {
					path:         "foo.txt",
					originalPath: "foo.txt",
				},
				"bar.txt": {
					path:         "bar.txt",
					originalPath: "bar.txt",
				},
			},
			inputSourceValue: "Be happy",
			mockedContents: map[string]string{
				"foo.txt": "Title\r\nGood Bye\r\nBe happy",
				"bar.txt": "Be happy\nDon't worry\nBe happy\nDon't worry",
			},
			wantedContents: map[string]string{
				"foo.txt": "Title\r\nGood Bye\r\nBe happy",
				"bar.txt": "Be happy\nDon't worry\nBe happy\nDon't worry",
			},
			wantedResult: false,
		},
		{
			name: "Files not updated (input source, no specified line)",
			spec: Spec{
				Files: []string{
					"foo.txt",
					"bar.txt",
				},
			},
			files: map[string]fileMetadata{
				"foo.txt": {
					originalPath: "foo.txt",
					path:         "foo.txt",
				},
				"bar.txt": {
					originalPath: "bar.txt",
					path:         "bar.txt",
				},
			},
			inputSourceValue: "current_version=1.2.3",
			mockedContents: map[string]string{
				"foo.txt": "current_version=1.2.3",
				"bar.txt": "current_version=1.2.3",
			},
			wantedContents: map[string]string{
				"foo.txt": "current_version=1.2.3",
				"bar.txt": "current_version=1.2.3",
			},
			wantedResult: false,
		},
		{
			name:             "Template with source value",
			inputSourceValue: "1.2.3",
			spec: Spec{
				File:     "version.txt",
				Template: "template1.tmpl",
			},
			files: map[string]fileMetadata{
				"version.txt": {
					originalPath: "version.txt",
					path:         "version.txt",
				},
			},
			mockedContents: map[string]string{
				"version.txt":    "Version: 1.0.0",
				"template1.tmpl": "Version: {{ .source }}",
			},
			wantedContents: map[string]string{
				"version.txt": "Version: 1.2.3",
			},
			wantedResult: true,
		},
		{
			name:             "Template with sprig functions",
			inputSourceValue: "hello-world",
			spec: Spec{
				File:     "config.txt",
				Template: "template2.tmpl",
			},
			files: map[string]fileMetadata{
				"config.txt": {
					originalPath: "config.txt",
					path:         "config.txt",
				},
			},
			mockedContents: map[string]string{
				"config.txt":     "Name: OLD",
				"template2.tmpl": "Name: {{ .source | upper }}",
			},
			wantedContents: map[string]string{
				"config.txt": "Name: HELLO-WORLD",
			},
			wantedResult: true,
		},
		{
			name:             "Template with multiple files",
			inputSourceValue: "2.0.0",
			spec: Spec{
				Files: []string{
					"foo.yaml",
					"bar.yaml",
				},
				Template: "template3.tmpl",
			},
			files: map[string]fileMetadata{
				"foo.yaml": {
					originalPath: "foo.yaml",
					path:         "foo.yaml",
				},
				"bar.yaml": {
					originalPath: "bar.yaml",
					path:         "bar.yaml",
				},
			},
			mockedContents: map[string]string{
				"foo.yaml":       "version: 1.0.0",
				"bar.yaml":       "version: 1.5.0",
				"template3.tmpl": "version: {{ .source }}",
			},
			wantedContents: map[string]string{
				"foo.yaml": "version: 2.0.0",
				"bar.yaml": "version: 2.0.0",
			},
			wantedResult: true,
		},
		{
			name:             "Template with complex structure",
			inputSourceValue: "v3.5.1",
			spec: Spec{
				File:     "deployment.yaml",
				Template: "template4.tmpl",
			},
			files: map[string]fileMetadata{
				"deployment.yaml": {
					originalPath: "deployment.yaml",
					path:         "deployment.yaml",
				},
			},
			mockedContents: map[string]string{
				"deployment.yaml": "old content",
				"template4.tmpl": `apiVersion: apps/v1
					kind: Deployment
					metadata:
					  name: myapp
					spec:
					  template:
						spec:
						  containers:
						  - name: app
							image: myapp:{{ .source }}`,
			},
			wantedContents: map[string]string{
				"deployment.yaml": `apiVersion: apps/v1
					kind: Deployment
					metadata:
					  name: myapp
					spec:
					  template:
						spec:
						  containers:
						  - name: app
							image: myapp:v3.5.1`,
			},
			wantedResult: true,
		},
		{
			name:             "Template with JSON array using fromJson",
			inputSourceValue: `["v1.0.0", "v2.0.0", "v3.0.0"]`,
			spec: Spec{
				File:     "versions.yaml",
				Template: "template5.tmpl",
			},
			files: map[string]fileMetadata{
				"versions.yaml": {
					originalPath: "versions.yaml",
					path:         "versions.yaml",
				},
			},
			mockedContents: map[string]string{
				"versions.yaml": "old content",
				"template5.tmpl": `versions:
					{{- $vers := .source | fromJson }}
					{{- range $vers }}
					  - {{ . }}
					{{- end }}`,
			},
			wantedContents: map[string]string{
				"versions.yaml": `versions:
					  - v1.0.0
					  - v2.0.0
					  - v3.0.0`,
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
			f := &File{
				spec:             tt.spec,
				contentRetriever: &mockedText,
				files:            tt.files,
			}

			gotResultTarget := result.Target{}
			gotErr := f.Target(tt.inputSourceValue, nil, tt.dryRun, &gotResultTarget)

			if tt.wantedErr {
				assert.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)

			assert.Equal(t, tt.wantedResult, gotResultTarget.Changed)
			for filePath := range tt.files {
				assert.Equal(t, tt.wantedContents[filePath], mockedText.Contents[filePath])
			}
		})
	}
}

func TestFile_TargetFromSCM(t *testing.T) {
	tests := []struct {
		name             string
		spec             Spec
		files            map[string]fileMetadata
		scm              scm.ScmHandler
		inputSourceValue string
		mockedContents   map[string]string
		mockedError      error
		wantedFiles      []string
		wantedContents   map[string]string
		wantedResult     bool
		wantedErr        bool
		dryRun           bool
	}{
		{
			name: "Passing case with 'Line' specified",
			spec: Spec{
				Files: []string{
					"foo.txt",
					"bar.txt",
				},
				Line: 3,
			},
			files: map[string]fileMetadata{
				"/tmp/foo.txt": {
					originalPath: "/tmp/foo.txt",
					path:         "/tmp/foo.txt",
				},
				"/tmp/bar.txt": {
					originalPath: "/tmp/bar.txt",
					path:         "/tmp/bar.txt",
				},
			},
			scm: &scm.MockScm{
				WorkingDir: "/tmp",
			},
			inputSourceValue: "current_version=1.2.3",
			mockedContents: map[string]string{
				"/tmp/foo.txt": "Title\r\nGood Bye\r\nThe End",
				"/tmp/bar.txt": "Be happy\nDon't worry\nBe happy\nDon't worry",
			},
			// returned files are sorted
			wantedFiles: []string{
				"/tmp/bar.txt",
				"/tmp/foo.txt",
			},
			wantedContents: map[string]string{
				"/tmp/foo.txt": "Title\r\nGood Bye\r\ncurrent_version=1.2.3",
				"/tmp/bar.txt": "Be happy\nDon't worry\ncurrent_version=1.2.3\nDon't worry",
			},
			wantedResult: true,
		},
		{
			name: "Passing case with 'ForceCreate' specified",
			spec: Spec{
				Files: []string{
					"foo.txt",
					"bar.txt",
				},
				ForceCreate: true,
			},
			files: map[string]fileMetadata{
				"/tmp/foo.txt": {
					originalPath: "/tmp/foo.txt",
					path:         "/tmp/foo.txt",
				},
				"/tmp/bar.txt": {
					originalPath: "/tmp/bar.txt",
					path:         "/tmp/bar.txt",
				},
			},
			scm: &scm.MockScm{
				WorkingDir: "/tmp",
			},
			inputSourceValue: "current_version=1.2.3",
			// Note there isn't any "bar.txt" defined here
			mockedContents: map[string]string{
				"/tmp/foo.txt": "Title\r\nGood Bye\r\ncurrent_version=1.2.3",
			},
			// returned files are sorted
			wantedFiles: []string{
				"/tmp/bar.txt",
				"/tmp/foo.txt",
			},
			wantedContents: map[string]string{
				"/tmp/foo.txt": "current_version=1.2.3",
				"/tmp/bar.txt": "current_version=1.2.3",
			},
			wantedResult: true,
		},
		{
			name: "No line matched with matchPattern and ReplacePattern defined",
			spec: Spec{
				Files: []string{
					"foo.txt",
					"bar.txt",
				},
				MatchPattern:   "notmatching=*",
				ReplacePattern: "maven_version= 3.9.0",
			},
			files: map[string]fileMetadata{
				"/tmp/bar.txt": {
					originalPath: "/tmp/bar.txt",
					path:         "/tmp/bar.txt",
				},
				"/tmp/foo.txt": {
					originalPath: "/tmp/foo.txt",
					path:         "/tmp/foo.txt",
				},
			},
			scm: &scm.MockScm{
				WorkingDir: "/tmp",
			},
			inputSourceValue: "3.9.0",
			// Note there is a match in "bar.txt" here
			mockedContents: map[string]string{
				"/tmp/foo.txt": `maven_version = "3.8.2"
				git_version = "2.33.1"
				jdk11_version = "11.0.12+7"
				jdk17_version = "17+35"
				jdk8_version = "8u302-b08"
				maven_major_release = "3"
				git_lfs_version = "3.0.1"
				compose_version = "1.29.2"`,

				"/tmp/bar.txt": `maven_version = "3.8.2"
				notmatching= "2.33.1"
				jdk11_version = "11.0.12+7"
				jdk17_version = "17+35"
				jdk8_version = "8u302-b08"
				maven_major_release = "3"
				git_lfs_version = "3.0.1"
				compose_version = "1.29.2"`,
			},
			wantedResult: false,
			wantedErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockedText := text.MockTextRetriever{
				Contents: tt.mockedContents,
				Err:      tt.mockedError,
			}
			f := &File{
				spec:             tt.spec,
				contentRetriever: &mockedText,
				files:            tt.files,
			}

			gotResultTarget := result.Target{}

			gotErr := f.Target(tt.inputSourceValue, tt.scm, tt.dryRun, &gotResultTarget)

			if tt.wantedErr {
				assert.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)

			assert.Equal(t, tt.wantedResult, gotResultTarget.Changed)
			assert.Equal(t, tt.wantedFiles, gotResultTarget.Files)

		})
	}
}

func TestFile_Target_CaptureGroupExtraction(t *testing.T) {
	tests := []struct {
		name                   string
		spec                   Spec
		files                  map[string]fileMetadata
		inputSourceValue       string
		mockedContents         map[string]string
		expectedInformation    string
		expectedNewInformation string
		wantedResult           bool
		description            string
	}{
		{
			name:             "Extract version from capture group in JSON format",
			inputSourceValue: "1.25.1",
			spec: Spec{
				File:           "go_version.json",
				MatchPattern:   `"GO_VERSION"\s*:\s*"(1\.24\.\d+)"`,
				ReplacePattern: `"GO_VERSION": "1.25.1"`,
			},
			files: map[string]fileMetadata{
				"go_version.json": {
					originalPath: "go_version.json",
					path:         "go_version.json",
				},
			},
			mockedContents: map[string]string{
				"go_version.json": `{
  "GO_VERSION": "1.24.5",
  "OTHER_VERSION": "2.1.0"
}`,
			},
			expectedInformation:    "1.24.5",
			expectedNewInformation: `"GO_VERSION": "1.25.1"`, // ReplacePattern used when specified
			wantedResult:           true,
			description:            "Should extract '1.24.5' from capture group instead of 'unknown'",
		},
		{
			name:             "Extract version from capture group in properties format",
			inputSourceValue: "3.9.0",
			spec: Spec{
				File:           "config.properties",
				MatchPattern:   `maven_version\s*=\s*"(3\.8\.\d+)"`,
				ReplacePattern: `maven_version = "3.9.0"`,
			},
			files: map[string]fileMetadata{
				"config.properties": {
					originalPath: "config.properties",
					path:         "config.properties",
				},
			},
			mockedContents: map[string]string{
				"config.properties": `maven_version = "3.8.2"
git_version = "2.33.1"
jdk_version = "11.0.12"`,
			},
			expectedInformation:    "3.8.2",
			expectedNewInformation: `maven_version = "3.9.0"`, // ReplacePattern used when specified
			wantedResult:           true,
			description:            "Should extract '3.8.2' from capture group",
		},
		{
			name:             "No capture group - should remain unknown",
			inputSourceValue: "1.25.1",
			spec: Spec{
				File:           "version.txt",
				MatchPattern:   `GO_VERSION.*1\.24\.\d+`, // No parentheses = no capture group
				ReplacePattern: `GO_VERSION: 1.25.1`,
			},
			files: map[string]fileMetadata{
				"version.txt": {
					originalPath: "version.txt",
					path:         "version.txt",
				},
			},
			mockedContents: map[string]string{
				"version.txt": `GO_VERSION: 1.24.5`,
			},
			expectedInformation:    "unknown",
			expectedNewInformation: `GO_VERSION: 1.25.1`, // ReplacePattern used when specified
			wantedResult:           true,
			description:            "Should remain 'unknown' when no capture group present",
		},
		{
			name:             "Multiple capture groups - should use first one",
			inputSourceValue: "2.0.0",
			spec: Spec{
				File:           "multi_version.txt",
				MatchPattern:   `version:\s*"(1\.\d+)\.(\d+)"`,
				ReplacePattern: `version: "2.0.0"`,
			},
			files: map[string]fileMetadata{
				"multi_version.txt": {
					originalPath: "multi_version.txt",
					path:         "multi_version.txt",
				},
			},
			mockedContents: map[string]string{
				"multi_version.txt": `version: "1.24.5"`,
			},
			expectedInformation:    "1.24",             // First capture group
			expectedNewInformation: `version: "2.0.0"`, // ReplacePattern used when specified
			wantedResult:           true,
			description:            "Should extract first capture group when multiple groups present",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockedText := text.MockTextRetriever{
				Contents: tt.mockedContents,
			}
			f := &File{
				spec:             tt.spec,
				contentRetriever: &mockedText,
				files:            tt.files,
			}

			gotResultTarget := result.Target{}
			err := f.Target(tt.inputSourceValue, nil, true, &gotResultTarget) // dry run
			require.NoError(t, err, tt.description)

			assert.Equal(t, tt.expectedInformation, gotResultTarget.Information,
				"Information field should be %q - %s", tt.expectedInformation, tt.description)
			assert.Equal(t, tt.expectedNewInformation, gotResultTarget.NewInformation,
				"NewInformation field should be %q", tt.expectedNewInformation)
			assert.Equal(t, tt.wantedResult, gotResultTarget.Changed,
				"Changed field should be %t", tt.wantedResult)
		})
	}
}
