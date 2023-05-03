package awsami

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func TestCondition(t *testing.T) {
	// Disable condition testing with running short test
	if testing.Short() {
		return
	}

	for _, d := range dataset {
		d.ami.apiClient = mockDescribeImagesOutput{
			Resp: d.mockedResponse,
		}
		gotResult := result.Condition{}
		err := d.ami.Condition("", nil, &gotResult)

		switch d.expectedError == nil {
		case true:
			require.NoError(t, err)
		case false:
			require.ErrorIs(t, d.expectedError, err)

		}

		assert.Equal(t, d.expectedCondition, gotResult.Pass)

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

	gotResult := result.Condition{}
	err := ami.Condition(imageID, nil, &gotResult)

	require.NoError(t, err)
	assert.Equal(t, true, gotResult.Pass)

}
