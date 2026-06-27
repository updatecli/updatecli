package systemd

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/coreos/go-systemd/v22/unit"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (s *Systemd) Target(_ context.Context, source string, scm scm.ScmHandler, dryRun bool, resultTarget *result.Target) error {
	expected := s.spec.Value
	if expected == "" {
		expected = source
	}

	rootDir := ""
	if scm != nil {
		rootDir = scm.GetDirectory()
	}

	filePath := s.spec.File
	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(rootDir, filePath)
	}

	opts, opt, err := s.readOptions(filePath)
	if err != nil {
		return err
	}

	resultTarget.Information = opt.Value
	resultTarget.NewInformation = expected
	resultTarget.Files = []string{filePath}

	if expected == opt.Value {
		resultTarget.Result = result.SUCCESS
		return nil
	}

	resultTarget.Changed = true
	resultTarget.Result = result.ATTENTION

	if dryRun {
		return nil
	}

	opt.Value = expected

	reader := unit.Serialize(opts)

	data, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("reading serialized unit file: %w", err)
	}

	err = s.contentRetriever.WriteToFile(string(data), filePath)
	if err != nil {
		return fmt.Errorf("writing systemd unit file: %w", err)
	}

	return nil
}
