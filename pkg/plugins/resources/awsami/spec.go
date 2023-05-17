package awsami

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

var (
	// ErrWrongSortByValue is returned if we use the wrong sortBy value
	ErrWrongSortByValue error = errors.New("wrong value for key 'sortBy'")
)

// Spec contains the updatecli configuration provided by users.
type Spec struct {
	// accesskey specifies the aws access key which combined with `secretkey`, is one of the way to authenticate
	AccessKey string `yaml:",omitempty"`
	// secretkey specifies the aws secret key which combined with `accesskey`, is one of the way to authenticate
	SecretKey string `yaml:",omitempty"`
	// Filters specifies a list of AMI filters
	Filters Filters `yaml:",omitempty"`
	// Region specifies the AWS region to use when looking for AMI
	Region string `yaml:",omitempty"`
	// Endpoint specifies the AWS endpoint to use when looking for AMI
	Endpoint string `yaml:",omitempty"`
	// Dryrun allows to Check whether you have the required permissions for the action.
	DryRun bool `yaml:",omitempty"`
	// Sortby specifies the order of AMI-ID that will be used to retrieve the last element such as `creationdateasc`
	SortBy string `yaml:",omitempty"`
}

// String return Spec information as a string
func (s *Spec) String() (output string) {
	output = output + "Region:\t" + s.Region
	output = output + "\nEndpoint:\t" + s.Endpoint
	output = output + fmt.Sprintf("\nFilters:\n  %s",
		strings.ReplaceAll(s.Filters.String(), "\n", "\n  "))
	return output
}

func getSortByAcceptedValues() []string {
	return []string{
		"creationdateasc",
		"creationdatedesc",
	}
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

	if len(s.SortBy) > 0 {
		found := false
		for _, acceptedValue := range getSortByAcceptedValues() {
			if strings.Compare(strings.ToLower(s.SortBy), strings.ToLower(acceptedValue)) == 0 {
				found = true
				// Ensure we use lowercase,
				s.SortBy = strings.ToLower(s.SortBy)
				break
			}
		}
		if !found {
			logrus.Printf("Invalid sortBy value %q", s.SortBy)
			logrus.Printf("Accepted values: %v", getSortByAcceptedValues())
			errs = append(errs, ErrWrongSortByValue)
		}
	}
	return errs
}
