package shell

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	defaultFakeValue = "xxx"
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
				Command:      "echo Hello",
				Shell:        "/bin/bash",
				Environments: &Environments{},
			},
			wantErr: false,
			wantShell: &Shell{
				executor:     &nativeCommandExecutor{},
				interpreter:  "/bin/bash",
				environments: Environments{},
				spec: Spec{
					Command:      "echo Hello",
					Shell:        "/bin/bash",
					Environments: &Environments{},
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
				Command: "echo World",
				Environments: &Environments{
					Environment{
						Value: &defaultFakeValue,
					},
				},
			},
			wantErr: true,
			wantShell: &Shell{
				executor: &nativeCommandExecutor{},
				environments: Environments{
					Environment{
						Value: &defaultFakeValue,
					},
				},
				spec: Spec{
					Command: "echo World",
					Environments: &Environments{
						Environment{
							Value: &defaultFakeValue,
						},
					},
				},
				interpreter: getDefaultShell(),
			},
		},
		{
			name: "Inherit PATH environment variable",
			spec: Spec{
				Command: "echo Hello",
				Environments: &Environments{
					Environment{
						Name: "PATH",
					},
				},
			},
			wantErr: false,
			wantShell: &Shell{
				executor:    &nativeCommandExecutor{},
				interpreter: getDefaultShell(),
				environments: Environments{
					Environment{
						Name: "PATH",
					},
				},
				spec: Spec{
					Command: "echo Hello",
					Environments: &Environments{
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
				Environments: &Environments{
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
				environments: Environments{
					Environment{
						Name: "PATH",
					},
					Environment{
						Name: "PATH",
					},
				},
				spec: Spec{
					Command: "echo Hello",
					Environments: &Environments{
						Environment{
							Name: "PATH",
						},
						Environment{
							Name: "PATH",
						},
					},
				},
				interpreter: getDefaultShell(),
			},
		},
		{
			name: "Not allowed to specify DRY_RUN environment variable",
			spec: Spec{
				Command: "echo Hello",
				Environments: &Environments{
					Environment{
						Name: "DRY_RUN",
					},
				},
			},
			wantErr: true,
			wantShell: &Shell{
				executor: &nativeCommandExecutor{},
				environments: Environments{
					Environment{
						Name: "DRY_RUN",
					},
				},
				spec: Spec{
					Command: "echo Hello",
					Shell:   "/bin/sh",
					Environments: &Environments{
						Environment{
							Name: "PATH",
						},
					},
				},
				interpreter: getDefaultShell(),
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

			// Init Success Criteria as I couldn't manage to do it via the tests
			gotErr = tt.wantShell.InitChangedIf()
			require.NoError(t, gotErr)

			assert.Equal(t, tt.wantShell, gotShell)
		})
	}
}

func TestGetWorkingDir(t *testing.T) {
	testdata := []struct {
		name               string
		spec               Spec
		currentWorkingDir  string
		expectedWorkingDir string
	}{
		{
			name: "Expecting merged path",
			spec: Spec{
				WorkDir: "pkg",
				Command: "true",
			},
			currentWorkingDir:  "/projects/updatecli",
			expectedWorkingDir: "/projects/updatecli/pkg",
		},
		{
			name: "Expecting spec absolute path",
			spec: Spec{
				WorkDir: "/pkg",
				Command: "true",
			},
			currentWorkingDir:  "/projects/updatecli",
			expectedWorkingDir: "/pkg",
		},
		{
			name: "Expecting current working dir path",
			spec: Spec{
				Command: "true",
			},
			currentWorkingDir:  "/projects/updatecli",
			expectedWorkingDir: "/projects/updatecli",
		},
	}

	for _, tt := range testdata {

		t.Run(tt.name, func(t *testing.T) {
			s, err := New(tt.spec)
			require.NoError(t, err)

			gotWorkingdir := s.getWorkingDirPath(tt.currentWorkingDir)
			assert.Equal(t, tt.expectedWorkingDir, gotWorkingdir)
		})
	}
}
