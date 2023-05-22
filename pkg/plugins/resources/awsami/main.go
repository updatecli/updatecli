package awsami

import (
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/mitchellh/mapstructure"
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

// AMI contains information to manipulate AWS AMI information
type AMI struct {
	Spec       Spec
	ec2Filters []*ec2.Filter
	apiClient  ec2iface.EC2API
}

// New returns a reference to a newly initialized AMI object from an AMISpec
// or an error if the provided AMISpec triggers a validation error.
func New(spec interface{}) (*AMI, error) {
	newSpec := Spec{}
	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return nil, err
	}

	errs := newSpec.Validate()
	if len(errs) > 0 {
		logrus.Errorln("failed to validate aws/ami configuration")
		for _, err := range errs {
			logrus.Errorf("%s\n", err.Error())
		}
		return nil, ErrSpecNotValid
	}

	var newFilters []*ec2.Filter
	for i := 0; i < len(newSpec.Filters); i++ {
		filter := ec2.Filter{
			Name:   aws.String(newSpec.Filters[i].Name),
			Values: aws.StringSlice(strings.Split(newSpec.Filters[i].Values, ","))}

		newFilters = append(newFilters, &filter)
	}

	newSession, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	newClient := ec2.New(newSession, &aws.Config{
		CredentialsChainVerboseErrors: func(verbose bool) *bool {
			return &verbose
		}(true),
		Region:   aws.String(newSpec.Region),
		Endpoint: aws.String(newSpec.Endpoint),
		Credentials: credentials.NewChainCredentials(
			[]credentials.Provider{
				&credentials.EnvProvider{},
				&credentials.SharedCredentialsProvider{},
				&credentials.StaticProvider{
					Value: credentials.Value{
						AccessKeyID:     newSpec.AccessKey,
						SecretAccessKey: newSpec.SecretKey,
					},
				},
			}),

		MaxRetries: func(val int) *int { return &val }(3),
	})

	if newClient == nil {
		return nil, ErrWrongServiceConnection
	}

	return &AMI{
		Spec:       newSpec,
		ec2Filters: newFilters,
		apiClient:  newClient,
	}, nil
}

// Changelog returns the changelog for this resource, or an empty string if not supported
func (a *AMI) Changelog() string {
	return ""
}
