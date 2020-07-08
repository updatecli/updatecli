package condition

import (
	"fmt"
	"os"
	"path/filepath"

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
	Name string
	Kind string
	Spec interface{}
	Scm  map[string]interface{}
}

// Spec is an interface that test if condition is met
type Spec interface {
	Condition() (bool, error)
}

// Execute tests if a specific condition is true
func (c *Condition) Execute(source string) (bool, error) {

	var s scm.Scm

	pwd, err := os.Executable()
	if err != nil {
		panic(err)
	}

	// By default workingDir is set to local directory
	workingDir := filepath.Dir(pwd)

	// If scm is defined then clone the repository
	if len(c.Scm) > 0 {
		s, err = scm.Unmarshal(c.Scm)
		if err != nil {
			return false, err
		}

		err = s.Init(source, c.Name)

		defer s.Clean()

		if err != nil {
			return false, err
		}

		s.Clone()

		workingDir = s.GetDirectory()
	}

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

	case "yaml":
		var y yaml.Yaml

		err := mapstructure.Decode(c.Spec, &y)

		if err != nil {
			return false, err
		}

		y.Path = workingDir

		spec = &y

	default:
		return false, fmt.Errorf("Don't support condition: %v", c.Kind)
	}

	ok, err = spec.Condition()

	if err != nil {
		return false, err
	}

	return ok, nil

}
