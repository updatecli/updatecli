package ami

import (
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/sirupsen/logrus"
)

var (
	// ErrNoFiltersSpecified is return when didn't narrow AMI ID result
	ErrNoFiltersSpecified error = errors.New("Error - no filters specified")
)

// https://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#EC2.DescribeImages

// AMI contains information to manipuliate AWS AMI information
type AMI struct {
	Spec       Spec
	ec2Filters []*ec2.Filter
}

// Init run basic parameter initiation
func (a *AMI) Init() (svc *ec2.EC2, errs []error) {

	errs = a.Spec.Validate()

	if len(errs) > 0 {
		for _, err := range errs {
			logrus.Printf("%s\n", err.Error())
		}
		return nil, errs
	}

	// Init ec2Filters
	for i := 0; i < len(a.Spec.Filters); i++ {
		filter := ec2.Filter{
			Name:   aws.String(a.Spec.Filters[i].Name),
			Values: aws.StringSlice(strings.Split(a.Spec.Filters[i].Values, ","))}

		a.ec2Filters = append(a.ec2Filters, &filter)
	}

	svc = ec2.New(session.New(), &aws.Config{
		CredentialsChainVerboseErrors: func(verbose bool) *bool {
			return &verbose
		}(true),
		Region:   aws.String(a.Spec.Region),
		Endpoint: aws.String(a.Spec.Endpoint),
		Credentials: credentials.NewChainCredentials(
			[]credentials.Provider{
				&credentials.EnvProvider{},
				&credentials.SharedCredentialsProvider{},
				&credentials.StaticProvider{
					Value: credentials.Value{
						AccessKeyID:     a.Spec.AccessKey,
						SecretAccessKey: a.Spec.SecretKey,
					},
				},
			}),

		MaxRetries: func(val int) *int { return &val }(3),
	})

	return svc, errs
}
