package yaml

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/text"
)

func Test_Target(t *testing.T) {
	tests := []struct {
		spec                  YamlSpec
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
			spec: YamlSpec{
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
			spec: YamlSpec{
				File:  "https://github.com/foo.yaml",
				Key:   "github.owner",
				Value: "obiwankenobi",
			},
			wantResult: false,
			wantErr:    true,
		},
		{
			name: "Not passing: file already up to date",
			spec: YamlSpec{
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
			spec: YamlSpec{
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
			spec: YamlSpec{
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
				Spec:             tt.spec,
				contentRetriever: &mockText,
			}
			gotResult, gotErr := y.Target(tt.inputSourceValue, tt.dryRun)
			if tt.wantErr {
				assert.Error(t, gotErr)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.wantResult, gotResult)
		})
	}
}
