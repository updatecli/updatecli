package compose

import (
	"fmt"
	"io"
	"os"

	"path/filepath"

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

	// Ensure that any relative file path is relative to the compose file
	sanitizePath := func(path []string) {
		for i := range path {
			if !filepath.IsAbs(path[i]) {
				path[i] = filepath.Join(filepath.Dir(filename), path[i])
			}
		}
	}

	sanitizePath(composeSpec.Env_files)

	for i := range composeSpec.Policies {
		sanitizePath(composeSpec.Policies[i].Config)
		sanitizePath(composeSpec.Policies[i].Values)
		sanitizePath(composeSpec.Policies[i].Secrets)
	}

	return &composeSpec, nil
}
