package helm

import (
	"bytes"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/resources/yaml"
)

// Target updates helm chart, it receives the default source value and a dryrun flag
// then return if it changed something or failed
func (c *Chart) Target(source string, dryRun bool) (changed bool, err error) {
	var out bytes.Buffer

	err = c.ValidateTarget()
	if err != nil {
		return false, err
	}

	yamlSpec := yaml.Spec{
		File: filepath.Join(c.spec.Name, c.spec.File),
		Key:  c.spec.Key,
	}
	if len(c.spec.Value) == 0 {
		c.spec.Value = source
		c.spec.Value = source
	} else {
		yamlSpec.Value = c.spec.Value
	}

	yamlResource, err := yaml.New(yamlSpec)
	if err != nil {
		return false, err
	}

	changed, err = yamlResource.Target(source, dryRun)

	if err != nil {
		return false, err
	} else if err == nil && !changed {
		return false, nil
	}

	// Update Chart.yaml file new Chart Version and appVersion if needed
	err = c.MetadataUpdate(c.spec.Name, dryRun)
	if err != nil {
		return false, err
	}

	err = c.RequirementsUpdate(c.spec.Name)
	if err != nil {
		return false, err
	}

	if !dryRun {
		err = c.DependencyUpdate(&out, c.spec.Name)

		logrus.Debugf("%s", out.String())

		if err != nil {
			return false, err
		}
	}

	return true, nil
}
