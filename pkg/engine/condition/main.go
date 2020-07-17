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

	pwd, err := os.Getwd()
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

		y.DryRun = true

		err := mapstructure.Decode(c.Spec, &y)

		if err != nil {
			return false, err
		}

		// Means a scm configuration is provided then use the directory from s.GetDirecotry() otherwise try to guess
		if len(c.Scm) > 0 {
			y.Path = workingDir

		} else {
			if dir, base, err := isFileExist(y.File); err == nil && y.Path == "" {
				// if no scm configuration has been provided and neither file path then we try to guess the file directory.
				// if file name contains a path then we use it otherwise we fallback to the current path
				y.Path = dir
				y.File = base
			} else if _, _, err := isFileExist(y.File); err != nil && y.Path == "" {

				y.Path = workingDir

			} else if y.Path != "" && !isDirectory(y.Path) {

				fmt.Printf("Directory '%s' is not valid so fallback to '%s'", y.Path, workingDir)
				y.Path = workingDir

			} else {
				return false, fmt.Errorf("Something weird happened while trying to set working directory")
			}

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

func isFileExist(file string) (dir string, base string, err error) {
	if _, err := os.Stat(file); err != nil {
		return "", "", err
	}

	absolutePath, err := filepath.Abs(file)
	if err != nil {
		return "", "", err
	}
	dir = filepath.Dir(absolutePath)
	base = filepath.Base(absolutePath)

	return dir, base, err
}

func isDirectory(path string) bool {

	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	if info.IsDir() {
		return true
	}
	return false
}
