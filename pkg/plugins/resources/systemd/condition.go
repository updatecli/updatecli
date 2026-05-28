package systemd

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

func (s *Systemd) Condition(_ context.Context, source string, scm scm.ScmHandler) (pass bool, message string, err error) {
	expected := source
	if expected == "" {
		expected = s.spec.Value
	}

	rootDir := ""
	if scm != nil {
		rootDir = scm.GetDirectory()
	}

	filePath := s.spec.File
	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(rootDir, filePath)
	}

	_, opt, err := s.readOptions(filePath)
	if err != nil {
		return false, "", err
	}

	if expected != opt.Value {
		return false, fmt.Sprintf("option %q value %q doesn't match expected value %q",
			opt.Name, opt.Value, expected), nil
	}

	return true, fmt.Sprintf("option %q matching value %q found in section %q",
		opt.Name, opt.Value, opt.Section), nil
}
