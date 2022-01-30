package yaml

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/text"
)

func Test_Target(t *testing.T) {
	tests := []struct {
		spec                  Spec
		wantMockState         text.MockTextRetriever
		name                  string
		inputSourceValue      string
		mockReturnedContent   string
		mockReturnedError     error
		mockReturnsFileExists bool
		wantResult            bool
		wantErr               bool
		dryRun                bool
	}{
		{
			name: "Passing Case with both input source and specified value (specified value should be used)",
			spec: Spec{
				File:  "test.yaml",
				Key:   "github.owner",
				Value: "obiwankenobi",
			},
			mockReturnsFileExists: true,
			inputSourceValue:      "olblak",
			mockReturnedContent: `---
github:
  owner: olblak
  repository: charts
`,
			wantResult: true,
		},
		{
			name: "Validation failure with an https:// URL instead of a file",
			spec: Spec{
				File:  "https://github.com/foo.yaml",
				Key:   "github.owner",
				Value: "obiwankenobi",
			},
			wantResult: false,
			wantErr:    true,
		},
		{
			name: "Not passing: file already up to date",
			spec: Spec{
				File: "test.yaml",
				Key:  "github.owner",
			},
			mockReturnsFileExists: true,
			inputSourceValue:      "olblak",
			mockReturnedContent: `---
github:
  owner: olblak
  repository: charts
`,
			wantResult: false,
		},
		{
			name: "Provided key does not exist",
			spec: Spec{
				File:  "test.yaml",
				Key:   "github.ship",
				Value: "obiwankenobi",
			},
			mockReturnsFileExists: true,
			inputSourceValue:      "olblak",
			mockReturnedContent: `---
github:
  owner: olblak
  repository: charts
`,
			wantResult: false,
			wantErr:    true,
		},
		{
			name: "Invalid YAML file",
			spec: Spec{
				File:  "test.yaml",
				Key:   "github.ship",
				Value: "obiwankenobi",
			},
			mockReturnsFileExists: true,
			inputSourceValue:      "olblak",
			mockReturnedContent: `---
github-
  owner: olblak
  repository: charts
`,
			wantResult: false,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockText := text.MockTextRetriever{
				Content: tt.mockReturnedContent,
				Err:     tt.mockReturnedError,
				Exists:  tt.mockReturnsFileExists,
			}
			y := &Yaml{
				spec:             tt.spec,
				contentRetriever: &mockText,
			}
			gotResult, gotErr := y.Target(tt.inputSourceValue, tt.dryRun)
			defer os.Remove(y.spec.File)
			if tt.wantErr {
				assert.Error(t, gotErr)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.wantResult, gotResult)
		})
	}
}
