package awsami

import (
	"fmt"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/sirupsen/logrus"
)

// getLatestAmiID queries the AWS API to return the newest AMI image id.
func (a *AMI) getLatestAmiID() (string, error) {
	input := ec2.DescribeImagesInput{
		DryRun:  &a.Spec.DryRun,
		Filters: a.ec2Filters,
	}

	result, err := a.apiClient.DescribeImages(&input)

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

		switch a.Spec.SortBy {
		case "creationdateasc":
			sort.Sort(ByCreationDateAsc(result.Images))
		case "creationdatedesc":
			sort.Sort(ByCreationDateDesc(result.Images))
		}

		logrus.Debugf("Latest AMI ID found:\n  ---\n  %s---\n\n",
			strings.ReplaceAll(
				showShortDescription(result.Images[len(result.Images)-1]),
				"\n",
				"\n  "))

		return *result.Images[len(result.Images)-1].ImageId, nil
	}

	return "", nil
}

// showShortDescription returns a short AMI description as a String.
func showShortDescription(AMI *ec2.Image) string {
	output := ""
	if AMI.Name != nil {
		output = fmt.Sprintf("* name: %s\n", *AMI.Name)
	}
	if AMI.CreationDate != nil {
		output = output + fmt.Sprintf("* creation date: %s\n", *AMI.CreationDate)
	}
	if AMI.Description != nil {
		output = output + fmt.Sprintf("* description: %s\n", *AMI.Description)
	}
	if AMI.Architecture != nil {
		output = output + fmt.Sprintf("* architecture: %s\n", *AMI.Architecture)
	}
	if AMI.Platform != nil {
		output = output + fmt.Sprintf("* Platform: %s\n", *AMI.Platform)
	}
	return output
}
