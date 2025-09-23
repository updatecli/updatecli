package awsami

import (
	"context"
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/go-viper/mapstructure/v2"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

var (
	// ErrNoFilter is return when didn't narrow AMI ID result
	ErrNoFilter error = errors.New("no filter specified")
	// ErrSpecNotValid is return when aws/ami spec are is not valid
	ErrSpecNotValid error = errors.New("ami spec not valid")
	// ErrWrongServiceConnection is returned when failing to connect to AWS api
	ErrWrongServiceConnection error = errors.New("can't connect to aws api")
)

// EC2ClientAPI defines the interface for EC2 operations we need
type EC2ClientAPI interface {
	DescribeImages(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error)
}

// AMI contains information to manipulate AWS AMI information
type AMI struct {
	Spec       Spec
	ec2Filters []types.Filter
	apiClient  EC2ClientAPI
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

	var newFilters []types.Filter
	for i := 0; i < len(newSpec.Filters); i++ {
		values := strings.Split(newSpec.Filters[i].Values, ",")
		filter := types.Filter{
			Name:   aws.String(newSpec.Filters[i].Name),
			Values: values,
		}
		newFilters = append(newFilters, filter)
	}

	ctx := context.Background()

	var configOptions []func(*config.LoadOptions) error

	if newSpec.Region != "" {
		configOptions = append(configOptions, config.WithRegion(newSpec.Region))
	}

	if newSpec.AccessKey != "" && newSpec.SecretKey != "" {
		staticCredentials := credentials.NewStaticCredentialsProvider(
			newSpec.AccessKey,
			newSpec.SecretKey,
			"",
		)
		configOptions = append(configOptions, config.WithCredentialsProvider(staticCredentials))
	}

	configOptions = append(configOptions, config.WithRetryMaxAttempts(3))

	cfg, err := config.LoadDefaultConfig(ctx, configOptions...)
	if err != nil {
		logrus.Errorf("Failed to load AWS config: %v", err)
		return nil, ErrWrongServiceConnection
	}

	if newSpec.Endpoint != "" {
		cfg.BaseEndpoint = aws.String(newSpec.Endpoint)
	}

	newClient := ec2.NewFromConfig(cfg)

	return &AMI{
		Spec:       newSpec,
		ec2Filters: newFilters,
		apiClient:  newClient,
	}, nil
}

// Changelog returns the changelog for this resource, or an empty string if not supported
func (a *AMI) Changelog(from, to string) *result.Changelogs {
	return nil
}

// ReportConfig returns a new configuration with only the necessary configuration fields
// to identify the resource without any sensitive information or context specific data.
func (a *AMI) ReportConfig() interface{} {
	return Spec{
		Region:   a.Spec.Region,
		Endpoint: a.Spec.Endpoint,
		Filters:  a.Spec.Filters,
	}
}
