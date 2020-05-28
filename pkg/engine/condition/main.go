package condition

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/olblak/updateCli/pkg/docker"
	"github.com/olblak/updateCli/pkg/helm/chart"
	"github.com/olblak/updateCli/pkg/maven"
)

// Condition defines which condition needs to be met
// in order to update targets based on the source output
type Condition struct {
	Name string
	Kind string
	Spec interface{}
}

// Spec is an interface to test if condition is met
type Spec interface {
	Condition() (bool, error)
}

// Execute tests if a specific condition is true
func (c *Condition) Execute(source string) (bool, error) {

	var spec Spec

	ok := true

	switch c.Kind {

	case "dockerImage":
		var d docker.Docker

		err := mapstructure.Decode(c.Spec, &d)

		if err != nil {
			return false, err
		}

		d.Tag = source

		spec = &d

	case "maven":
		var m maven.Maven

		err := mapstructure.Decode(c.Spec, &m)

		if err != nil {
			panic(err)
		}

		m.Version = source

		spec = &m

	case "helmChart":
		ch := chart.Chart{}
		err := mapstructure.Decode(c.Spec, &ch)

		if err != nil {
			return false, err
		}

		spec = &ch

	default:
		return false, fmt.Errorf("⚠ Don't support condition: %v", c.Kind)
	}

	ok, err := spec.Condition()

	if err != nil {
		return false, err
	}

	return ok, nil

}
