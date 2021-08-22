package ami

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/sirupsen/logrus"
)

// Source returns the latest AMI matching the filter
func (a *AMI) Source(workingDir string) (string, error) {

	errs := a.Init()

	for _, err := range errs {
		logrus.Error(err)
	}
	if len(errs) > 0 {
		return "", errors.New("Too many errors")
	}

	svc := ec2.New(session.New(), &aws.Config{
		CredentialsChainVerboseErrors: func(verbose bool) *bool {
			return &verbose
		}(true),
		Region:      aws.String(a.Region),
		Endpoint:    aws.String(a.Endpoint),
		Credentials: a.credentials,
		MaxRetries:  func(val int) *int { return &val }(3),
	})

	input := &ec2.DescribeImagesInput{
		DryRun:  &a.DryRun,
		Filters: a.ec2Filters,
	}

	result, err := svc.DescribeImages(input)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				logrus.Errorln(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			logrus.Errorln(err.Error())
		}
		return "", err
	}

	if nbImages := len(result.Images); nbImages > 0 {
		logrus.Infof("\u2714 %d AMI found\n", nbImages)

		ShowShortDescription(result.Images[len(result.Images)-1])

		return *result.Images[len(result.Images)-1].ImageId, nil
	}

	fmt.Printf("\u2717 No AMI found matching criteria\n")

	return "", nil
}
