package shell

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/scm"
)

// mockCommandExecutor is a stub implementation of the `commandExecutor` interface
// to be used in our test suite.
// It stores the received `command` and returns the preconfigured `result` and `err`.
type mockCommandExecutor struct {
	gotCommand command
	result     commandResult
	err        error
}

func (mce *mockCommandExecutor) ExecuteCommand(cmd command) (commandResult, error) {
	mce.gotCommand = cmd
	return mce.result, mce.err
}

// mocking SCM object (no introspection: only get values)
type mockScm struct {
	scm.Scm

	workingDir   string
	changedFiles []string
	err          error
}

func (m *mockScm) GetDirectory() (directory string) {
	return m.workingDir
}

func (m *mockScm) GetChangedFiles(workingDir string) ([]string, error) {
	m.workingDir = workingDir
	return m.changedFiles, m.err
}

func TestShell_New(t *testing.T) {
	tests := []struct {
		name      string
		spec      ShellSpec
		wantErr   bool
		wantShell *Shell
	}{
		{
			name: "Normal case",
			spec: ShellSpec{
				Command: "echo Hello",
			},
			wantErr: false,
			wantShell: &Shell{
				executor: &nativeCommandExecutor{},
				spec: ShellSpec{
					Command: "echo Hello",
				},
			},
		},
		{
			name: "raises an error when command is empty",
			spec: ShellSpec{
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
