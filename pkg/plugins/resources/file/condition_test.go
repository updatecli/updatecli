package file

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/text"
)

// TODO: find a way to test when there are multiple files but only some are successful
func TestFile_Condition(t *testing.T) {
	tests := []struct {
		name             string
		spec             Spec
		files            map[string]string
		inputSourceValue string
		mockedContents   map[string]string
		mockedError      error
		wantedResult     bool
		wantedErr        bool
	}{
		{
			name: "Passing case with 'Line' specified",
			spec: Spec{
				Files: []string{
					"foo.txt",
				},
				Line: 3,
			},
			files: map[string]string{
				"foo.txt": "",
			},
			inputSourceValue: "current_version=1.2.3",
			mockedContents: map[string]string{
				"foo.txt": "Hello World\r\nAnother line\r\ncurrent_version=1.2.3",
			},
			wantedResult: true,
		},
		{
			name: "Passing case with 'Content' specified and no source specified",
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
				"foo.txt": "Hello World",
			},
			wantedResult: true,
		},
		{
			name: "Validation failure with more than one element in 'Files'",
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
			wantedErr: true,
		},
		{
			name: "Validation failure with both source and 'Content' specified",
			spec: Spec{
				Files: []string{
					"foo.txt",
				},
				Content: "Hello World",
			},
			files: map[string]string{
				"foo.txt": "",
			},
			inputSourceValue: "1.2.3",
			wantedErr:        true,
		},
		{
			name: "Validation failure with 'ReplacePattern' specified",
			spec: Spec{
				Files: []string{
					"foo.txt",
				},
				MatchPattern:   "maven_(.*)",
				ReplacePattern: "gradle_$1",
			},
			files: map[string]string{
				"foo.txt": "",
			},
			inputSourceValue: "1.2.3",
			wantedErr:        true,
		},
		{
			name: "Failing case with empty 'Line' specified and no source nor 'Content' specified",
			spec: Spec{
				Files: []string{
					"foo.txt",
				},
				Line: 11,
			},
			files: map[string]string{
				"foo.txt": "",
			},
			mockedContents: map[string]string{
				"foo.txt": "",
			},
			wantedErr: true,
		},
		{
			name: "Case with not empty 'Line' specified and no source nor 'Content' specified",
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
				"foo.txt": "A first line\r\nAnother line\r\nSomething On The Specified Line",
			},
			wantedResult: true,
		},
		{
			name: "Failing case with non existing 'Line' specified",
			spec: Spec{
				Files: []string{
					"foo.txt",
				},
				Line: 5,
			},
			files: map[string]string{
				"foo.txt": "",
			},
			wantedErr: true,
		},
		{
			name: "Failing case with non existing 'Files'",
			spec: Spec{
				Files: []string{
					"foo.txt",
				},
			},
			files: map[string]string{
				"foo.txt": "",
			},
			wantedErr: true,
		},
		{
			name: "Failing case with non existing URL as 'Files'",
			spec: Spec{
				Files: []string{
					"https://do.not.exists/foo",
				},
			},
			files: map[string]string{
				"https://do.not.exists/foo": "",
			},
			mockedError: fmt.Errorf("URL %q not found or in error", "https://do.not.exists/foo"),
			wantedErr:   true,
		},
		{
			name: "'No result' case with not empty 'Line' not matching the 'Content' of the 'Files' at the 'Line' specified",
			spec: Spec{
				Files: []string{
					"foo.txt",
				},
				Line:    2,
				Content: "Not In All Files",
			},
			files: map[string]string{
				"foo.txt": "",
			},
			inputSourceValue: "",
			mockedContents: map[string]string{
				"foo.txt": "Title\r\nGood Bye\r\nThe End",
			},
			wantedResult: false,
			wantedErr:    false,
		},
		{
			name: "Failing case with more than one 'Files'",
			spec: Spec{
				Files: []string{
					"foo.txt",
					"bar.txt",
				},
				Line:    2,
				Content: "Not In All Files",
			},
			files: map[string]string{
				"foo.txt": "",
				"bar.txt": "",
			},
			inputSourceValue: "",
			mockedContents: map[string]string{
				"foo.txt": "Title\r\nGood Bye\r\nThe End",
				"bar.txt": "Title\r\nNot In All Files\r\nThe End",
			},
			wantedErr: true,
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

			gotResult, gotErr := f.Condition(tt.inputSourceValue)
			if tt.wantedErr {
				assert.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, tt.wantedResult, gotResult)
		})
	}
}

