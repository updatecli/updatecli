package systemd

import (
	"fmt"
	"strings"

	"github.com/coreos/go-systemd/v22/unit"
	"github.com/go-viper/mapstructure/v2"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
)

// Systemd defines a resource of kind "systemd"
type Systemd struct {
	spec             Spec
	contentRetriever text.TextRetriever
}

// New returns a reference to a newly initialized Systemd object from a Spec
// or an error if the provided Spec triggers a validation error.
func New(spec any) (*Systemd, error) {
	newSpec := Spec{}
	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return nil, err
	}

	if newSpec.Section == "" {
		newSpec.Section = "Container"
	}

	if newSpec.Option == "" {
		newSpec.Option = "Image"
	}

	newResource := &Systemd{
		spec:             newSpec,
		contentRetriever: &text.Text{},
	}

	err = newResource.spec.Validate()
	if err != nil {
		return nil, err
	}

	return newResource, nil
}

// Changelog returns the changelog for this resource, or an empty string if not supported
func (s *Systemd) Changelog(from, to string) *result.Changelogs {
	return nil
}

// ReportConfig returns a new configuration with only the necessary fields
// to identify the resource without any sensitive information
// and context specific data.
func (s *Systemd) ReportConfig() any {
	return Spec{
		File:    s.spec.File,
		Section: s.spec.Section,
		Option:  s.spec.Option,
		Value:   s.spec.Value,
	}
}

// readOptions reads and parses a systemd unit file, returning all options
// and the one matching the configured section/option.
func (s *Systemd) readOptions(filePath string) ([]*unit.UnitOption, *unit.UnitOption, error) {
	if !s.contentRetriever.FileExists(filePath) {
		return nil, nil, fmt.Errorf("the file %s does not exist", filePath)
	}

	content, err := s.contentRetriever.ReadAll(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("reading systemd unit file: %w", err)
	}

	opts, err := unit.DeserializeOptions(strings.NewReader(content))
	if err != nil {
		return nil, nil, fmt.Errorf("parsing systemd unit file: %w", err)
	}

	for _, opt := range opts {
		if opt.Section == s.spec.Section && opt.Name == s.spec.Option {
			return opts, opt, nil
		}
	}

	return nil, nil, fmt.Errorf("option %q not found in section %q", s.spec.Option, s.spec.Section)
}
