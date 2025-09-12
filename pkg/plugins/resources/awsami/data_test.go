package awsami

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type Data struct {
	ami               AMI
	mockedResponse    ec2.DescribeImagesOutput
	expectedGetAMI    string
	expectedSource    string
	expectedCondition bool
	expectedError     error
}

// MockEC2Client implements EC2ClientAPI for testing
type mockDescribeImagesOutput struct {
	Resp ec2.DescribeImagesOutput
}

func (m mockDescribeImagesOutput) DescribeImages(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error) {
	// Only need to return mocked response output
	return &m.Resp, nil
}

type DataSet []Data

var (
	dataset = DataSet{
		{
			ami: AMI{
				Spec: Spec{
					Filters: Filters{},
				},
			},
			mockedResponse: ec2.DescribeImagesOutput{
				Images: []types.Image{},
			},
			expectedGetAMI:    "",
			expectedSource:    "",
			expectedCondition: false,
			expectedError:     ErrNoFilter,
		},
		{
			ami: AMI{
				Spec: Spec{
					Region: "eu-west-1",
					Filters: Filters{
						{
							Name:   "name",
							Values: "openSUSE-Tumbleweed-v202006*",
						},
					},
				},
			},
			mockedResponse: ec2.DescribeImagesOutput{
				Images: []types.Image{ // Changed from []*ec2.Image
					{
						Name:         aws.String("openSUSE-Tumbleweed-v20200626-HVM-x86_64-48127030-1a96-4fef-b318-56ab8588c3c2-ami-0fe97336dfbbcbb07.4"),
						CreationDate: aws.String("2020-06-26"),
						ImageId:      aws.String("ami-0626a14b9b39e862f"),
						Description:  aws.String("openSUSE Tumbleweed (HVM, 64-bit) cabelo@opensuse.org"),
						Architecture: types.ArchitectureValuesX8664, // Added architecture enum
					},
					{
						Name:         aws.String("openSUSE-Tumbleweed-v20200604-HVM-x86_64-48127030-1a96-4fef-b318-56ab8588c3c2-ami-0ce36c26c006545c9.4"),
						CreationDate: aws.String("2020-06-04"),
						ImageId:      aws.String("ami-08c7016cda7d370a5"),
						Description:  aws.String("openSUSE Tumbleweed (HVM, 64-bit) cabelo@opensuse.org"),
						Architecture: types.ArchitectureValuesX8664,
					},
					{
						Name:         aws.String("openSUSE-Tumbleweed-v20200627-48127030-1a96-4fef-b318-56ab8588c3c2-ami-0941971a046aba5d4.4"),
						CreationDate: aws.String("2020-06-27"),
						ImageId:      aws.String("ami-0a9972d9b4dbdabc7"),
						Description:  aws.String("openSUSE Tumbleweed (HVM, 64-bit) cabelo@opensuse.org"),
						Architecture: types.ArchitectureValuesX8664,
					},
				},
			},
			expectedGetAMI:    "ami-0a9972d9b4dbdabc7",
			expectedSource:    "ami-0a9972d9b4dbdabc7",
			expectedCondition: true,
		},
		{
			ami: AMI{
				Spec: Spec{
					Region: "eu-west-1",
					Filters: Filters{
						{
							Name:   "name",
							Values: "openSUSE-Tumbleweed-v20200627-48127030-1a96-4fef-b318-56ab8588c3c2-ami-0941971a046aba5d4.4",
						},
					},
				},
			},
			mockedResponse: ec2.DescribeImagesOutput{
				Images: []types.Image{ // Changed from []*ec2.Image
					{
						Name:         aws.String("openSUSE-Tumbleweed-v20200627-48127030-1a96-4fef-b318-56ab8588c3c2-ami-0941971a046aba5d4.4"),
						CreationDate: aws.String("2020-06-27"),
						ImageId:      aws.String("ami-0a9972d9b4dbdabc7"),
						Description:  aws.String("openSUSE Tumbleweed (HVM, 64-bit) cabelo@opensuse.org"),
						Architecture: types.ArchitectureValuesX8664,
					},
				},
			},
			expectedGetAMI:    "ami-0a9972d9b4dbdabc7",
			expectedSource:    "ami-0a9972d9b4dbdabc7",
			expectedCondition: true,
		},
		{
			ami: AMI{
				Spec: Spec{
					Filters: Filters{
						{
							Name:   "name",
							Values: "doNotExist",
						},
					},
				},
			},
			mockedResponse: ec2.DescribeImagesOutput{
				Images: []types.Image{},
			},
			expectedGetAMI:    "",
			expectedSource:    "",
			expectedCondition: false,
		},
	}
)
