package ami

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

// Spec contains updatecli configuration provided by users
type Spec struct {
	AccessKey string // AWs access key
	SecretKey string // AWS secret key
	Filters   Filters
	Region    string
	Endpoint  string
	DryRun    bool
}

// String return Spec information as a string
func (s *Spec) String() (output string) {
	output = output + "Region:\t" + s.Region
	output = output + "\nEndpoint:\t" + s.Endpoint
	output = output + fmt.Sprintf("\nFilters:\n  %s",
		strings.ReplaceAll(s.Filters.String(), "\n", "\n  "))
	return output
}

// Validate ensure that configuration inject are correct
func (s *Spec) Validate() (errs []error) {
	if len(s.Region) == 0 {
		logrus.Printf("No region specified, falling back to %s\n", "us-east-1")
		s.Region = "us-east-1"
	}

	if len(s.Endpoint) == 0 {
		s.Endpoint = fmt.Sprintf("https://ec2.%s.amazonaws.com", s.Region)
	}

	if len(s.Filters) == 0 {
		errs = append(errs, ErrNoFiltersSpecified)
	}
	return errs
}
