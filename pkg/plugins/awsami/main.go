package awsami

import (
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/sirupsen/logrus"
)

var (
	// ErrNoFilter is return when didn't narrow AMI ID result
	ErrNoFilter error = errors.New("no filter specified")
	// ErrSpecNotValid is return when aws/ami spec are is not valid
	ErrSpecNotValid error = errors.New("ami spec not valid")
	// ErrWrongServiceConnection is returned when failing to connect to AWS api
	ErrWrongServiceConnection error = errors.New("can't connect to aws api")
)

// https://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#EC2.DescribeImages

// AMI contains information to manipuliate AWS AMI information
type AMI struct {
	Spec       Spec
	ec2Filters []*ec2.Filter
	apiClient  ec2iface.EC2API
}

// New returns a reference to a newly initialized AMI object from an AMISpec
// or an error if the provided AMISpec triggers a validation error.
func New(spec Spec) (*AMI, error) {
	errs := spec.Validate()
	if len(errs) > 0 {
		logrus.Errorln("failed to validate aws/ami configuration")
		for _, err := range errs {
			logrus.Errorf("%s\n", err.Error())
		}
		return nil, ErrSpecNotValid
	}

	var newFilters []*ec2.Filter
	for i := 0; i < len(spec.Filters); i++ {
		filter := ec2.Filter{
			Name:   aws.String(spec.Filters[i].Name),
			Values: aws.StringSlice(strings.Split(spec.Filters[i].Values, ","))}

		newFilters = append(newFilters, &filter)
	}

	newClient := ec2.New(session.New(), &aws.Config{
		CredentialsChainVerboseErrors: func(verbose bool) *bool {
			return &verbose
		}(true),
		Region:   aws.String(spec.Region),
		Endpoint: aws.String(spec.Endpoint),
		Credentials: credentials.NewChainCredentials(
			[]credentials.Provider{
				&credentials.EnvProvider{},
				&credentials.SharedCredentialsProvider{},
				&credentials.StaticProvider{
					Value: credentials.Value{
						AccessKeyID:     spec.AccessKey,
						SecretAccessKey: spec.SecretKey,
					},
				},
			}),

		MaxRetries: func(val int) *int { return &val }(3),
	})

	if newClient == nil {
		return nil, ErrWrongServiceConnection
	}

	return &AMI{
		Spec:       spec,
		ec2Filters: newFilters,
		apiClient:  newClient,
	}, nil
}
