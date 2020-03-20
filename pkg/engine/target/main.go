package target

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/olblak/updateCli/pkg/yaml"
)

// Target defines which file need to be updated based on source output
type Target struct {
	Name       string
	Kind       string
	Spec       interface{}
	Repository interface{}
}

// Execute apply a specific target configuration
func (t *Target) Execute(source string) error {

	switch t.Kind {

	case "yaml":
		var spec yaml.Yaml

		err := mapstructure.Decode(t.Spec, &spec)

		if err != nil {
			return err
		}

		err = spec.Update(source)

		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("⚠ Don't support target: %v", t.Kind)
	}

	return nil
}
