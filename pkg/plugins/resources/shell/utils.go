package shell

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/updatecli/updatecli/pkg/core/tmp"
)

// newShellScript copies the command to a temporary shell script located in
// the updatecli temparory working directory.
// This technique allows to executed complex command directly from an Updatecli
// manifest
func newShellScript(command string) (string, error) {
	// Ensure Updatecli bin directory exists
	bindDir, err := tmp.InitBin()
	if err != nil {
		return "", err
	}

	// Generate uniq script name
	h := sha256.New()
	_, err = io.WriteString(h, command)
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

	_, err = f.WriteString(command)
	if err != nil {
		return "", err
	}

	return scriptFilename, nil
}
