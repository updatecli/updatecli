package compose

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"go.yaml.in/yaml/v3"
)

var (
	// DefaultComposeFilename is the default name of the Updatecli compose file
	DefaultComposeFilename = "updatecli-compose.yaml"
	// DeprecatedDefaultComposeFilename is the old default name of the Updatecli compose file
	// cfr for more https://github.com/updatecli/updatecli/issues/2284
	// To be removed in the future
	DeprecatedDefaultComposeFilename = "update-compose.yaml"
)

// GetDefaultComposeFilename is the old default name of the Updatecli compose file
// cfr for more https://github.com/updatecli/updatecli/issues/2284
func GetDefaultComposeFilename() string {
	result := DefaultComposeFilename

	if _, err := os.Stat(DeprecatedDefaultComposeFilename); err == nil {
		logrus.Warnf("Deprecated default compose file %q detected. Please rename it to %q", DeprecatedDefaultComposeFilename, DefaultComposeFilename)
		result = DeprecatedDefaultComposeFilename
	}

	if _, err := os.Stat(DefaultComposeFilename); err == nil {
		if result == DeprecatedDefaultComposeFilename {
			logrus.Warnf("Both default compose files %q and %q detected. Please remove %q to start using %q", DeprecatedDefaultComposeFilename, DefaultComposeFilename, DeprecatedDefaultComposeFilename, DefaultComposeFilename)
		}
	}

	return result
}

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
		if composeSpec.Policies[i].Name == "" {
			composeSpec.Policies[i].Name = fmt.Sprintf("policy-%d - local", i)
			if composeSpec.Policies[i].Policy != "" {
				composeSpec.Policies[i].Name = fmt.Sprintf("policy-%d- %s", i, composeSpec.Policies[i].Policy)
			}
		}
		sanitizePath(composeSpec.Policies[i].Config)
		sanitizePath(composeSpec.Policies[i].Values)
		sanitizePath(composeSpec.Policies[i].Secrets)
	}

	return &composeSpec, nil
}
