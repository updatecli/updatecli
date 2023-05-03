package shell

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func TestShell_Target(t *testing.T) {
	tests := []struct {
		name              string
		command           string
		shell             string
		source            string
		dryrun            bool
		wantChanged       bool
		wantErr           bool
		wantCommandInMock string
		commandResult     commandResult
		commandEnv        []string
	}{
		{
			name:              "runs a target that does not change anything and no dryrun",
			command:           "do_not_change.sh",
			shell:             "/bin/bash",
			source:            "1.2.3",
			wantChanged:       false,
			wantErr:           false,
			wantCommandInMock: "/bin/bash" + " " + wantedScriptFilename(t, "do_not_change.sh 1.2.3"),
			commandResult: commandResult{
				ExitCode: 0,
				Stdout:   "",
			},
			commandEnv: []string{"DRY_RUN=false"},
		},
		{
			name:              "runs a target that changes a value and no dryrun",
			command:           "change.sh",
			source:            "1.2.3",
			wantChanged:       true,
			wantErr:           false,
			wantCommandInMock: "/bin/bash" + " " + wantedScriptFilename(t, "change.sh 1.2.3"),
			commandResult: commandResult{
				ExitCode: 0,
				Stdout:   "1.2.3",
			},
			commandEnv: []string{"DRY_RUN=false"},
			shell:      "/bin/bash",
		},
		{
			name:              "runs a target with exit code 2 and no dryrun",
			command:           "change.sh",
			source:            "1.2.3",
			wantChanged:       false,
			wantErr:           true,
			wantCommandInMock: "/bin/bash" + " " + wantedScriptFilename(t, "change.sh 1.2.3"),
			commandResult: commandResult{
				ExitCode: 2,
				Stderr:   "Error: unable to change value to 1.2.3.",
			},
			commandEnv: []string{"DRY_RUN=false"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := MockCommandExecutor{
				Result: tt.commandResult,
			}
			s := Shell{
				executor: &mock,
				spec: Spec{
					Command: tt.command,
					Shell:   tt.shell,
				},
				interpreter: tt.shell,
			}

			// InitSuccess Criteria
			err := s.InitChangedIf()
			require.NoError(t, err)

			gotResult := result.Target{}

			err = s.Target(tt.source, nil, tt.dryrun, &gotResult)

			if tt.wantErr {
				assert.Error(t, err)
				assert.False(t, gotResult.Changed)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantChanged, gotResult.Changed)

			assert.Equal(t, tt.wantCommandInMock, mock.GotCommand.Cmd)
			for _, wantEnv := range tt.commandEnv {
				assert.Contains(t, mock.GotCommand.Env, wantEnv)
			}
		})
	}
}

func TestShell_TargetFromSCM(t *testing.T) {
	tests := []struct {
		commandResult            commandResult
		wantFilesChanged         []string
		commandEnv               []string
		name                     string
		command                  string
		source                   string
		scmDir                   string
		mockReturnedChangedFiles []string
		wantMessage              string
		wantCommandInMock        string
		wantErr                  bool
		dryrun                   bool
		shell                    string
	}{
		{
			name:                     "runs a target that changes a value and no dryrun",
			command:                  "change.sh",
			source:                   "1.2.3",
			mockReturnedChangedFiles: []string{"pom.xml"},
			wantMessage:              `ran shell command "change.sh 1.2.3"`,
			wantErr:                  false,
			wantCommandInMock:        "/bin/bash" + " " + wantedScriptFilename(t, "change.sh 1.2.3"),
			commandResult: commandResult{
				ExitCode: 0,
				Stdout:   "Changed value from 1.2.2 to 1.2.3.",
			},
			wantFilesChanged: []string{"pom.xml"},
			commandEnv:       []string{"DRY_RUN=false"},
			shell:            "/bin/bash",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := MockCommandExecutor{
				Result: tt.commandResult,
			}
			ms := scm.MockScm{
				WorkingDir:   tt.scmDir,
				ChangedFiles: tt.mockReturnedChangedFiles,
			}
			s := Shell{
				executor: &mock,
				spec: Spec{
					Command: tt.command,
					Shell:   tt.shell,
				},
				interpreter: tt.shell,
			}

			// InitSuccess Criteria
			err := s.InitChangedIf()
			require.NoError(t, err)

			gotResult := result.Target{}
			err = s.Target(tt.source, &ms, tt.dryrun, &gotResult)

			if tt.wantErr {
				assert.Error(t, err)
				assert.False(t, gotResult.Changed)
				return
			}

			require.NoError(t, err)

			assert.Equal(t, len(tt.wantFilesChanged) > 0, gotResult.Changed)
			assert.Equal(t, tt.wantFilesChanged, gotResult.Files)
			assert.Equal(t, tt.wantMessage, gotResult.Description)

			assert.Equal(t, tt.wantCommandInMock, mock.GotCommand.Cmd)
			for _, wantEnv := range tt.commandEnv {
				assert.Contains(t, mock.GotCommand.Env, wantEnv)
			}
		})
	}
}
