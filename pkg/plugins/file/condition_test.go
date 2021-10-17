package file

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/scm"
	"github.com/updatecli/updatecli/pkg/core/text"
)

func TestFile_Condition(t *testing.T) {
	tests := []struct {
		name                string
		spec                FileSpec
		inputSourceValue    string
		wantResult          bool
		wantErr             bool
		wantMockState       text.MockTextRetriever
		mockReturnedContent string
		mockReturnedError   error
	}{
		{
			name: "Passing Case with Line",
			spec: FileSpec{
				Line: 3,
				File: "foo.txt",
			},
			inputSourceValue:    "current_version=1.2.3",
			mockReturnedContent: "current_version=1.2.3",
			wantResult:          true,
			wantMockState: text.MockTextRetriever{
				Line:     3,
				Location: "foo.txt",
			},
		},
		{
			name: "File does not exist",
			spec: FileSpec{
				Line: 3,
				File: "not_existing.txt",
			},
			mockReturnedError: fmt.Errorf("no such file or directory"),
			wantResult:        false,
			wantErr:           true,
		},
		{
			name: "Failing Case with Line",
			spec: FileSpec{
				Line: 5,
				File: "/bar.txt",
			},
			inputSourceValue:    "current_version=1.2.3",
			mockReturnedContent: "current_version=1.2.4",
			wantMockState: text.MockTextRetriever{
				Line:     5,
				Location: "/bar.txt",
			},
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
			name: "Passing Case with no input source and only specified content",
			spec: FileSpec{
				Content: "Hello World",
				File:    "foo.txt",
			},
			mockReturnedContent: "Hello World",
			wantResult:          true,
			wantMockState: text.MockTextRetriever{
				Location: "foo.txt",
			},
		},
		{
			name: "Case with no input source, no specified content but a specified line which is empty",
			spec: FileSpec{
				Line: 11,
				File: "foo.txt",
			},
			wantResult: false,
			wantMockState: text.MockTextRetriever{
				Line:     11,
				Location: "foo.txt",
			},
		},
		{
			name: "Case with no input source, no specified content but the specified line exists and is not empty",
			spec: FileSpec{
				Line: 13,
				File: "bar.txt",
			},
			mockReturnedContent: "Something On The Line",
			wantResult:          true,
			wantMockState: text.MockTextRetriever{
				Line:     13,
				Location: "bar.txt",
			},
		},
		{
			name: "Case with no input source but a specified content which matches the file content",
			spec: FileSpec{
				Content: "Hello World",
				File:    "foo.txt",
			},
			mockReturnedContent: "Hello World",
			wantResult:          true,
			wantMockState: text.MockTextRetriever{
				Location: "foo.txt",
			},
		},
		{
			name: "Failing case with no input source and a specified line that does not matches the file line content",
			spec: FileSpec{
				Line:    11,
				Content: "Not In The File",
				File:    "foo.txt",
			},
			inputSourceValue:    "",
			mockReturnedContent: "Hello World",
			wantResult:          false,
			wantMockState: text.MockTextRetriever{
				Location: "foo.txt",
				Line:     11,
			},
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
			gotResult, gotErr := f.Condition(tt.inputSourceValue)
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
