package compose

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadFile loads an Updatecli compose file into a compose Spec
func LoadFile(filename string) (*Spec, error) {

	var composeSpec Spec

	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("opening Updatecli compose file %q: %s", filename, err)
	}
	defer f.Close()

	composeFileByte, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("reading Updatecli compose file %q: %s", filename, err)
	}

	err = yaml.Unmarshal(composeFileByte, &composeSpec)
	if err != nil {
		return nil, fmt.Errorf("parsing Updatecli compose file %q: %s", filename, err)
	}

	return &composeSpec, nil
}
