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
		mockTest         text.MockTextRetriever
		wantResult       bool
		wantErr          bool
	}{
		{
			name: "Passing case with 'Line' specified",
			spec: Spec{
				Line: 3,
				File: "foo.txt",
			},
			files: map[string]string{
				"foo.txt": "",
			},
			inputSourceValue: "current_version=1.2.3",
			mockTest: text.MockTextRetriever{
				Content: "current_version=1.2.3",
				Exists:  true,
			},
			wantResult: true,
		},
		{
			name: "Passing case with 'Content' specified and no source specified",
			spec: Spec{
				Content: "Hello World",
				File:    "foo.txt",
			},
			files: map[string]string{
				"foo.txt": "",
			},
			mockTest: text.MockTextRetriever{
				Content: "Hello World",
				Exists:  true,
			},
			wantResult: true,
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
			wantErr: true,
		},
		{
			name: "Validation failure with both source and 'Content' specified",
			spec: Spec{
				Content: "Hello World",
				File:    "foo.txt",
			},
			files: map[string]string{
				"foo.txt": "",
			},
			inputSourceValue: "1.2.3",
			wantErr:          true,
		},
		{
			name: "Validation failure with 'ReplacePattern' specified",
			spec: Spec{
				MatchPattern:   "maven_(.*)",
				ReplacePattern: "gradle_$1",
				File:           "foo.txt",
			},
			files: map[string]string{
				"foo.txt": "",
			},
			inputSourceValue: "1.2.3",
			wantErr:          true,
		},
		{
			name: "Case with empty 'Line' specified and no source nor 'Content' specified",
			spec: Spec{
				Line: 11,
				File: "foo.txt",
			},
			files: map[string]string{
				"foo.txt": "",
			},
			wantResult: false,
			mockTest: text.MockTextRetriever{
				Line:    11,
				Content: "",
				Exists:  true,
			},
		},
		{
			name: "Case with not empty 'Line' specified and no source nor 'Content' specified",
			spec: Spec{
				Line: 13,
				File: "foo.txt",
			},
			files: map[string]string{
				"foo.txt": "",
			},
			mockTest: text.MockTextRetriever{
				Line:    13,
				Content: "Something On The Line",
				Exists:  true,
			},
			wantResult: true,
		},
		{
			name: "Failing case with non existant 'Line' specified",
			spec: Spec{
				Line: 5,
				File: "foo.txt",
			},
			files: map[string]string{
				"foo.txt": "",
			},
			mockTest: text.MockTextRetriever{
				Exists: false,
			},
			wantErr: true,
		},
		{
			name: "Failing case with non existant 'File'",
			spec: Spec{
				File: "foo.txt",
			},
			files: map[string]string{
				"foo.txt": "",
			},
			mockTest: text.MockTextRetriever{
				Exists: false,
			},
			wantResult: false,
			wantErr:    true,
		},
		{
			name: "Failing case with non existant URL as 'File'",
			spec: Spec{
				File: "https://do.not.exists/foo",
			},
			files: map[string]string{
				"https://do.not.exists/foo": "",
			},
			mockTest: text.MockTextRetriever{
				Err:    fmt.Errorf("URL %q not found or in error", "https://do.not.exists/foo"),
				Exists: false,
			},
			wantResult: false,
			wantErr:    true,
		},
		{
			name: "Failing case with not empty 'Line' not matching the 'Content' of the 'File' at the 'Line' specified",
			spec: Spec{
				Line:    11,
				Content: "Not In The File",
				File:    "foo.txt",
			},
			files: map[string]string{
				"foo.txt": "",
			},
			inputSourceValue: "",
			mockTest: text.MockTextRetriever{
				Content: "Hello World",
			},
			wantResult: false,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockText := tt.mockTest
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
		name                string
		spec                Spec
		files               map[string]string
		inputSourceValue    string
		scm                 scm.ScmHandler
		mockReturnedContent string
		mockReturnedError   error
		mockFileExist       bool
		wantMockState       text.MockTextRetriever
		wantResult          bool
		wantErr             bool
	}{
		{
			name: "Passing case with both 'Line' and 'Content' specified, 'File' with a relative path, and no source",
			spec: Spec{
				File:    "foo.txt",
				Content: "current_version=1.2.3",
				Line:    3,
			},
			files: map[string]string{
				"foo.txt": "",
			},
			scm: &scm.MockScm{
				WorkingDir: "/tmp",
			},
			mockReturnedContent: "current_version=1.2.3",
			mockFileExist:       true,
			wantResult:          true,
			wantMockState: text.MockTextRetriever{
				Location: "/tmp/foo.txt",
				Line:     3,
			},
		},
		{
			name: "Passing case with 'MatchPattern' specified",
			spec: Spec{
				File:         "foo.txt",
				MatchPattern: "current_version.*",
			},
			files: map[string]string{
				"foo.txt": "",
			},
			scm: &scm.MockScm{
				WorkingDir: "/tmp",
			},
			mockReturnedContent: "current_version=1.2.3",
			mockFileExist:       true,
			wantResult:          true,
			wantMockState: text.MockTextRetriever{
				Location: "/tmp/foo.txt",
			},
		},
		{
			name: "Validation failure with 'ForceCreate' specified",
			spec: Spec{
				File:        "foo.txt",
				ForceCreate: true,
			},
			files: map[string]string{
				"foo.txt": "",
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
				File:         "foo.txt",
				MatchPattern: "^^[[[",
			},
			files: map[string]string{
				"foo.txt": "",
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
				File:         "foo.txt",
				MatchPattern: "notMatching.*",
			},
			files: map[string]string{
				"foo.txt": "",
			},
			scm: &scm.MockScm{
				WorkingDir: "/tmp",
			},
			mockReturnedContent: "current_version=1.2.3",
			mockFileExist:       true,
			wantResult:          false,
			wantMockState: text.MockTextRetriever{
				Location: "/tmp/foo.txt",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockText := text.MockTextRetriever{
				Content: tt.mockReturnedContent,
				Err:     tt.mockReturnedError,
				Exists:  tt.mockFileExist,
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
			assert.Equal(t, tt.wantMockState.Location, mockText.Location)
			assert.Equal(t, tt.wantMockState.Line, mockText.Line)
		})
	}
}