func TestFile_ConditionFromSCM(t *testing.T) {
	tests := []struct {
		name             string
		spec             Spec
		files            map[string]string
		scm              scm.ScmHandler
		inputSourceValue string
		mockedContents   map[string]string
		mockedError      error
		wantedContents   map[string]string
		wantedResult     bool
		wantedErr        bool
	}{
		{
			name: "Passing case with both 'Line' and 'Content' specified, 'Files' with a relative path, and no source",
			spec: Spec{
				Files: []string{
					"foo.txt",
				},
				Content: "current_version=1.2.3",
				Line:    3,
			},
			files: map[string]string{
				"/tmp/foo.txt": "",
			},
			scm: &scm.MockScm{
				WorkingDir: "/tmp",
			},
			mockedContents: map[string]string{
				"/tmp/foo.txt": "Title\nGood Bye\ncurrent_version=1.2.3",
			},
			wantedContents: map[string]string{
				"/tmp/foo.txt": "Title\nGood Bye\ncurrent_version=1.2.3",
			},
			wantedResult: true,
		},
		{
			name: "Passing case with 'MatchPattern' specified",
			spec: Spec{
				Files: []string{
					"foo.txt",
				},
				MatchPattern: "current_version.*",
			},
			files: map[string]string{
				"/tmp/foo.txt": "",
			},
			scm: &scm.MockScm{
				WorkingDir: "/tmp",
			},
			mockedContents: map[string]string{
				"/tmp/foo.txt": "Title\nGood Bye\ncurrent_version=1.2.3",
			},
			wantedContents: map[string]string{
				"/tmp/foo.txt": "Title\nGood Bye\ncurrent_version=1.2.3",
			},
			wantedResult: true,
		},
		{
			name: "Validation failure with 'ForceCreate' specified",
			spec: Spec{
				Files: []string{
					"foo.txt",
				},
				ForceCreate: true,
			},
			files: map[string]string{
				"/tmp/foo.txt": "",
			},
			scm: &scm.MockScm{
				WorkingDir: "/tmp",
			},
			inputSourceValue: "1.2.3",
			wantedErr:        true,
		},
		{
			name: "Validation failure with invalid 'Regexp' specified",
			spec: Spec{
				Files: []string{
					"foo.txt",
				},
				MatchPattern: "^^[[[",
			},
			files: map[string]string{
				"/tmp/foo.txt": "",
			},
			scm: &scm.MockScm{
				WorkingDir: "/tmp",
			},
			inputSourceValue: "1.2.3",
			wantedErr:        true,
		},
		{
			name: "Failing case with non matching 'MatchPattern' specified",
			spec: Spec{
				Files: []string{
					"foo.txt",
				},
				MatchPattern: "notMatching.*",
			},
			files: map[string]string{
				"/tmp/foo.txt": "",
			},
			scm: &scm.MockScm{
				WorkingDir: "/tmp",
			},
			mockedContents: map[string]string{
				"/tmp/foo.txt": "Title\r\nGood Bye\r\nThe End",
			},
			wantedContents: map[string]string{
				"/tmp/foo.txt": "Title\r\nGood Bye\r\nThe End",
			},
			wantedResult: false,
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

			gotResult, gotErr := f.ConditionFromSCM(tt.inputSourceValue, tt.scm)
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
