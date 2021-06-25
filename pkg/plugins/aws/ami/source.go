package ami

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/sirupsen/logrus"
)

// Source returns the latest AMI matching the filter
func (a *AMI) Source(workingDir string) (string, error) {
	//

	svc := ec2.New(session.New(), &aws.Config{
		Region:      aws.String("us-east-1"),
		Endpoint:    aws.String("https://ec2.us-,east-1.amazonaws.com"),
		Credentials: aws.String("xxx"),
	})

	//input := &ec2.DescribeImagesInput{
	//	ImageIds: []*string{
	//		aws.String("ami-5731123e"),
	//	},
	//}

	//result, err := svc.DescribeImages(input)

	result, err := svc.DescribeImages(&a.Filters)
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

	fmt.Println(result)

	return result.String(), nil
}
