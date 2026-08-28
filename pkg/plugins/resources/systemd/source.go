package systemd

import (
	"context"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils"
)

func (s *Systemd) Source(_ context.Context, resolver utils.Resolver, sourceResult *result.Source) error {
	filePath, err := resolver.Resolve(s.spec.File)
	if err != nil {
		return fmt.Errorf("invalid file path %q: %w", s.spec.File, err)
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
