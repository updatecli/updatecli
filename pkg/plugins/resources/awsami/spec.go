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

/*
"aws/ami" defines the specification for retrieving an AWS AMI ID.
It can be used as a "source" or a "condition".
*/
type Spec struct {
	// "accesskey" defines the AWS access key used to authenticate.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// remark:
	//   * it is only used when "secretkey" is also set.
	//   * when unset, the default AWS credential chain is used.
	//
	AccessKey string `yaml:",omitempty"`
	// "secretkey" defines the AWS secret key used to authenticate.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// remark:
	//   * it is only used when "accesskey" is also set.
	//   * when unset, the default AWS credential chain is used.
	//
	SecretKey string `yaml:",omitempty"`
	// "filters" defines the list of filters used to select AMIs.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// remark:
	//   * at least one filter is required in a source.
	//   * in a condition, when no "image-id" filter is set, the source output is used as the "image-id" filter.
	//   * the accepted filter names are listed on
	//     https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeImages.html
	//
	// example:
	// ```
	//   filters:
	//     - name: "name"
	//       values: "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"
	//     - name: "architecture"
	//       values: "x86_64"
	// ```
	//
	Filters Filters `yaml:",omitempty"`
	// "region" defines the AWS region used to look for AMIs.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// default:
	//   us-east-1
	//
	// example:
	//   * region: eu-west-1
	//
	Region string `yaml:",omitempty"`
	// "endpoint" defines the AWS endpoint used to look for AMIs.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// default:
	//   https://ec2.<region>.amazonaws.com
	//
	Endpoint string `yaml:",omitempty"`
	// "dryrun" checks whether you have the required permissions for the action, without making the request.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// default:
	//   false
	//
	DryRun bool `yaml:",omitempty"`
	// "sortby" defines the order of the AMIs found, the last one being returned.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// remark:
	//   * accepted values are "creationdateasc" and "creationdatedesc", case insensitive.
	//   * when unset, the order returned by the AWS API is kept.
	//
	// example:
	//   * sortby: creationdateasc
	//
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
