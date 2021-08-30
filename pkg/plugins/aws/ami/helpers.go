package ami

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/sirupsen/logrus"
)

// getLatestAmiID query the AWS API to return the newest AMI image id
func (a *AMI) getLatestAmiID(svc ec2iface.EC2API) (string, error) {
	input := ec2.DescribeImagesInput{
		DryRun:  &a.DryRun,
		Filters: a.ec2Filters,
	}

	result, err := svc.DescribeImages(&input)

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
		logrus.Infof("%d AMI found\n", nbImages)

		showShortDescription(result.Images[len(result.Images)-1])

		return *result.Images[len(result.Images)-1].ImageId, nil
	}

	return "", nil
}

// showShortDescription returns a short AMI description as a String.
func showShortDescription(AMI *ec2.Image) string {
	output := ""
	if AMI.Name != nil {
		output = fmt.Sprintf("\tName: %s\n", *AMI.Name)
	}
	if AMI.CreationDate != nil {
		output = output + fmt.Sprintf("\n\tCreation Date: %s\n", *AMI.CreationDate)
	}
	if AMI.Description != nil {
		output = output + fmt.Sprintf("\n\tDescription: %s\n", *AMI.Description)
	}
	if AMI.Architecture != nil {
		output = output + fmt.Sprintf("\n\tArchitecture: %s\n", *AMI.Architecture)
	}
	if AMI.Platform != nil {
		output = output + fmt.Sprintf("\n\tPlatform: %s\n", *AMI.Platform)
	}
	return output
}
