package shell

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShell_Condition(t *testing.T) {
	tests := []struct {
		name          string
		command       string
		source        string
		wantResult    bool
		wantErr       bool
		wantCommand   string
		commandResult commandResult
	}{
		{
			name:        "Successful Condition",
			command:     "echo Hello",
			source:      "1.2.3",
			wantResult:  true,
			wantErr:     false,
			wantCommand: "echo Hello 1.2.3",
			commandResult: commandResult{
				ExitCode: 0,
				Stdout:   "Hello",
			},
		},
		{
			name:        "Failed Condition",
			command:     "ls",
			source:      "1.2.3",
			wantResult:  false,
			wantErr:     false,
			wantCommand: "ls 1.2.3",
			commandResult: commandResult{
				ExitCode: 1,
				Stderr:   "ls: 1.2.3: No such file or directory",
			},
		},
		{
			name:       "Empty command with empty source",
			command:    "",
			source:     "",
			wantResult: false,
			wantErr:    true,
		},
		{
			name:       "Empty command with non empty source",
			command:    "",
			source:     "1.2.3",
			wantResult: false,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockCommandExecutor{
				result: tt.commandResult,
			}
			s := Shell{
				executor: &mock,
				spec: ShellSpec{
					Command: tt.command,
				},
			}

			gotResult, err := s.Condition(tt.source)

			if tt.wantErr {
				assert.Error(t, err)
				assert.False(t, gotResult)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantResult, gotResult)

			assert.Equal(t, tt.wantCommand, mock.gotCommand.Cmd)
		})
	}
}

func TestShell_ConditionFromSCM(t *testing.T) {
	tests := []struct {
		name          string
		command       string
		source        string
		scmDir        string
		wantResult    bool
		wantErr       bool
		wantCommand   string
		commandResult commandResult
	}{
		{
			name:        "Successful Condition in existing SCM",
			command:     "echo Hello",
			source:      "1.2.3",
			scmDir:      "/dummy/dir",
			wantResult:  true,
			wantErr:     false,
			wantCommand: "echo Hello 1.2.3",
			commandResult: commandResult{
				ExitCode: 0,
				Stdout:   "Hello",
			},
		},
		{
			name:        "Failed Condition in existing SCM",
			command:     "ls",
			source:      "1.2.3",
			scmDir:      "/dummy/dir",
			wantResult:  false,
			wantErr:     false,
			wantCommand: "ls 1.2.3",
			commandResult: commandResult{
				ExitCode: 1,
				Stderr:   "ls: 1.2.3: No such file or directory",
			},
		},
		{
			name:       "Empty command with non empty source in existing SCM",
			command:    "",
			source:     "1.2.3",
			scmDir:     "/dummy/dir",
			wantResult: false,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mce := mockCommandExecutor{
				result: tt.commandResult,
			}
			ms := mockScm{
				workingDir: tt.scmDir,
			}
			s := Shell{
				executor: &mce,
				spec: ShellSpec{
					Command: tt.command,
				},
			}

			gotResult, err := s.ConditionFromSCM(tt.source, &ms)

			if tt.wantErr {
				assert.Error(t, err)
				assert.False(t, gotResult)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantResult, gotResult)

			assert.Equal(t, tt.wantCommand, mce.gotCommand.Cmd)
			assert.Equal(t, tt.scmDir, mce.gotCommand.Dir)
		})
	}
}
