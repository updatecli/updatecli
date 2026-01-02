package scaffold

import (
	_ "embed"

	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/sirupsen/logrus"
)

var (
	//go:embed assets/Policy.yaml.tmpl
	policyTemplate string

	defaultAuthors       []string = []string{"Please insert an author for your policy"}
	defaultDocumentation string   = "Please insert a documentation url for your policy"
	defaultURL           string   = "Please insert a url for your policy"
	defaultDescription   string   = "Please insert a description for your policy"
	defaultLicense       string   = "Please insert a license for your policy"
	defaultSource        string   = "Please insert the source url of the policy"
	defaultVendor        string   = "Please insert a vendor for your policy"
	defaultVersion       string   = "0.1.0"
)

// PolicySpec is the policy specification
type PolicySpec struct {
	// Authors is the policy authors
	Authors []string
	// Description is the policy description
	Description string
	// Documentation is the policy documentation URL
	Documentation string
	// License is the policy license
	License string
	// Source is the policy source URL
	Source string
	// Version is the policy version
	Version string
	// Vendor is the policy vendor
	Vendor string
	// URL is the policy url
	URL string
}

// sanitize set default values for the policy specification
func (p *PolicySpec) sanitize() {

	setDefaultValues := func(s *string, defaultValue string) {
		if *s == "" {
			*s = defaultValue
		}
	}

	setDefaultArrayValues := func(s *[]string, defaultValue []string) {
		if len(*s) == 0 {
			*s = defaultValue
		}
	}

	setDefaultValues(&p.Description, defaultDescription)
	setDefaultValues(&p.Documentation, defaultDocumentation)
	setDefaultValues(&p.License, defaultLicense)
	setDefaultValues(&p.Source, defaultSource)
	setDefaultValues(&p.Vendor, defaultVendor)
	setDefaultValues(&p.Version, defaultVersion)
	setDefaultValues(&p.URL, defaultURL)

	setDefaultArrayValues(&p.Authors, defaultAuthors)

}

// scaffoldPolicy scaffold a new Updatecli policy file
func (s *Scaffold) scaffoldPolicy(p *PolicySpec, dirname string) error {

	p.sanitize()

	if _, err := os.Stat(dirname); os.IsNotExist(err) {
		err := os.MkdirAll(dirname, 0755)
		if err != nil {
			return err
		}
	}

	policyFilePath := filepath.Join(dirname, s.PolicyFile)

	if _, err := os.Stat(policyFilePath); err == nil {
		logrus.Infof("Skipping, policy already exist: %s", policyFilePath)
		return nil
	}

	f, err := os.Create(policyFilePath)
	if err != nil {
		return err
	}

	defer f.Close()

	tmpl, err := template.New("policy").Parse(policyTemplate)
	if err != nil {
		return fmt.Errorf("unable to parse policy template: %s", err)
	}

	err = tmpl.Execute(f, p)
	if err != nil {
		return err
	}

	return nil
}
