package helm

import (
	"bytes"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/resources/yaml"
)

// Target updates helm chart, it receives the default source value and a "dry-run" flag
// then return if it changed something or failed
func (c *Chart) Target(source string, scm scm.ScmHandler, dryRun bool, resultTarget *result.Target) error {
	var out bytes.Buffer
	err := c.ValidateTarget()
	if err != nil {
		return err
	}

	yamlSpec := yaml.Spec{
		File: filepath.Join(c.spec.Name, c.spec.File),
		Key:  c.spec.Key,
	}

	if c.spec.Value != "" {
		yamlSpec.Value = c.spec.Value
	}

	yamlResource, err := yaml.New(yamlSpec)
	if err != nil {
		return err
	}

	err = yamlResource.Target(source, scm, dryRun, resultTarget)

	if err != nil {
		return err
	}

	chartPath := c.spec.Name
	if scm != nil {
		chartPath = filepath.Join(scm.GetDirectory(), c.spec.Name)
	}

	err = c.MetadataUpdate(resultTarget.NewInformation, scm, dryRun, resultTarget)
	if err != nil {
		return err
	}

	err = c.RequirementsUpdate(chartPath)
	if err != nil {
		return err
	}

	if !dryRun {
		err = c.DependencyUpdate(&out, chartPath)

		logrus.Debug(out.String())

		if err != nil {
			return err
		}

	}

	resultTarget.Files = append(resultTarget.Files, chartPath)

	return err
}
