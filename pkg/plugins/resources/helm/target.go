package helm

import (
	"bytes"
	"fmt"
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
		return fmt.Errorf("unable to update chart %s: %s", c.spec.Name, err)
	}

	chartPath := c.spec.Name
	if scm != nil {
		chartPath = filepath.Join(scm.GetDirectory(), c.spec.Name)
	}

	/*
	  We only want to update the Chart metadata if the chart has been modified during the current target execution.
	  To make this process more idempotent in the context of a scm, we could also check if the helm chart
	  has been modified during one of the previous target execution by comparing the current chart versus the one defined
	  on the source branch. But the code complexity induced by this check is probably not worth the effort today.
	*/
	if resultTarget.Changed {
		err = c.MetadataUpdate(resultTarget.NewInformation, scm, dryRun, resultTarget)
		if err != nil {
			return fmt.Errorf("unable to update chart metadata: %s", err)
		}
	}

	err = c.RequirementsUpdate(chartPath)
	if err != nil {
		return fmt.Errorf("unable to update chart requirements: %s", err)
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
