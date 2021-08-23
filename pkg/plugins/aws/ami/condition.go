package ami

import (
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/scm"
)

// Condition test if a image matching specific filters exist.
func (a *AMI) Condition(source string) (bool, error) {

	if source != "" {
		// Based on source information,
		// we try to define a default image-id resource
		// if not researched
		isImageIDDefined := false
		for i := 0; i < len(a.Filters); i++ {
			if strings.Compare(a.Filters[i].Name, "image-id") == 0 {
				isImageIDDefined = true
				break
			}

		}

		// Set image-id to source output if not yet defined
		if !isImageIDDefined {
			a.Filters = append(a.Filters, Filter{
				Name:   "image-id",
				Values: source,
			})
		}

	}

	errs := a.Init()

	for _, err := range errs {
		logrus.Error(err)
	}
	if len(errs) > 0 {
		return false, errors.New("Too many errors")
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
		return false, err
	}

	if nbImages := len(result.Images); nbImages > 0 {
		logrus.Infof("\u2714 %d AMI found\n", nbImages)

		ShowShortDescription(result.Images[len(result.Images)-1])

		return true, nil
	}

	fmt.Printf("\u2717 No AMI found matching criteria\n")

	return false, nil
}

// ConditionFromSCM is a placeholder to validate the condition interface
func (a *AMI) ConditionFromSCM(source string, scm scm.Scm) (bool, error) {

	fmt.Printf("\u2717 Condition with SCM is not supported, please remove the scm block \n")

	return false, errors.New("condition with SCM is not supported")
}
