package helm

import (
	"bytes"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/resources/yaml"
)

// Target updates helm chart, it receives the default source value and a dryrun flag
// then returns if it changed something or failed, along with the list of changed files and the change message
func (c *Chart) Target(source, workingDir string, dryRun bool) (bool, []string, string, error) {
	var out bytes.Buffer

	err := c.ValidateTarget()
	if err != nil {
		return false, []string{}, "", err
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
		return false, []string{}, "", err
	}

	changed, files, message, err := yamlResource.Target(source, workingDir, dryRun)
	if err != nil {
		return false, files, message, err
	} else if err == nil && !changed {
		return false, files, message, nil
	}

	chartPath := filepath.Join(workingDir, c.spec.Name)

	err = c.MetadataUpdate(chartPath, dryRun)
	if err != nil {
		return false, files, message, err
	}

	err = c.RequirementsUpdate(chartPath)
	if err != nil {
		return false, files, message, err
	}

	err = c.DependencyUpdate(&out, chartPath)
	if err != nil {
		return false, files, message, err
	}
	logrus.Infof("%s", out.String())

	files = append(files, chartPath)

	return changed, files, message, err
}
