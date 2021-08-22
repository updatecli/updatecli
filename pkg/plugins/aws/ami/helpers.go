package ami

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
)

// ShowShortDescription print a short AMI description.
func ShowShortDescription(AMI *ec2.Image) string {
	output := ""
	if AMI.Name != nil {
		output = fmt.Sprintf("\tName: %s\n", *AMI.Name)
	}
	if AMI.CreationDate != nil {
		output = output + fmt.Sprintf("\n\tCreation Date: %s\n", *AMI.CreationDate)
	}
	if AMI.Description != nil {
		output = output + fmt.Sprintf("\n\tDescription: %s\n", *AMI.Description)
	}
	if AMI.Architecture != nil {
		output = output + fmt.Sprintf("\n\tArchitecture: %s\n", *AMI.Architecture)
	}
	if AMI.Platform != nil {
		output = output + fmt.Sprintf("\n\tPlatform: %s\n", *AMI.Platform)
	}
	return output
}
