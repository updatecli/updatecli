package shell

import (
	"crypto/sha256"
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/tmp"
)

// wantedScriptFilename is an utility used to get the filescript named generated
// by Updatecli. Outside of testing, it's not supposed to be used by Updatecli
// as it ignore error handling
func wantedScriptFilename(t *testing.T, command string) string {
	h := sha256.New()
	_, err := io.WriteString(h, command)

	require.NoError(t, err)

	suffix := ""

	switch runtime.GOOS {
	case WINOS:
		suffix = ".ps1"
	default:
		suffix = ".sh"

	}

	return filepath.Join(tmp.BinDirectory, fmt.Sprintf("%x%s", h.Sum(nil), suffix))
}

func TestShell_Source(t *testing.T) {
	tests := []struct {
		name          string
		command       string
		shell         string
		workingDir    string
		wantSource    string
		wantCommand   string
		wantErr       bool
		commandResult commandResult
	}{
		{
			name:        "Get a source from a successful command in working directory",
			command:     "echo Hello",
			shell:       "/bin/bash",
			workingDir:  "/home/ucli",
			wantSource:  "Hello",
			wantCommand: "/bin/bash" + " " + wantedScriptFilename(t, "echo Hello"),
			wantErr:     false,
			commandResult: commandResult{
				ExitCode: 0,
				Stdout:   "Hello",
			},
		},
		{
			name:       "Raise an error with a failing command in working directory",
			command:    "false",
			shell:      "/bin/bash",
			workingDir: "/home/ucli",
			wantSource: "",
			wantErr:    true,
			commandResult: commandResult{
				ExitCode: 1,
			},
		},
		{
			name:       "Raise an error with an empty command in working directory",
			command:    "",
			shell:      "/bin/bash",
			workingDir: "/home/ucli",
			wantSource: "",
			wantErr:    true,
			commandResult: commandResult{
				ExitCode: 1,
			},
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
			gotErr := s.InitChangedIf()
			require.NoError(t, gotErr)

			gotResult := result.Source{}
			err := s.Source(tt.workingDir, &gotResult)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.wantSource, gotResult.Information)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantSource, gotResult.Information)

			assert.Equal(t, tt.wantCommand, mock.GotCommand.Cmd)
		})
	}
}
