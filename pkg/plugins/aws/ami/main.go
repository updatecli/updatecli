package ami

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// https://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#EC2.DescribeImages

// AMI contains information to manipuliate AWS AMI information
type AMI struct {
	AccessKey   string // AWs access key
	SecretKey   string // AWS secret key
	Filters     Filters
	Region      string
	Endpoint    string
	DryRun      bool
	ec2Filters  []*ec2.Filter
	credentials *credentials.Credentials
}

// Filters represents all filter type that can be used to identify a specific AMI
type Filters struct {
	Architecture       string // architecture        - The image architecture (i386 | x86_64 | arm64).
	ImageID            string // image-id            - The ID of the image.
	ImageType          string // image-type          - The image type (machine | kernel | ramdisk).
	IsPublic           bool   // is-public           - A Boolean that indicates whether the image is public.
	HyperVisor         string // hypervisor          - The hypervisor type (ovm | xen).
	Name               string // name 	             - The name of the AMI (provided during image creation).
	OwnerID            string // owner-id            - The AWS account ID of the owner. We recommend that you use. the Owner request parameter instead of this filter.
	VirtualizationType string // virtualization-type - The virtualization type (paravirtual | hvm).
}

func (f *Filters) String() string {
	output := ""
	output = fmt.Sprintf("Architecture: \t%q", f.Architecture)
	output = output + fmt.Sprintf("/\nArchitecture: \t%q", f.Architecture)
	output = output + fmt.Sprintf("/\nImageID: \t%q", f.ImageID)
	output = output + fmt.Sprintf("/\nImageType: \t%q", f.ImageType)
	output = output + fmt.Sprintf("/\nName: \t%q", f.Name)
	return output
}

// Init run basic parameter initiation
func (a *AMI) Init() (errs []error) {
	if len(a.Region) == 0 {
		a.Region = "us-east-1"
	}

	if len(a.Endpoint) == 0 {
		a.Endpoint = fmt.Sprintf("https://ec2.%s.amazonaws.com", a.Region)
	}

	// Convert []string to []*string as required by the ec2.Filter values field
	values := func(input []string) []*string {
		var output []*string
		for _, s := range input {
			output = append(output, &s)
		}
		return output
	}

	if len(a.Filters.Architecture) > 0 {
		name := "architecture"
		filter := ec2.Filter{
			Name:   &name,
			Values: values(strings.Split(a.Filters.Architecture, ",")),
		}
		a.ec2Filters = append(a.ec2Filters, &filter)
	}

	if len(a.Filters.Name) > 0 {
		name := "name"
		filter := ec2.Filter{
			Name:   &name,
			Values: values(strings.Split(a.Filters.Name, ",")),
		}
		a.ec2Filters = append(a.ec2Filters, &filter)
	}

	if len(a.Filters.ImageID) > 0 {
		name := "image-id"
		filter := ec2.Filter{
			Name:   &name,
			Values: values(strings.Split(a.Filters.ImageID, ",")),
		}
		a.ec2Filters = append(a.ec2Filters, &filter)
	}

	a.credentials = credentials.NewChainCredentials(
		[]credentials.Provider{
			&credentials.EnvProvider{},
			&credentials.SharedCredentialsProvider{},
			&credentials.StaticProvider{
				Value: credentials.Value{
					AccessKeyID:     a.AccessKey,
					SecretAccessKey: a.SecretKey,
				},
			},
		})

	return errs
}
