package registry

import (
	"fmt"
	"io"
	"os"

	"github.com/Masterminds/semver/v3"
	"go.yaml.in/yaml/v3"
)

type PolicySpec struct {
	// Authors is a list of authors of the policy
	Authors []string `yaml:",omitempty"`
	// Changelog is the URL of the policy changelog
	Changelog string `yaml:",omitempty"`
	// Description is the policy description
	Description string `yaml:",omitempty"`
	// Documentation is the URL of the policy documentation
	Documentation string `yaml:",omitempty"`
	// Licenses is the policy license
	Licenses []string `yaml:",omitempty"`
	// Source is the URL of the policy source code
	Source string `yaml:",omitempty"`
	// Vendor is the policy vendor
	Vendor string `yaml:",omitempty"`
	// Version is the policy version, it must be semantic versioning compliant without the leading v
	Version string `yaml:",omitempty"`
	// URL is the URL of the policy source code
	URL string `yaml:",omitempty"`
}

// LoadPolicyFile loads an Updatecli compose file into a compose Spec
func LoadPolicyFile(filename string) (*PolicySpec, error) {

	var policySpec PolicySpec

	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("opening Updatecli policy file %q: %s", filename, err)
	}
	defer f.Close()

	policyFileByte, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("reading Updatecli policy file %q: %s", filename, err)
	}

	if err = yaml.Unmarshal(policyFileByte, &policySpec); err != nil {
		return nil, fmt.Errorf("parsing Updatecli policy file %q: %s", filename, err)
	}

	if err = policySpec.Sanitize(); err != nil {
		return nil, fmt.Errorf("validating Updatecli policy file %q: %s", filename, err)
	}

	return &policySpec, nil
}

// Sanitize validates the policy spec and set default values accordingly
func (s *PolicySpec) Sanitize() error {
	if s.Version == "" {
		s.Version = "0.0.1"
	}

	v, err := semver.NewVersion(s.Version)
	if err != nil {
		return fmt.Errorf("invalid policy version %q: %s", s.Version, err)
	}

	// Trim leading v
	s.Version = v.String()

	return nil
}
