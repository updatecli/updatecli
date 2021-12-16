package awsami

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func TestCondition(t *testing.T) {
	// Disable condition testing with running short test
	if testing.Short() {
		return
	}

	for id, d := range dataset {
		d.ami.apiClient = mockDescribeImagesOutput{
			Resp: d.mockedResponse,
		}
		got, err := d.ami.Condition("")

		if !errors.Is(err, d.expectedError) {
			t.Errorf("[%d] Wrong error:\nExpected Error:\t%v\nGot:\t\t%v\n",
				id, d.expectedError, err)
		}

		if got != d.expectedCondition {
			t.Errorf("[%d] Wrong AMI conditional result:\nExpected Result:\t\t%v\nGot:\t\t\t\t\t%v",
				id,
				d.expectedCondition,
				got)
		}
	}

	// Test inject image-id
	imageID := "ami-0a9972d9b4dbdabc7"

	ami := AMI{
		Spec: Spec{
			Region:  "eu-west-1",
			Filters: Filters{},
		},
		apiClient: mockDescribeImagesOutput{
			Resp: ec2.DescribeImagesOutput{
				Images: []*ec2.Image{
					{
						Name:         aws.String("openSUSE-Tumbleweed-v20200626-HVM-x86_64-48127030-1a96-4fef-b318-56ab8588c3c2-ami-0fe97336dfbbcbb07.4"),
						CreationDate: aws.String("2020-06-26"),
						ImageId:      aws.String(imageID),
						Description:  aws.String("openSUSE Tumbleweed (HVM, 64-bit) cabelo@opensuse.org"),
					},
				},
			},
		},
	}

	exist, err := ami.Condition(imageID)
	if err != nil {
		t.Errorf("Unexpected error: %q",
			err)
	}

	if !exist {
		t.Errorf("[%s] Wrong AMI conditional result:\nExpected Result:\t\t%v\nGot:\t\t\t\t\t%v",
			imageID,
			true,
			exist)
	}
}
