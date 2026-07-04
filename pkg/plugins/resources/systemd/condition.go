package systemd

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

func (s *Systemd) Condition(_ context.Context, source string, scm scm.ScmHandler) (pass bool, message string, err error) {
	expected := s.spec.Value
	if expected == "" {
		// Override the default value with the source output when the `spec.value` is not set.
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

	_, matchingOpts, err := s.readOptions(filePath)
	if err != nil {
		return false, "", err
	}

	if matchingOpts == nil {
		return false, fmt.Sprintf("option %q not found in section %q in the systemd unit file %q",
			s.spec.Option, s.spec.Section, filePath), nil
	}

	optIndex := 0
	if s.spec.Index != nil {
		optIndex = *s.spec.Index
	}

	switch s.spec.Index {
	case nil:
		totalNotMatchingOpts := 0
		for i, opt := range matchingOpts {
			if opt.Section == s.spec.Section && opt.Name == s.spec.Option {
				if expected != opt.Value {
					logrus.Debugf("option %d %q value %q doesn't match expected value %q",
						i, opt.Name, opt.Value, expected)
					totalNotMatchingOpts++
				}
			}
		}

		if totalNotMatchingOpts > 0 {
			return false, fmt.Sprintf("%d option(s) %q value(s) doesn't match expected value %q",
				totalNotMatchingOpts, s.spec.Option, expected), nil
		}

	default:

		if optIndex >= len(matchingOpts) {
			return false, fmt.Sprintf("option %q with index %d not found in section %q in the systemd unit file %q",
				s.spec.Option, optIndex, s.spec.Section, filePath), nil
		}

		if matchingOpts[optIndex].Value != expected {
			return false, fmt.Sprintf("option %q value %q doesn't match expected value %q",
				matchingOpts[optIndex].Name, matchingOpts[optIndex].Value, expected), nil
		}
	}

	return true, fmt.Sprintf("option %q matching value %q found in section %q",
		matchingOpts[optIndex].Name, matchingOpts[optIndex].Value, matchingOpts[optIndex].Section), nil
}
