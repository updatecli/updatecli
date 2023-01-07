package shell

import (
	"crypto/sha256"
	"fmt"
	"io"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/tmp"
)

// wanteScriptFilename is an utility used to get the filescript named generated
// by Updatecli. Outside of testing, it's not supposed to be used by Updatecli
// as it ignore error handling
func wantedScriptFilename(t *testing.T, command string) string {
	h := sha256.New()
	_, err := io.WriteString(h, command)

	require.NoError(t, err)

	return filepath.Join(tmp.BinDirectory, fmt.Sprintf("%x", h.Sum(nil)))
}

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
				Shell:   "/bin/bash",
			},
			wantErr: false,
			wantShell: &Shell{
				executor:    &nativeCommandExecutor{},
				interpreter: "/bin/bash",
				spec: Spec{
					Command: "echo Hello",
					Shell:   "/bin/bash",
				},
				scriptFilename: wantedScriptFilename(t, "echo Hello"),
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
				interpreter:    getDefaultShell(),
				scriptFilename: wantedScriptFilename(t, "echo Hello"),
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
				executor:    &nativeCommandExecutor{},
				interpreter: getDefaultShell(),
				spec: Spec{
					Command: "echo Hello",
					Environments: Environments{
						Environment{
							Name: "PATH",
						},
					},
				},
				scriptFilename: wantedScriptFilename(t, "echo Hello"),
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
				interpreter:    getDefaultShell(),
				scriptFilename: wantedScriptFilename(t, "echo Hello"),
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
					Shell:   "/bin/sh",
					Environments: Environments{
						Environment{
							Name: "PATH",
						},
					},
				},
				interpreter:    getDefaultShell(),
				scriptFilename: wantedScriptFilename(t, "echo Hello"),
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

			assert.Equal(t, tt.wantShell, gotShell)
		})
	}
}
