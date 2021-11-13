package file

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/text"
)

func TestFile_Condition(t *testing.T) {
	tests := []struct {
		name             string
		spec             FileSpec
		inputSourceValue string
		wantResult       bool
		wantErr          bool
		mockTest         text.MockTextRetriever
	}{
		{
			name: "Passing Case with Line",
			spec: FileSpec{
				Line: 3,
				File: "foo.txt",
			},
			inputSourceValue: "current_version=1.2.3",
			mockTest: text.MockTextRetriever{
				Content: "current_version=1.2.3",
				Exists:  true,
			},
			wantResult: true,
		},
		{
			name: "Failing Case with Line",
			spec: FileSpec{
				Line: 5,
				File: "/bar.txt",
			},
			mockTest: text.MockTextRetriever{
				Content: "current_version=1.2.4",
				Exists:  true,
			},
			inputSourceValue: "current_version=1.2.3",
		},
		{
			name: "Validation Failure with both source and specified content",
			spec: FileSpec{
				Content: "Hello World",
				File:    "/bar.txt",
			},
			inputSourceValue: "1.2.3",
			wantErr:          true,
		},
		{
			name: "Validation Failure with specified ReplacePattern",
			spec: FileSpec{
				MatchPattern:   "maven_(.*)",
				ReplacePattern: "gradle_$1",
				File:           "/bar.txt",
			},
			inputSourceValue: "1.2.3",
			wantErr:          true,
		},
		{
			name: "Passing Case with no input source and only specified content",
			spec: FileSpec{
				Content: "Hello World",
				File:    "foo.txt",
			},
			mockTest: text.MockTextRetriever{
				Content: "Hello World",
				Exists:  true,
			},
			wantResult: true,
		},
		{
			name: "Case with no input source, no specified content but a specified line which is empty",
			spec: FileSpec{
				Line: 11,
				File: "foo.txt",
			},
			wantResult: false,
			mockTest: text.MockTextRetriever{
				Line:    11,
				Content: "",
				Exists:  true,
			},
		},
		{
			name: "Case with no input source, no specified content but the specified line exists and is not empty",
			spec: FileSpec{
				Line: 13,
				File: "bar.txt",
			},
			mockTest: text.MockTextRetriever{
				Line:    13,
				Content: "Something On The Line",
				Exists:  true,
			},
			wantResult: true,
		},
		{
			name: "Failing case with only file existence checking",
			spec: FileSpec{
				File: "foo.txt",
			},
			mockTest: text.MockTextRetriever{
				Exists: false,
			},
			wantResult: false,
		},
		{
			name: "Failing case with only URL existence checking",
			spec: FileSpec{
				File: "https://do.not.exists/foo",
			},
			mockTest: text.MockTextRetriever{
				Err: fmt.Errorf("URL %q not found or in error", "https://do.not.exists/foo"),
			},
			wantResult: false,
			wantErr:    true,
		},
		{
			name: "Failing case with no input source and a specified line that does not matches the file line content",
			spec: FileSpec{
				Line:    11,
				Content: "Not In The File",
				File:    "foo.txt",
			},
			inputSourceValue: "",
			mockTest: text.MockTextRetriever{
				Content: "Hello World",
			},
			wantResult: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockText := tt.mockTest
			f := &File{
				spec:             tt.spec,
				contentRetriever: &mockText,
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
		spec                FileSpec
		inputSourceValue    string
		wantResult          bool
		wantErr             bool
		wantMockState       text.MockTextRetriever
		mockReturnedContent string
		mockReturnedError   error
		scm                 scm.Scm
	}{
		{
			name: "Passing Case with no input source, but a specified line and content and a relative path to file",
			spec: FileSpec{
				File:    "foo.txt",
				Content: "current_version=1.2.3",
				Line:    3,
			},
			scm: &scm.MockScm{
				WorkingDir: "/tmp",
			},
			mockReturnedContent: "current_version=1.2.3",
			wantResult:          true,
			wantMockState: text.MockTextRetriever{
				Location: "/tmp/foo.txt",
				Line:     3,
			},
		},
		{
			name: "Passing Case with matchPattern",
			spec: FileSpec{
				File:         "foo.txt",
				MatchPattern: "current_version.*",
			},
			scm: &scm.MockScm{
				WorkingDir: "/tmp",
			},
			mockReturnedContent: "current_version=1.2.3",
			wantResult:          true,
			wantMockState: text.MockTextRetriever{
				Location: "/tmp/foo.txt",
			},
		},
		{
			name: "Failing Case with matchPattern",
			spec: FileSpec{
				File:         "foo.txt",
				MatchPattern: "notMatching.*",
			},
			scm: &scm.MockScm{
				WorkingDir: "/tmp",
			},
			mockReturnedContent: "current_version=1.2.3",
			wantResult:          false,
			wantMockState: text.MockTextRetriever{
				Location: "/tmp/foo.txt",
			},
		},
		{
			name: "Validation Failure with forcecreate specified",
			spec: FileSpec{
				File:        "foo.txt",
				ForceCreate: true,
			},
			scm: &scm.MockScm{
				WorkingDir: "/tmp",
			},
			inputSourceValue: "1.2.3",
			wantErr:          true,
		},
		{
			name: "Validation Failure with invalid Regexp",
			spec: FileSpec{
				File:         "foo.txt",
				MatchPattern: "^^[[[",
			},
			scm: &scm.MockScm{
				WorkingDir: "/tmp",
			},
			inputSourceValue: "1.2.3",
			wantErr:          true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockText := text.MockTextRetriever{
				Content: tt.mockReturnedContent,
				Err:     tt.mockReturnedError,
			}
			f := &File{
				spec:             tt.spec,
				contentRetriever: &mockText,
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
