package ami

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/sirupsen/logrus"
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

// Filter represents the updatecli configuration which describes AMI filters
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
		logrus.Println("No region specified, falling back to %s", "us-east-1")
		a.Region = "us-east-1"
	}

	if len(a.Endpoint) == 0 {
		a.Endpoint = fmt.Sprintf("https://ec2.%s.amazonaws.com", a.Region)
	}

	// Init ec2Filters
	for i := 0; i < len(a.Filters); i++ {
		filter := ec2.Filter{
			Name:   aws.String(a.Filters[i].Name),
			Values: aws.StringSlice(strings.Split(a.Filters[i].Values, ","))}

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
