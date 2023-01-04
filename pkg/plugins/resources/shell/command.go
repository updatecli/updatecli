package shell

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/tmp"
)

type command struct {
	Shell string
	Cmd   string
	Dir   string
	Env   []string
}

type commandResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Cmd      string
}

type commandExecutor interface {
	ExecuteCommand(cmd command) (commandResult, error)
}

type nativeCommandExecutor struct{}

func (nce *nativeCommandExecutor) ExecuteCommand(inputCmd command) (commandResult, error) {
	scriptName, err := newShellScript(inputCmd)
	if err != nil {
		return commandResult{}, err
	}

	var stdout, stderr bytes.Buffer

	logrus.Debugf("\tcommand: %s %s\n", inputCmd.Shell, scriptName)

	command := exec.Command(inputCmd.Shell, scriptName)
	command.Dir = inputCmd.Dir
	command.Stdout = &stdout
	command.Stderr = &stderr
	// Pass current environment to process and append the customized environment variables used internally by updatecli (such as DRY_RUN)
	command.Env = append(command.Env, inputCmd.Env...)
	err = command.Run()

	// Display environment variables in debug mode
	logrus.Debugf("Environment variables\n")
	for _, env := range command.Env {
		logrus.Debugf("\t* %s\n", env)
	}

	// Remove line returns from stdout
	out := strings.TrimSuffix(stdout.String(), "\n")

	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return commandResult{
			ExitCode: ee.ExitCode(),
			Stdout:   out,
			Stderr:   stderr.String(),
			Cmd:      inputCmd.Cmd,
		}, nil
	}
	if err != nil {
		return commandResult{}, err
	}

	return commandResult{
		ExitCode: 0,
		Stdout:   out,
		Stderr:   stderr.String(),
		Cmd:      inputCmd.Cmd,
	}, nil
}

func newShellScript(inputCmd command) (string, error) {
	// Ensure Updatecli bin directory exists
	bindDir, err := tmp.InitBin()
	if err != nil {
		return "", err
	}

	// Generate uniq script name
	h := sha256.New()
	_, err = io.WriteString(h, inputCmd.Cmd)
	if err != nil {
		return "", err
	}

	scriptFilename := filepath.Join(bindDir, fmt.Sprintf("%x", h.Sum(nil)))

	// Save command in script name
	f, err := os.Create(scriptFilename)
	if err != nil {
		return "", err
	}

	defer f.Close()

	_, err = f.WriteString(inputCmd.Cmd)
	if err != nil {
		return "", err
	}

	return scriptFilename, nil
}
