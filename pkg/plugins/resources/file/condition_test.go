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
		mockContents     map[string]string
		mockError        error
		wantResult       bool
		wantErr          bool
	}{
		// {
		// 	name: "Passing case with 'Line' specified",
		// 	spec: Spec{
		// 		Files: []string{
		// 			"foo.txt",
		// 		},
		// 		Line: 3,
		// 	},
		// 	files: map[string]string{
		// 		"foo.txt": "",
		// 	},
		// 	inputSourceValue: "current_version=1.2.3",
		// 	mockContents: map[string]string{
		// 		"foo.txt": "Hello World\r\nAnother line\r\ncurrent_version=1.2.3",
		// 	},
		// 	wantResult: true,
		// },
		// {
		// 	name: "Passing case with 'Content' specified and no source specified",
		// 	spec: Spec{
		// 		Files: []string{
		// 			"foo.txt",
		// 		},
		// 		Content: "Hello World",
		// 	},
		// 	files: map[string]string{
		// 		"foo.txt": "",
		// 	},
		// 	mockContents: map[string]string{
		// 		"foo.txt": "Hello World",
		// 	},
		// 	wantResult: true,
		// },
		// {
		// 	name: "Validation failure with more than one element in 'Files'",
		// 	spec: Spec{
		// 		Files: []string{
		// 			"foo.txt",
		// 			"bar.txt",
		// 		},
		// 	},
		// 	files: map[string]string{
		// 		"foo.txt": "",
		// 		"bar.txt": "",
		// 	},
		// 	wantErr: true,
		// },
		// {
		// 	name: "Validation failure with both source and 'Content' specified",
		// 	spec: Spec{
		// 		Files: []string{
		// 			"foo.txt",
		// 		},
		// 		Content: "Hello World",
		// 	},
		// 	files: map[string]string{
		// 		"foo.txt": "",
		// 	},
		// 	inputSourceValue: "1.2.3",
		// 	wantErr:          true,
		// },
		// {
		// 	name: "Validation failure with 'ReplacePattern' specified",
		// 	spec: Spec{
		// 		Files: []string{
		// 			"foo.txt",
		// 		},
		// 		MatchPattern:   "maven_(.*)",
		// 		ReplacePattern: "gradle_$1",
		// 	},
		// 	files: map[string]string{
		// 		"foo.txt": "",
		// 	},
		// 	inputSourceValue: "1.2.3",
		// 	wantErr:          true,
		// },
		// {
		// 	name: "Case with empty 'Line' specified and no source nor 'Content' specified",
		// 	spec: Spec{
		// 		Files: []string{
		// 			"foo.txt",
		// 		},
		// 		Line: 11,
		// 	},
		// 	files: map[string]string{
		// 		"foo.txt": "",
		// 	},
		// 	mockContents: map[string]string{
		// 		"foo.txt": "",
		// 	},
		// 	wantResult: false,
		// },
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
			mockContents: map[string]string{
				"foo.txt": "A first line\r\nAnother line\r\nSomething On The Specified Line",
			},
			wantResult: true,
		},
		{
			name: "Failing case with non existent 'Line' specified",
			spec: Spec{
				Files: []string{
					"foo.txt",
				},
				Line: 5,
			},
			files: map[string]string{
				"foo.txt": "",
			},
			wantErr: true,
		},
		{
			name: "Failing case with non existent 'Files'",
			spec: Spec{
				Files: []string{
					"foo.txt",
				},
			},
			files: map[string]string{
				"foo.txt": "",
			},
			wantErr: true,
		},
		{
			name: "Failing case with non existent URL as 'Files'",
			spec: Spec{
				Files: []string{
					"https://do.not.exists/foo",
				},
			},
			files: map[string]string{
				"https://do.not.exists/foo": "",
			},
			mockError: fmt.Errorf("URL %q not found or in error", "https://do.not.exists/foo"),
			wantErr:   true,
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
			mockContents: map[string]string{
				"foo.txt": "Title\r\nGood Bye\r\nThe End",
			},
			wantResult: false,
			wantErr:    false,
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
			mockContents: map[string]string{
				"foo.txt": "Title\r\nGood Bye\r\nThe End",
				"bar.txt": "Title\r\nNot In All Files\r\nThe End",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockText := text.MockTextRetriever{
				Contents: tt.mockContents,
			}
			f := &File{
				spec:             tt.spec,
				contentRetriever: &mockText,
				files:            tt.files,
			}

			gotResult, gotErr := f.Condition(tt.inputSourceValue)
			if tt.wantErr {
				assert.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, tt.wantResult, gotResult)
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
		mockContents     map[string]string
		mockLines        map[string]int
		mockError        error
		wantMockContents map[string]string
		wantMockLines    map[string]int
		wantResult       bool
		wantErr          bool
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
			mockContents: map[string]string{
				"/tmp/foo.txt": "Title\nGood Bye\ncurrent_version=1.2.3",
			},
			wantMockContents: map[string]string{
				"/tmp/foo.txt": "Title\nGood Bye\ncurrent_version=1.2.3",
			},
			wantResult: true,
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
			mockContents: map[string]string{
				"/tmp/foo.txt": "Title\nGood Bye\ncurrent_version=1.2.3",
			},
			wantMockContents: map[string]string{
				"/tmp/foo.txt": "Title\nGood Bye\ncurrent_version=1.2.3",
			},
			wantResult: true,
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
			wantErr:          true,
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
			wantErr:          true,
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
			mockContents: map[string]string{
				"/tmp/foo.txt": "Title\r\nGood Bye\r\nThe End",
			},
			wantMockContents: map[string]string{
				"/tmp/foo.txt": "Title\r\nGood Bye\r\nThe End",
			},
			wantResult: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockText := text.MockTextRetriever{
				Contents: tt.mockContents,
				Lines:    tt.mockLines,
			}
			f := &File{
				spec:             tt.spec,
				contentRetriever: &mockText,
				files:            tt.files,
			}

			gotResult, gotErr := f.ConditionFromSCM(tt.inputSourceValue, tt.scm)
			if tt.wantErr {
				assert.Error(t, gotErr)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.wantResult, gotResult)
			for filePath := range tt.files {
				assert.Equal(t, tt.wantMockContents[filePath], mockText.Contents[filePath])
				assert.Equal(t, tt.wantMockLines[filePath], mockText.Lines[filePath])
			}
		})
	}
}
