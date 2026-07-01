package systemd

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
)

func TestSource(t *testing.T) {
	tests := []struct {
		name           string
		spec           Spec
		mockedContents map[string]string
		mockedError    error
		expectedResult string
		wantErr        bool
	}{
		{
			name: "Found option in Container section",
			spec: Spec{
				File:    "test.container",
				Section: "Container",
				Option:  "Image",
			},
			mockedContents: map[string]string{
				"test.container": "[Unit]\nDescription=test\n\n[Container]\nImage=nginx:1.25\n",
			},
			expectedResult: "nginx:1.25",
		},
		{
			name: "Found option in Service section",
			spec: Spec{
				File:    "test.service",
				Section: "Service",
				Option:  "ExecStart",
			},
			mockedContents: map[string]string{
				"test.service": "[Unit]\nDescription=test\n\n[Service]\nExecStart=/usr/bin/myapp\nRestart=always\n",
			},
			expectedResult: "/usr/bin/myapp",
		},
		{
			name: "Found repeated option by index",
			spec: Spec{
				File:    "test.container",
				Section: "Container",
				Option:  "Volume",
				Index:   intPtr(1),
			},
			mockedContents: map[string]string{
				"test.container": "[Container]\nVolume=/lib/modules:/lib/modules:ro\nVolume=/etc/wg-easy:/etc/wireguard:rw\n",
			},
			expectedResult: "/etc/wg-easy:/etc/wireguard:rw",
		},
		{
			name: "File does not exist",
			spec: Spec{
				File:    "nonexistent.container",
				Section: "Container",
				Option:  "Image",
			},
			mockedContents: map[string]string{},
			wantErr:        true,
		},
		{
			name: "Option not found in file",
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
		{
			name: "Section not found in file",
			spec: Spec{
				File:    "test.container",
				Section: "Install",
				Option:  "WantedBy",
			},
			mockedContents: map[string]string{
				"test.container": "[Unit]\nDescription=test\n\n[Container]\nImage=nginx:1.25\n",
			},
			wantErr: true,
		},
		{
			name: "Option index not found in file",
			spec: Spec{
				File:    "test.container",
				Section: "Container",
				Option:  "Volume",
				Index:   intPtr(2),
			},
			mockedContents: map[string]string{
				"test.container": "[Container]\nVolume=/lib/modules:/lib/modules:ro\nVolume=/etc/wg-easy:/etc/wireguard:rw\n",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockedText := text.MockTextRetriever{
				Contents: tt.mockedContents,
				Err:      tt.mockedError,
			}
			s := &Systemd{
				spec:             tt.spec,
				contentRetriever: &mockedText,
			}

			gotResult := result.Source{}
			gotErr := s.Source(context.Background(), "", &gotResult)
			if tt.wantErr {
				assert.Error(t, gotErr)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, result.SUCCESS, gotResult.Result)
			assert.Equal(t, tt.expectedResult, gotResult.Information)
		})
	}
}
