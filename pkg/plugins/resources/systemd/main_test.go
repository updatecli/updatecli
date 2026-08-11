package systemd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		spec    Spec
		want    Spec
		wantErr string
	}{
		{
			name: "Missing section and option",
			spec: Spec{
				File: "test.container",
			},
			wantErr: "the attribute `spec.section` is required",
		},
		{
			name: "Missing option",
			spec: Spec{
				File:    "test.container",
				Section: "Service",
			},
			wantErr: "the attribute `spec.option` is required",
		},
		{
			name: "Explicit section and option",
			spec: Spec{
				File:    "test.service",
				Section: "Service",
				Option:  "ExecStart",
			},
			want: Spec{
				File:    "test.service",
				Section: "Service",
				Option:  "ExecStart",
			},
		},
		{
			name: "Missing file",
			spec: Spec{
				Section: "Service",
				Option:  "ExecStart",
			},
			wantErr: "the attribute `spec.file` is required.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := New(tt.spec)
			if tt.wantErr != "" {
				require.Error(t, gotErr)
				assert.Contains(t, gotErr.Error(), tt.wantErr)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.want, got.spec)
		})
	}
}
