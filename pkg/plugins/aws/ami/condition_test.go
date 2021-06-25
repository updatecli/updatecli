package ami

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type Data struct {
	ami         AMI
	expectedAMI string
}

type DataSet []Data

var (
	dataset = DataSet{
		{
			ami: AMI{
				Filters: ec2.DescribeImagesInput{
					ImageIds: []*string{
						aws.String("ami-5731123e"),
					},
					Filters: []*ec2.Filter{
						{
							Name: aws.String("architecture"),
							Values: []*string{
								aws.String("arm64"),
							},
						},
					},
				},
			},
		},
	}
)

func TestSource(t *testing.T) {

	for _, d := range dataset {
		got, err := d.ami.Source("")
		if err != nil {
			t.Errorf("Unexpected error: %q", err)
		}
		if strings.Compare(got, d.expectedAMI) != 0 {
			t.Errorf("ReExpected Result %q, got %q", d.expectedAMI, got)

		}
	}
}
