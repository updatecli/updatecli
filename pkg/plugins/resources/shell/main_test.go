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
		{
			name: "Missing env name despite env value specified",
			spec: Spec{
				Command: "echo Hello",
				Environments: Environments{
					Environment{
						Value: "xxx",
					},
				},
			},
			wantErr: true,
			wantShell: &Shell{
				executor: &nativeCommandExecutor{},
				spec: Spec{
					Command: "echo Hello",
					Environments: Environments{
						Environment{
							Value: "xxx",
						},
					},
				},
			},
		},
		{
			name: "Inherit PATH environment variable",
			spec: Spec{
				Command: "echo Hello",
				Environments: Environments{
					Environment{
						Name: "PATH",
					},
				},
			},
			wantErr: false,
			wantShell: &Shell{
				executor: &nativeCommandExecutor{},
				spec: Spec{
					Command: "echo Hello",
					Environments: Environments{
						Environment{
							Name: "PATH",
						},
					},
				},
			},
		},
		{
			name: "can't specify PATH environment variable multiple times",
			spec: Spec{
				Command: "echo Hello",
				Environments: Environments{
					Environment{
						Name: "PATH",
					},
					Environment{
						Name: "PATH",
					},
				},
			},
			wantErr: true,
			wantShell: &Shell{
				executor: &nativeCommandExecutor{},
				spec: Spec{
					Command: "echo Hello",
					Environments: Environments{
						Environment{
							Name: "PATH",
						},
						Environment{
							Name: "PATH",
						},
					},
				},
			},
		},
		{
			name: "Not allowed to specify DRY_RUN environment variable",
			spec: Spec{
				Command: "echo Hello",
				Environments: Environments{
					Environment{
						Name: "DRY_RUN",
					},
				},
			},
			wantErr: true,
			wantShell: &Shell{
				executor: &nativeCommandExecutor{},
				spec: Spec{
					Command: "echo Hello",
					Environments: Environments{
						Environment{
							Name: "PATH",
						},
					},
				},
			},
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

			// InitOutcome as I couldn't manage to do it via the tests
			gotErr = tt.wantShell.InitOutcome()
			require.NoError(t, gotErr)

			assert.Equal(t, tt.wantShell, gotShell)
		})
	}
}
