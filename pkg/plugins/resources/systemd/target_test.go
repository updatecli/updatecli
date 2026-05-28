package systemd

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
)

func TestTarget(t *testing.T) {
	tests := []struct {
		name             string
		spec             Spec
		source           string
		dryRun           bool
		mockedContents   map[string]string
		expectedResult   string
		expectChange     bool
		expectedContents map[string]string
		wantErr          bool
	}{
		{
			name:   "Dry run with value change",
			source: "nginx:1.26",
			spec: Spec{
				File:    "test.container",
				Section: "Container",
				Option:  "Image",
			},
			dryRun: true,
			mockedContents: map[string]string{
				"test.container": "[Unit]\nDescription=test\n\n[Container]\nImage=nginx:1.25\n",
			},
			expectedResult: result.ATTENTION,
			expectChange:   true,
			expectedContents: map[string]string{
				"test.container": "[Unit]\nDescription=test\n\n[Container]\nImage=nginx:1.25\n",
			},
		},
		{
			name:   "Real run with value change",
			source: "nginx:1.26",
			spec: Spec{
				File:    "test.container",
				Section: "Container",
				Option:  "Image",
			},
			dryRun: false,
			mockedContents: map[string]string{
				"test.container": "[Unit]\nDescription=test\n\n[Container]\nImage=nginx:1.25\n",
			},
			expectedResult: result.ATTENTION,
			expectChange:   true,
			expectedContents: map[string]string{
				"test.container": "[Unit]\nDescription=test\n\n[Container]\nImage=nginx:1.26\n",
			},
		},
		{
			name:   "Value already matches",
			source: "nginx:1.25",
			spec: Spec{
				File:    "test.container",
				Section: "Container",
				Option:  "Image",
			},
			dryRun: false,
			mockedContents: map[string]string{
				"test.container": "[Unit]\nDescription=test\n\n[Container]\nImage=nginx:1.25\n",
			},
			expectedResult: result.SUCCESS,
			expectChange:   false,
			expectedContents: map[string]string{
				"test.container": "[Unit]\nDescription=test\n\n[Container]\nImage=nginx:1.25\n",
			},
		},
		{
			name:   "File does not exist",
			source: "nginx:1.25",
			spec: Spec{
				File:    "nonexistent.container",
				Section: "Container",
				Option:  "Image",
			},
			mockedContents: map[string]string{},
			wantErr:        true,
		},
		{
			name:   "Option not found in file",
			source: "nginx:1.25",
			spec: Spec{
				File:    "test.container",
				Section: "Container",
				Option:  "NonExistent",
			},
			mockedContents: map[string]string{
				"test.container": "[Unit]\nDescription=test\n\n[Container]\nImage=nginx:1.25\n",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockedText := text.MockTextRetriever{
				Contents: tt.mockedContents,
			}
			s := &Systemd{
				spec:             tt.spec,
				contentRetriever: &mockedText,
			}

			gotResult := result.Target{}
			gotErr := s.Target(context.Background(), tt.source, nil, tt.dryRun, &gotResult)
			if tt.wantErr {
				assert.Error(t, gotErr)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.expectedResult, gotResult.Result)
			assert.Equal(t, tt.expectChange, gotResult.Changed)
			assert.Equal(t, "nginx:1.25", gotResult.Information)
			assert.Equal(t, tt.source, gotResult.NewInformation)
			assert.Equal(t, []string{"test.container"}, gotResult.Files)

			for filePath, expectedContent := range tt.expectedContents {
				assert.Equal(t, expectedContent, mockedText.Contents[filePath])
			}
		})
	}
}
