package awsami

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
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
		got, _, gotErr := d.ami.Condition("", nil)

		switch d.expectedError == nil {
		case true:
			require.NoError(t, gotErr)
		case false:
			require.ErrorIs(t, d.expectedError, gotErr)
		}

		assert.Equal(t, d.expectedCondition, got)

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
				Images: []types.Image{
					{
						Name:         aws.String("openSUSE-Tumbleweed-v20200626-HVM-x86_64-48127030-1a96-4fef-b318-56ab8588c3c2-ami-0fe97336dfbbcbb07.4"),
						CreationDate: aws.String("2020-06-26"),
						ImageId:      aws.String(imageID),
						Description:  aws.String("openSUSE Tumbleweed (HVM, 64-bit) cabelo@opensuse.org"),
						Architecture: types.ArchitectureValuesX8664,
						Platform:     types.PlatformValuesWindows,
					},
				},
			},
		},
	}

	got, _, gotErr := ami.Condition(imageID, nil)

	require.NoError(t, gotErr)
	assert.Equal(t, true, got)
}
