package systemd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		spec    Spec
		wantErr string
	}{
		{
			name: "Valid spec",
			spec: Spec{
				File:    "test.container",
				Section: "Container",
				Option:  "Image",
			},
		},
		{
			name: "Missing file",
			spec: Spec{
				Section: "Container",
				Option:  "Image",
			},
			wantErr: "the attribute `spec.file` is required",
		},
		{
			name: "Missing section",
			spec: Spec{
				File:   "test.container",
				Option: "Image",
			},
			wantErr: "the attribute `spec.section` is required",
		},
		{
			name: "Missing option",
			spec: Spec{
				File:    "test.container",
				Section: "Container",
			},
			wantErr: "the attribute `spec.option` is required",
		},
		{
			name:    "All fields missing",
			spec:    Spec{},
			wantErr: "the attribute `spec.file` is required",
		},
		{
			name: "Negative index",
			spec: Spec{
				File:    "test.container",
				Section: "Container",
				Option:  "Image",
				Index:   -1,
			},
			wantErr: "the attribute `spec.index` must be greater than or equal to 0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.spec.Validate()
			if tt.wantErr != "" {
				assert.Error(t, gotErr)
				assert.Contains(t, gotErr.Error(), tt.wantErr)
				return
			}

			assert.NoError(t, gotErr)
		})
	}
}
