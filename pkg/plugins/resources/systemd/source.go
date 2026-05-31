package systemd

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/updatecli/updatecli/pkg/core/result"
)

func (s *Systemd) Source(_ context.Context, workingDir string, sourceResult *result.Source) error {
	filePath := s.spec.File
	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(workingDir, filePath)
	}

	_, opt, err := s.readOptions(filePath)
	if err != nil {
		return err
	}

	sourceResult.Information = opt.Value
	sourceResult.Result = result.SUCCESS
	sourceResult.Description = fmt.Sprintf("value %q found for option %q in section %q in the systemd unit file %q",
		opt.Value, s.spec.Option, s.spec.Section, filePath)

	return nil
}
