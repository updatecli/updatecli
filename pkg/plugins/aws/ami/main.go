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
}

// Init run basic parameter initiation
func (a *AMI) Init() (svc *ec2.EC2, err error) {

	errs := a.Spec.Validate()

	if len(errs) > 0 {
		logrus.Errorln("failed to validate aws/ami configuration")
		for _, err := range errs {
			logrus.Errorf("%s\n", err.Error())
		}
		return nil, ErrSpecNotValid
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

	return svc, err
}
