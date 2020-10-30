package condition

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/olblak/updateCli/pkg/docker"
	"github.com/olblak/updateCli/pkg/helm/chart"
	"github.com/olblak/updateCli/pkg/maven"
	"github.com/olblak/updateCli/pkg/scm"
	"github.com/olblak/updateCli/pkg/yaml"
)

// Condition defines which condition needs to be met
// in order to update targets based on the source output
type Condition struct {
	Name   string
	Kind   string
	Spec   interface{}
	Scm    map[string]interface{}
	Result string `yaml:"-"` // Ignore this field when unmarshal YAML
}

// Spec is an interface that test if condition is met
type Spec interface {
	Condition(version string) (bool, error)
	ConditionFromSCM(version string, scm scm.Scm) (bool, error)
}

// Run tests if a specific condition is true
func (c *Condition) Run(source string) (ok bool, err error) {

	ok = true

	spec, err := Unmarshal(c)
	if err != nil {
		return false, err
	}

	// If scm is defined then clone the repository
	if len(c.Scm) > 0 {
		s, err := scm.Unmarshal(c.Scm)

		if err != nil {
			return false, err
		}

		err = s.Init(source, c.Name)

		if err != nil {
			return false, err
		}

		ok, err = spec.ConditionFromSCM(source, s)

		if err != nil {
			return false, err
		}

	} else if len(c.Scm) == 0 {

		ok, err = spec.Condition(source)

		if err != nil {
			return false, err
		}
	} else {
		return false, fmt.Errorf("Don't support condition: %v", c.Kind)
	}

	return ok, nil

}

// Unmarshal decodes a condition struct
func Unmarshal(condition *Condition) (spec Spec, err error) {

	switch condition.Kind {

	case "dockerImage":
		d := docker.Docker{}

		err := mapstructure.Decode(condition.Spec, &d)

		if err != nil {
			return nil, err
		}

		spec = &d

	case "maven":
		m := maven.Maven{}

		err := mapstructure.Decode(condition.Spec, &m)

		if err != nil {
			return nil, err
		}

		spec = &m

	case "helmChart":
		ch := chart.Chart{}

		err := mapstructure.Decode(condition.Spec, &ch)

		if err != nil {
			return nil, err
		}

		spec = &ch

	case "yaml":
		y := yaml.Yaml{}

		err := mapstructure.Decode(condition.Spec, &y)

		if err != nil {
			return nil, err
		}

		spec = &y

	default:
		return nil, fmt.Errorf("Don't support condition: %v", condition.Kind)
	}
	return spec, nil

}
