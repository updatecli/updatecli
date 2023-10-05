package registry

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type PolicySpec struct {
	// Authors is a list of authors of the policy
	Authors []string `yaml:",omitempty"`
	// URL is the URL of the policy source code
	URL string `yaml:",omitempty"`
	// Documentation is the URL of the policy documentation
	Documentation string `yaml:",omitempty"`
	// Source is the URL of the policy source code
	Source string `yaml:",omitempty"`
	// Version is the policy version
	Version string `yaml:",omitempty"`
	// Vendor is the policy vendor
	Vendor string `yaml:",omitempty"`
	// Licenses is the policy license
	Licenses []string `yaml:",omitempty"`
	// I don't understand why if set, it creates a file locally named with the value
	// To investigate...
	// Disalbed for now
	// Title string `yaml:",omitempty"`
	// Description is the policy description
	Description string `yaml:",omitempty"`
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

	err = yaml.Unmarshal(policyFileByte, &policySpec)
	if err != nil {
		return nil, fmt.Errorf("parsing Updatecli policy file %q: %s", filename, err)
	}

	return &policySpec, nil
}
