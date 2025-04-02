package dockerfile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (df *Dockerfile) Source(workingDir string, resultSource *result.Source) error {
	// By the default workingdir is set to the current working directory
	// it would be better to have it empty by default but it must be changed in the
	// source core codebase.
	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		return errors.New("fail getting current working directory")
	}

	switch len(df.files) {
	case 1:
		//
	case 0:
		return fmt.Errorf("no dockerfile specified")
	default:
		return fmt.Errorf("validation error in sources of type 'dockerfile': the attributes `spec.files` can't contain more than one element for sources")
	}

	if workingDir == currentWorkingDirectory {
		workingDir = ""
	}

	// loop over the only file
	for _, file := range df.files {
		if workingDir != "" {
			file = filepath.Join(workingDir, file)
		}

		if !df.contentRetriever.FileExists(file) {
			return fmt.Errorf("the file %s does not exist", file)
		}

		dockerfileContent, err := df.contentRetriever.ReadAll(file)
		if err != nil {
			return fmt.Errorf("reading dockerfile: %w", err)
		}

		logrus.Debugf("\n🐋 On (Docker)file %q:\n\n", file)

		value := df.parser.GetInstruction([]byte(dockerfileContent), df.spec.Stage)
		stageInfo := "last stage"
		if df.spec.Stage != "" {
			stageInfo = fmt.Sprintf("stage %q", df.spec.Stage)
		}
		resultSource.Result = result.SUCCESS
		resultSource.Information = value
		resultSource.Description = fmt.Sprintf("value %q found for %s in the dockerfile file %q",
			value,
			stageInfo,
			file,
		)
	}

	return nil
}
