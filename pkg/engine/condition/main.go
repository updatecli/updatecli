package condition

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/olblak/updateCli/pkg/docker"
	"github.com/olblak/updateCli/pkg/maven"
)

// Condition defines which condition needs to be met
// in order to update targets based on the source output
type Condition struct {
	Name string
	Kind string
	Spec interface{}
}

// Execute tests if a specific condition is true
func (c *Condition) Execute(source string) (bool, error) {

	ok := true

	switch c.Kind {

	case "dockerImage":
		var d docker.Docker

		err := mapstructure.Decode(c.Spec, &d)

		if err != nil {
			return false, err
		}

		d.Tag = source

		ok = d.IsTagPublished()

		fmt.Printf("\n")

	case "maven":
		var m maven.Maven

		err := mapstructure.Decode(c.Spec, &m)

		if err != nil {
			panic(err)
		}

		m.Version = source

		ok = m.IsTagPublished()
		fmt.Printf("\n")

	default:
		return false, fmt.Errorf("⚠ Don't support condition: %v", c.Kind)
	}

	return ok, nil

}
