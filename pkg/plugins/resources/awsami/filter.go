package awsami

import "fmt"

// Filter represents the updatecli configuration describing AMI filters.
type Filter struct {
	// Name specifies a filter name.
	Name string `yaml:",omitempty"`
	// Values specifies a filter value for a specific filter name.
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
