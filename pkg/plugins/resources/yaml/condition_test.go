package yaml

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/text"
)

func Test_Condition(t *testing.T) {
	tests := []struct {
		name                string
		spec                Spec
		inputSourceValue    string
		wantResult          bool
		wantErr             bool
		wantMockState       text.MockTextRetriever
		mockReturnedContent string
		mockReturnedError   error
	}{
		{
			name: "Passing Case",
			spec: Spec{
				File: "test.yaml",
				Key:  "github.owner",
			},
			inputSourceValue: "olblak",
			mockReturnedContent: `---
github:
  owner: olblak
  repository: charts
`,
			wantResult: true,
			wantMockState: text.MockTextRetriever{
				Location: "test.yaml",
			},
		},
		{
			name: "Passing Case with keyonly and input source",
			spec: Spec{
				File:    "test.yaml",
				Key:     "github.owner",
				KeyOnly: true,
			},
			inputSourceValue: "olblak",
			mockReturnedContent: `---
github:
  owner: olblak
  repository: charts
`,
			wantResult: true,
			wantMockState: text.MockTextRetriever{
				Location: "test.yaml",
			},
		},
		{
			name: "Failing case with keyonly and input source",
			spec: Spec{
				File:    "test.yaml",
				Key:     "github.country",
				KeyOnly: true,
			},
			inputSourceValue: "",
			mockReturnedContent: `---
github:
  owner: olblak
  repository: charts
`,
			wantResult: false,
			wantMockState: text.MockTextRetriever{
				Location: "test.yaml",
			},
		},
		{
			name: "Validation error with both keyonly and specified value",
			spec: Spec{
				File:    "test.yaml",
				Key:     "github.owner",
				KeyOnly: true,
				Value:   "olblak",
			},
			inputSourceValue: "",
			mockReturnedContent: `---
github:
  owner: olblak
  repository: charts
`,
			wantResult: false,
			wantErr:    true,
			wantMockState: text.MockTextRetriever{
				Location: "test.yaml",
			},
		},
		{
			name: "File does not exist",
			spec: Spec{
				File: "not_existing.txt",
			},
			mockReturnedError: fmt.Errorf("no such file or directory"),
			wantResult:        false,
			wantErr:           true,
		},
		{
			name: "Failing Case (key found but not the correct value)",
			spec: Spec{
				File: "test.yaml",
				Key:  "github.owner",
			},
			inputSourceValue: "asterix",
			mockReturnedContent: `---
github:
  owner: olblak
  repository: charts
`,
			wantResult: false,
			wantMockState: text.MockTextRetriever{
				Location: "test.yaml",
			},
		},
		{
			name: "Failing Case (key not found)",
			spec: Spec{
				File: "test.yaml",
				Key:  "github.admin",
			},
			inputSourceValue: "asterix",
			mockReturnedContent: `---
github:
  owner: olblak
  repository: charts
`,
			wantResult: false,
			wantErr:    true,
			wantMockState: text.MockTextRetriever{
				Location: "test.yaml",
			},
		},
		{
			name: "Validation Failure with both source and specified value",
			spec: Spec{
				File:  "test.yaml",
				Key:   "github.owner",
				Value: "asterix",
			},
			inputSourceValue: "olblak",
			wantErr:          true,
		},
		{
			name: "Failure due to unvalid Yaml",
			spec: Spec{
				File: "test.yaml",
				Key:  "github.owner",
			},
			inputSourceValue: "olblak",
			mockReturnedContent: `---
github
  owner: olblak
  repository: charts
`,
			wantErr: true,
		},
		{
			name: "Passing Case with no input source and only specified value",
			spec: Spec{
				File:  "test.yaml",
				Key:   "github.owner",
				Value: "olblak",
			},
			mockReturnedContent: `---
github:
  owner: olblak
  repository: charts
`,
			wantResult: true,
			wantMockState: text.MockTextRetriever{
				Location: "test.yaml",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockText := text.MockTextRetriever{
				Content: tt.mockReturnedContent,
				Err:     tt.mockReturnedError,
				Exists:  true,
			}
			y := &Yaml{
				spec:             tt.spec,
				contentRetriever: &mockText,
			}
			gotResult, gotErr := y.Condition(tt.inputSourceValue)
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

func Test_ConditionFromSCM(t *testing.T) {
	tests := []struct {
		name                string
		spec                Spec
		inputSourceValue    string
		wantResult          bool
		wantErr             bool
		wantMockState       text.MockTextRetriever
		mockReturnedContent string
		mockReturnedError   error
		scm                 scm.ScmHandler
	}{
		{
			name: "Passing Case with no input source and only specified value",
			spec: Spec{
				File:  "test.yaml",
				Key:   "github.owner",
				Value: "olblak",
			},
			mockReturnedContent: `---
github:
  owner: olblak
  repository: charts
`,
			scm: &scm.MockScm{
				WorkingDir: "/tmp",
			},
			wantResult: true,
			wantMockState: text.MockTextRetriever{
				Location: "/tmp/test.yaml",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockText := text.MockTextRetriever{
				Content: tt.mockReturnedContent,
				Err:     tt.mockReturnedError,
				Exists:  true,
			}
			y := &Yaml{
				spec:             tt.spec,
				contentRetriever: &mockText,
			}
			gotResult, gotErr := y.ConditionFromSCM(tt.inputSourceValue, tt.scm)
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
