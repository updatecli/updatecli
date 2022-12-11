package file

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/text"
)

func TestFile_TargetMultiples(t *testing.T) {
	tests := []struct {
		name             string
		spec             Spec
		files            map[string]string
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
			files: map[string]string{
				"foo.txt": "",
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
			files: map[string]string{
				"foo.txt": "",
				"bar.txt": "",
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
			files: map[string]string{
				"foo.txt": "",
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
			files: map[string]string{
				"foo.txt": "",
				"bar.txt": "",
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
			files: map[string]string{
				"foo.txt": "",
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
			files: map[string]string{
				"foo.txt": "",
				"bar.txt": "",
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
			files: map[string]string{
				"https://github.com/foo.txt": "",
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
			files: map[string]string{
				"foo.txt":                    "",
				"https://github.com/bar.txt": "",
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
			files: map[string]string{
				"foo.txt": "",
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
			files: map[string]string{
				"foo.txt": "",
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
			files: map[string]string{
				"not_existing.txt": "",
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
			files: map[string]string{
				"not_existing.txt": "",
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
			files: map[string]string{
				"foo.txt": "",
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
			files: map[string]string{
				"foo.txt": "",
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
			files: map[string]string{
				"foo.txt": "",
				"bar.txt": "",
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
			files: map[string]string{
				"foo.txt": "",
				"bar.txt": "",
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

			gotResult, gotErr := f.Target(tt.inputSourceValue, tt.dryRun)

			if tt.wantedErr {
				assert.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)

			assert.Equal(t, tt.wantedResult, gotResult)
			for filePath := range tt.files {
				assert.Equal(t, tt.wantedContents[filePath], mockedText.Contents[filePath])
			}
		})
	}
}
