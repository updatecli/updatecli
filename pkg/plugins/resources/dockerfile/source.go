package dockerfile

import (
	"errors"
	"fmt"
	"os"

	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerfile/types"
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

		stageValues := df.parser.GetInstruction([]byte(dockerfileContent))
		stageName := "last stage"
		if len(stageValues) == 0 {
			return fmt.Errorf("could not get source value %q for %s in the dockerfile %q",
				resultSource.Name,
				stageName,
				file,
			)
		}
		var value *types.StageInstructionValue
		if df.spec.Stage != "" {
			for _, stageInstruction := range stageValues {
				if stageInstruction.StageName == df.spec.Stage {
					value = &stageInstruction
				}
				break
			}
			stageName := fmt.Sprintf("stage %q", df.spec.Stage)
			if value == nil {
				return fmt.Errorf("could not find %s in %q", stageName, file)
			}
		} else {
			// No Stage name provider, using last
			value = &stageValues[len(stageValues)-1]
		}
		resultSource.Result = result.SUCCESS
		resultSource.Information = value.Value
		resultSource.Description = fmt.Sprintf("value %q found for %s in the dockerfile file %q",
			value.Value,
			stageName,
			file,
		)
		return nil

	}
	return fmt.Errorf("Source is not supported for the plugin 'dockerfile'")
}
