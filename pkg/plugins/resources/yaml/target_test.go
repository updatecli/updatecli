package yaml

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/text"
)

const sampleYamlContent string = `---
github:
  owner: olblak
  repository: charts
...`

func Test_Target(t *testing.T) {
	tests := []struct {
		spec                Spec
		name                string
		inputSourceValue    string
		dryRun              bool
		workingDir          string
		mockReturnedContent string
		wantResult          bool
		wantFiles           []string
		wantMessage         string
		wantErr             bool
		wantMockState       text.MockTextRetriever
	}{
		{
			name: "Passing case with relative working dir and both input source and specified value (specified value should be used)",
			spec: Spec{
				File:  "test.yaml",
				Key:   "github.owner",
				Value: "obiwankenobi",
			},
			inputSourceValue:    "olblak",
			workingDir:          "",
			mockReturnedContent: sampleYamlContent,
			wantResult:          true,
			wantFiles:           []string{"test.yaml"},
			wantMessage:         "Update the YAML key \"github.owner\" from file \"test.yaml\"",
		},
		{
			name: "Passing case with absolute working dir and both input source and specified value (specified value should be used)",
			spec: Spec{
				File:  "test.yaml",
				Key:   "github.owner",
				Value: "obiwankenobi",
			},
			inputSourceValue:    "olblak",
			workingDir:          "/tmp",
			mockReturnedContent: sampleYamlContent,
			wantResult:          true,
			wantFiles:           []string{"/tmp/test.yaml"},
			wantMessage:         "Update the YAML key \"github.owner\" from file \"/tmp/test.yaml\"",
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
			inputSourceValue:    "olblak",
			mockReturnedContent: sampleYamlContent,
			wantResult:          false,
		},
		{
			name: "Provided key does not exist",
			spec: Spec{
				File:  "test.yaml",
				Key:   "github.ship",
				Value: "obiwankenobi",
			},
			inputSourceValue:    "olblak",
			mockReturnedContent: sampleYamlContent,
			wantResult:          false,
			wantErr:             true,
		},
		{
			name: "Invalid YAML file",
			spec: Spec{
				File:  "test.yaml",
				Key:   "github.ship",
				Value: "obiwankenobi",
			},
			inputSourceValue:    "olblak",
			mockReturnedContent: sampleYamlContent,
			wantResult:          false,
			wantErr:             true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockText := &text.MockTextRetriever{
				Contents: map[string]string{
					filepath.Join(tt.workingDir, tt.spec.File): tt.mockReturnedContent,
				},
			}
			y := &Yaml{
				spec:             tt.spec,
				contentRetriever: mockText,
			}
			gotResult, gotFiles, gotMessage, gotErr := y.Target(tt.inputSourceValue, tt.workingDir, tt.dryRun)
			defer os.Remove(y.spec.File)
			if tt.wantErr {
				assert.Error(t, gotErr)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.wantResult, gotResult)
			assert.Equal(t, tt.wantFiles, gotFiles)
			assert.Equal(t, tt.wantMessage, gotMessage)
		})
	}
}
