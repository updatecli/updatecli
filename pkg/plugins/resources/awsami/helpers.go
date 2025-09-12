package awsami

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"
	"github.com/sirupsen/logrus"
)

// getLatestAmiID queries the AWS API to return the newest AMI image id.
func (a *AMI) getLatestAmiID() (string, error) {
	input := ec2.DescribeImagesInput{
		DryRun:  &a.Spec.DryRun,
		Filters: a.ec2Filters,
	}

	result, err := a.apiClient.DescribeImages(context.TODO(), &input)

	if err != nil {
		var ae smithy.APIError
		if errors.As(err, &ae) {
			logrus.Errorf("AWS API Error - Code: %s, Message: %s", ae.ErrorCode(), ae.ErrorMessage())
		} else {
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
func showShortDescription(AMI types.Image) string {
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
	if len(AMI.Architecture) > 0 {
		output = output + fmt.Sprintf("* architecture: %s\n", string(AMI.Architecture))
	}
	if len(AMI.Platform) > 0 {
		output = output + fmt.Sprintf("* Platform: %s\n", string(AMI.Platform))
	}
	return output
}
