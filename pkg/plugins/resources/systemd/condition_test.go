package systemd

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/text"
)

func TestCondition(t *testing.T) {
	tests := []struct {
		name           string
		spec           Spec
		source         string
		mockedContents map[string]string
		shouldPass     bool
		wantErr        bool
	}{
		{
			name:   "Value matches from source",
			source: "nginx:1.25",
			spec: Spec{
				File:    "test.container",
				Section: "Container",
				Option:  "Image",
			},
			mockedContents: map[string]string{
				"test.container": "[Unit]\nDescription=test\n\n[Container]\nImage=nginx:1.25\n",
			},
			shouldPass: true,
		},
		{
			name:   "Value matches from spec.Value",
			source: "",
			spec: Spec{
				File:    "test.container",
				Section: "Container",
				Option:  "Image",
				Value:   "nginx:1.25",
			},
			mockedContents: map[string]string{
				"test.container": "[Unit]\nDescription=test\n\n[Container]\nImage=nginx:1.25\n",
			},
			shouldPass: true,
		},
		{
			name:   "Value does not match",
			source: "nginx:1.26",
			spec: Spec{
				File:    "test.container",
				Section: "Container",
				Option:  "Image",
			},
			mockedContents: map[string]string{
				"test.container": "[Unit]\nDescription=test\n\n[Container]\nImage=nginx:1.25\n",
			},
			shouldPass: false,
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

			pass, _, err := s.Condition(context.Background(), tt.source, nil)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.shouldPass, pass)
		})
	}
}
