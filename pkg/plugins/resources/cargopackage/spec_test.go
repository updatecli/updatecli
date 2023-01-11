package cargopackage

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		spec    Spec
		wantErr bool
	}{
		{
			name: "Normal case without authentication",
			spec: Spec{
				Package: "test",
			},
		},
		{
			name: "Normal case with username / password",
			spec: Spec{
				Package:  "test",
				Username: "test",
				Password: "test",
			},
		},
		{
			name: "Error case with username and no password",
			spec: Spec{
				Package:  "test",
				Username: "test",
			},
			wantErr: true,
		},
		{
			name: "Normal case with private key",
			spec: Spec{
				Package:    "test",
				PrivateKey: "test",
			},
		},
		{
			name: "Error case with username and private key",
			spec: Spec{
				Package:    "test",
				Username:   "test",
				Password:   "test",
				PrivateKey: "test",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.spec.Validate()
			if tt.wantErr {
				require.Error(t, gotErr)
			} else {
				require.NoError(t, gotErr)
			}
		})
	}
}
