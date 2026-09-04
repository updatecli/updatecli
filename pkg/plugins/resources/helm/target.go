package helm

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/resources/yaml"
	"github.com/updatecli/updatecli/pkg/plugins/utils"
)

// Target updates helm chart, it receives the default source value and a "dry-run" flag
// then return if it changed something or failed
func (c *Chart) Target(ctx context.Context, source string, scm scm.ScmHandler, resolver utils.Resolver, dryRun bool, resultTarget *result.Target) error {
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

	err = yamlResource.Target(ctx, source, scm, resolver, dryRun, resultTarget)
	if err != nil {
		return fmt.Errorf("unable to update chart %s: %s", c.spec.Name, err)
	}

	chartPath := resolver.Join(c.spec.Name)

	err = c.MetadataUpdate(ctx, resultTarget.NewInformation, scm, resolver, dryRun, resultTarget)
	if err != nil {
		return fmt.Errorf("unable to update chart metadata: %s", err)
	}

	err = c.RequirementsUpdate(chartPath)
	if err != nil {
		return fmt.Errorf("unable to update chart requirements: %s", err)
	}

	if !dryRun && !c.spec.SkipPackaging {
		err = c.DependencyUpdate(&out, chartPath)

		logrus.Debug(out.String())

		if err != nil {
			return err
		}

	}

	resultTarget.Files = append(resultTarget.Files, chartPath)

	return err
}
