package awsami

import "fmt"

// Filter defines an AMI filter used to narrow down the AMIs returned by the AWS API.
type Filter struct {
	// "name" defines the filter name.
	//
	// example:
	//   * name: architecture
	//
	Name string `yaml:",omitempty"`
	// "values" defines the filter values.
	//
	// remark:
	//   * several values are separated by a comma.
	//
	// example:
	//   * values: x86_64
	//   * values: "x86_64,arm64"
	//
	Values string `yaml:",omitempty"`
}

// Filters represent a list of Filter
type Filters []Filter

func (f *Filters) String() string {
	str := ""
	filters := *f

	for i := 0; i < len(filters); i++ {
		filter := filters[i]
		str = str + fmt.Sprintf("* %s:\t%q\n", filter.Name, filter.Values)

		if i < len(filters)-1 {
			str = str + "\n"
		}

	}

	return str
}
