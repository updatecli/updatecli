package systemd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/updatecli/updatecli/pkg/core/result"
)

func (s *Systemd) Source(_ context.Context, workingDir string, sourceResult *result.Source) error {
	filePath := s.spec.File

	// By the default workingdir is set to the current working directory
	// it would be better to have it empty by default but it must be changed in the
	// source core codebase.
	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("fail getting current working directory: %w", err)
	}

	if workingDir == currentWorkingDirectory {
		workingDir = ""
	}

	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(workingDir, filePath)
	}

	_, matchingOpts, err := s.readOptions(filePath)
	if err != nil {
		return fmt.Errorf("fail reading systemd unit file %q: %w", filePath, err)
	}

	if matchingOpts == nil {
		sourceResult.Result = result.FAILURE
		sourceResult.Description = fmt.Sprintf("option %q not found in section %q in the systemd unit file %q",
			s.spec.Option, s.spec.Section, filePath)
		return nil
	}

	optIndex := 0
	if s.spec.Index != nil {
		optIndex = *s.spec.Index
	}

	if optIndex >= len(matchingOpts) {
		sourceResult.Result = result.FAILURE
		sourceResult.Description = fmt.Sprintf("option %q with index %d not found in section %q in the systemd unit file %q",
			s.spec.Option, optIndex, s.spec.Section, filePath)

		return fmt.Errorf("index not found in systemd file")
	}

	sourceResult.Information = matchingOpts[optIndex].Value
	sourceResult.Result = result.SUCCESS
	sourceResult.Description = fmt.Sprintf("value %q found for option %q in section %q in the systemd unit file %q",
		matchingOpts[optIndex].Value, s.spec.Option, s.spec.Section, filePath)

	return nil
}
