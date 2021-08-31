package ami

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

type Data struct {
	ami               AMI
	resp              ec2.DescribeImagesOutput
	expectedGetAMI    string
	expectedSource    string
	expectedCondition bool
}

type mockDescribeImagesOutput struct {
	ec2iface.EC2API
	Resp ec2.DescribeImagesOutput
}

func (m mockDescribeImagesOutput) DescribeImages(in *ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error) {
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
			resp: ec2.DescribeImagesOutput{
				Images: []*ec2.Image{},
			},
			expectedGetAMI:    "",
			expectedSource:    "",
			expectedCondition: false,
		},
		{
			ami: AMI{
				Spec: Spec{
					Region: "us-east-2",
					Filters: Filters{
						{
							Name:   "name",
							Values: "jenkins-agent-ubuntu*",
						},
					},
				},
			},
			resp: ec2.DescribeImagesOutput{
				Images: []*ec2.Image{
					{
						Name:         aws.String("ubuntu-20.04"),
						CreationDate: aws.String("2020-04-03"),
						ImageId:      aws.String("ami-0ff3b7aefa91e0933"),
						Description:  aws.String("Default jenkins agent based on ubuntu"),
						Platform:     aws.String("linux"),
					},
					{
						Name:         aws.String("ubuntu-20.04"),
						CreationDate: aws.String("2020-04-04"),
						ImageId:      aws.String("ami-0ff3b7aefa91e0934"),
						Description:  aws.String("Default jenkins agent based on ubuntu"),
						Platform:     aws.String("linux"),
					},
					{
						Name:         aws.String("ubuntu-20.04"),
						CreationDate: aws.String("2020-04-04"),
						ImageId:      aws.String("ami-0ff3b7aefa91e0935"),
						Description:  aws.String("Default jenkins agent based on ubuntu"),
						Platform:     aws.String("linux"),
					},
				},
			},
			expectedGetAMI:    "ami-0ff3b7aefa91e0935",
			expectedSource:    "ami-0fbefe596801fce98",
			expectedCondition: true,
		},
		{
			ami: AMI{
				Spec: Spec{
					Filters: Filters{
						{
							Name:   "name",
							Values: "jenkins-agent-ubuntu*",
						},
					},
				},
			},
			resp:              ec2.DescribeImagesOutput{},
			expectedGetAMI:    "",
			expectedSource:    "",
			expectedCondition: false,
		},
		{
			ami: AMI{
				Spec: Spec{
					Region: "us-east-1",
					Filters: Filters{
						{
							Name:   "name",
							Values: "centos*",
						},
					},
				},
			},
			resp: ec2.DescribeImagesOutput{
				Images: []*ec2.Image{
					{
						ImageId: aws.String(""),
					},
				},
			},
			expectedGetAMI:    "",
			expectedSource:    "ami-f2dc849a",
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
			resp: ec2.DescribeImagesOutput{
				Images: []*ec2.Image{},
			},
			expectedGetAMI:    "",
			expectedSource:    "",
			expectedCondition: false,
		},
	}
)
