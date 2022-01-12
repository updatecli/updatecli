package shell

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShell_New(t *testing.T) {
	tests := []struct {
		name      string
		spec      Spec
		wantErr   bool
		wantShell *Shell
	}{
		{
			name: "Normal case",
			spec: Spec{
				Command: "echo Hello",
			},
			wantErr: false,
			wantShell: &Shell{
				executor: &nativeCommandExecutor{},
				spec: Spec{
					Command: "echo Hello",
				},
			},
		},
		{
			name: "raises an error when command is empty",
			spec: Spec{
				Command: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotShell, gotErr := New(tt.spec)

			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)

			assert.Equal(t, tt.wantShell, gotShell)
		})
	}
}
