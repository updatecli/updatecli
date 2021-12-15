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
		spec                YamlSpec
		inputSourceValue    string
		wantResult          bool
		wantErr             bool
		wantMockState       text.MockTextRetriever
		mockReturnedContent string
		mockReturnedError   error
	}{
		{
			name: "Passing Case",
			spec: YamlSpec{
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
			name: "File does not exist",
			spec: YamlSpec{
				File: "not_existing.txt",
			},
			mockReturnedError: fmt.Errorf("no such file or directory"),
			wantResult:        false,
			wantErr:           true,
		},
		{
			name: "Failing Case (key found but not the correct value)",
			spec: YamlSpec{
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
			spec: YamlSpec{
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
			wantMockState: text.MockTextRetriever{
				Location: "test.yaml",
			},
		},
		{
			name: "Validation Failure with both source and specified value",
			spec: YamlSpec{
				File:  "test.yaml",
				Key:   "github.owner",
				Value: "asterix",
			},
			inputSourceValue: "olblak",
			wantErr:          true,
		},
		{
			name: "Failure due to unvalid Yaml",
			spec: YamlSpec{
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
			spec: YamlSpec{
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
			}
			y := &Yaml{
				Spec:             tt.spec,
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
		spec                YamlSpec
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
			spec: YamlSpec{
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
			}
			y := &Yaml{
				Spec:             tt.spec,
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
