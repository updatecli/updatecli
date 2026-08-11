package systemd

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/coreos/go-systemd/v22/unit"
	"github.com/sirupsen/logrus"
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

	opts, matchingOpts, err := s.readOptions(filePath)
	if err != nil {
		return err
	}

	if matchingOpts == nil {
		resultTarget.Result = result.FAILURE
		resultTarget.Description = fmt.Sprintf("option %q not found in section %q in the systemd unit file %q",
			s.spec.Option, s.spec.Section, filePath)
		return nil
	}

	optIndex := 0
	if s.spec.Index != nil {
		optIndex = *s.spec.Index
	}

	if optIndex >= len(matchingOpts) {
		resultTarget.Result = result.FAILURE
		resultTarget.Description = fmt.Sprintf("option %q with index %d not found in section %q in the systemd unit file %q",
			s.spec.Option, optIndex, s.spec.Section, filePath)
		return fmt.Errorf("index not found in systemd file")
	}

	resultTarget.NewInformation = expected
	resultTarget.Files = []string{filePath}

	totalNotMatchingOpts := []int{}
	switch s.spec.Index {
	// nil means that we want to update all the options matching the section/option and not only the one matching the index.
	case nil:
		for i, opt := range matchingOpts {
			if opt.Section == s.spec.Section && opt.Name == s.spec.Option {
				resultTarget.Information = opt.Value
				if expected != opt.Value {
					logrus.Debugf("option %d %q value %q doesn't match expected value %q",
						i, opt.Name, opt.Value, expected)
					totalNotMatchingOpts = append(totalNotMatchingOpts, i)
				}
			}
		}

		if len(totalNotMatchingOpts) == 0 {
			resultTarget.Description = fmt.Sprintf("option %q value(s) match expected value %q for file %q",
				s.spec.Option, expected, s.spec.File)
			resultTarget.Result = result.SUCCESS
			resultTarget.Information = expected

			return nil
		}

	default:
		resultTarget.Information = expected

		if expected == matchingOpts[optIndex].Value {
			resultTarget.Description = fmt.Sprintf("option %q value %q matches expected value %q for file %q",
				matchingOpts[optIndex].Name, matchingOpts[optIndex].Value, expected, s.spec.File)
			resultTarget.Result = result.SUCCESS

			return nil
		}
	}

	resultTarget.Changed = true
	resultTarget.Result = result.ATTENTION

	if dryRun {
		resultTarget.Description = fmt.Sprintf("option %q value %q should be updated to %q for file %q",
			matchingOpts[optIndex].Name, matchingOpts[optIndex].Value, expected, s.spec.File)
		return nil
	}

	resultTarget.Description = fmt.Sprintf("option %q value %q doesn't match expected value %q for file %q",
		matchingOpts[optIndex].Name, matchingOpts[optIndex].Value, expected, s.spec.File)

	switch s.spec.Index {
	// nil means that we want to update all the options matching the section/option and not only the one matching the index.
	case nil:
		for _, i := range totalNotMatchingOpts {
			resultTarget.Information = strings.Join([]string{resultTarget.Information, matchingOpts[i].Value}, ", ")

			oldValue := matchingOpts[i].Value
			matchingOpts[i].Value = expected
			resultTarget.Description = fmt.Sprintf("option %q value %q updated to %q for file %q",
				matchingOpts[i].Name, oldValue, expected, s.spec.File)
		}
	default:
		resultTarget.Information = matchingOpts[optIndex].Value

		matchingOpts[optIndex].Value = expected

		oldValue := matchingOpts[optIndex].Value
		matchingOpts[optIndex].Value = expected
		resultTarget.Description = fmt.Sprintf("option %q value %q updated to %q for file %q",
			matchingOpts[optIndex].Name, oldValue, expected, s.spec.File)

	}

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
