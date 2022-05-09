package helm

import (
	"bytes"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
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

	err = c.DependencyUpdate(&out, "")

	logrus.Infof("%s", out.String())

	if err != nil {
		return false, err
	}

	return true, nil
}

// TargetFromSCM updates helm chart then push changed to a scm, it receives the default source value and dryrun flag
// then return if it changed something or failed
func (c *Chart) TargetFromSCM(source string, scm scm.ScmHandler, dryRun bool) (
	changed bool, files []string, message string, err error) {

	var out bytes.Buffer
	err = c.ValidateTarget()
	if err != nil {
		return false, files, message, err
	}

	yamlSpec := yaml.Spec{
		File: filepath.Join(c.spec.Name, c.spec.File),
		Key:  c.spec.Key,
	}
	if len(c.spec.Value) == 0 {
		c.spec.Value = source
	} else {
		yamlSpec.Value = c.spec.Value
	}

	yamlResource, err := yaml.New(yamlSpec)
	if err != nil {
		return false, files, message, err
	}

	changed, files, message, err = yamlResource.TargetFromSCM(source, scm, dryRun)
	if err != nil {
		return false, files, message, err
	} else if err == nil && !changed {
		return false, files, message, nil
	}

	chartPath := filepath.Join(scm.GetDirectory(), c.spec.Name)

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

	logrus.Debug(out.String())

	files = append(files, chartPath)

	return changed, files, message, err
}
