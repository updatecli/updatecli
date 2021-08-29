package ami

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// https://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#EC2.DescribeImages

// AMI contains information to manipuliate AWS AMI information
type AMI struct {
	AccessKey  string // AWs access key
	SecretKey  string // AWS secret key
	Filters    Filters
	Region     string
	Endpoint   string
	DryRun     bool
	ec2Filters []*ec2.Filter
}

// Filter represents the updatecli configuration to define AMI filter
// This datatype need to be convert the ec2.Filter.
type Filter struct {
	Name   string
	Values string
}

// Filters represent a list of Filter
type Filters []Filter

func (f *Filters) String() string {
	str := ""
	filters := *f

	for i := 0; i < len(filters); i++ {
		filter := filters[i]
		str = str + fmt.Sprintf("%s: \t%q", filter.Name, filter.Values)

		if i < len(filters)-1 {
			str = str + "\n"
		}

	}

	return str
}

// Init run basic parameter initiation
func (a *AMI) Init() (svc *ec2.EC2, errs []error) {
	if len(a.Region) == 0 {
		a.Region = "us-east-1"
	}

	if len(a.Endpoint) == 0 {
		a.Endpoint = fmt.Sprintf("https://ec2.%s.amazonaws.com", a.Region)
	}

	// Convert []string to []*string as required by the ec2.Filter values field
	values := func(input []string) []*string {
		var output []*string
		for i := range input {
			s := input[i]
			output = append(output, &s)
		}
		return output
	}

	// Init ec2Filters
	for i := 0; i < len(a.Filters); i++ {
		filter := ec2.Filter{
			Name: func(input string) *string {
				output := strings.ToLower(input)
				return &output
			}(a.Filters[i].Name),
			Values: values(strings.Split(a.Filters[i].Values, ",")),
		}
		a.ec2Filters = append(a.ec2Filters, &filter)
	}

	svc = ec2.New(session.New(), &aws.Config{
		CredentialsChainVerboseErrors: func(verbose bool) *bool {
			return &verbose
		}(true),
		Region:   aws.String(a.Region),
		Endpoint: aws.String(a.Endpoint),
		Credentials: credentials.NewChainCredentials(
			[]credentials.Provider{
				&credentials.EnvProvider{},
				&credentials.SharedCredentialsProvider{},
				&credentials.StaticProvider{
					Value: credentials.Value{
						AccessKeyID:     a.AccessKey,
						SecretAccessKey: a.SecretKey,
					},
				},
			}),

		MaxRetries: func(val int) *int { return &val }(3),
	})

	return svc, errs
}
