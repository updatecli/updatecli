package compose

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/engine/manifest"
	"github.com/updatecli/updatecli/pkg/core/registry"
	"go.yaml.in/yaml/v3"
)

// Compose is a struct that contains a compose object
type Compose struct {
	// spec contains the compose spec
	spec     Spec
	filename string
	name     string
}

// New creates a new Compose object
func New(filename string, visitedComposeFiles map[string]bool) ([]Compose, error) {

	composeFiles := []Compose{}

	if _, ok := visitedComposeFiles[filename]; ok {
		logrus.Debugf("Compose file %q already loaded, skipping to avoid circular reference", filename)
		return nil, nil
	}

	visitedComposeFiles[filename] = true

	logrus.Infof("Loading compose file: %q", filename)

	spec, err := LoadFile(filename)
	if err != nil {
		return nil, err
	}

	if len(spec.Name) != 0 {
		logrus.Infof("\tName: %q", spec.Name)
	}

	switch len(spec.Policies) {
	case 0:
		//
	case 1:
		logrus.Infof("One policy detected:\n\t* Policy: %s", spec.Policies[0].Name)
	default:
		logrus.Infof("%d policies detected:", len(spec.Policies))
		for i := range spec.Policies {
			logrus.Infof("\t* Policy %d: %q", i, spec.Policies[i].Name)
		}
	}

	composeFile := Compose{
		spec:     *spec,
		filename: filename,
		name:     spec.Name,
	}

	composeFiles = append(composeFiles, composeFile)
	errs := []error{}
	for i := range spec.Include {
		composeFile := relativePathToFile(filename, []string{spec.Include[i]})[0]
		includedCompose, err := New(composeFile, visitedComposeFiles)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		composeFiles = append(composeFiles, includedCompose...)
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("included compose files errors: %s", errors.Join(errs...))
	}

	return composeFiles, nil
}

// GetPolicies returns a list of policies defined in the compose file
func (c *Compose) GetPolicies(disableTLS bool) ([]manifest.Manifest, error) {

	manifests := []manifest.Manifest{}
	var errs []error

	if err := c.spec.Environments.SetEnv(); err != nil {
		errs = append(errs, err)
	}

	if err := c.spec.Env_files.SetEnv(); err != nil {
		errs = append(errs, err)
	}

	var globalInlineValues string
	if c.spec.ValuesInline != nil {
		var err error
		globalInlineValues, err = parseValuesInline(*c.spec.ValuesInline)
		if err != nil {
			errs = append(errs, fmt.Errorf("parsing global inline values: %s", err))
		}
	}

	// Fail fast if there is an error with the compose file before processing policies
	if len(errs) > 0 {
		return nil, fmt.Errorf("compose file %q errors: %s", c.name, errors.Join(errs...))
	}

	for i := range c.spec.Policies {
		if c.spec.Policies[i].IsZero() {
			continue
		}

		logrus.Infof("\nInitializing policy: %q\n", c.spec.Policies[i].Name)

		var policyManifest, policyValues, policySecrets []string
		var err error

		if c.spec.Policies[i].Policy != "" {
			policyManifest, policyValues, policySecrets, err = registry.Pull(c.spec.Policies[i].Policy, disableTLS)
			if err != nil {
				errs = append(errs, fmt.Errorf("pulling policy %q: %s", c.spec.Policies[i].Policy, err))
				continue
			}
		}

		policyManifest = append(policyManifest, c.spec.Policies[i].Config...)
		policyValues = append(policyValues, c.spec.Policies[i].Values...)
		policySecrets = append(policySecrets, c.spec.Policies[i].Secrets...)
		showDetectedFiles := func(files []string, fileType string) {
			switch len(files) {
			case 0:
				logrus.Debugf("\t%s: nothing detected", fileType)
			case 1:
				logrus.Infof("\t%s: %q", fileType, files[0])
			default:
				logrus.Infof("\t%ss:", fileType)
				for i := range files {
					logrus.Infof("\t\t* %q", files[i])
				}
			}
		}

		showDetectedFiles(policyManifest, "manifest")
		showDetectedFiles(policyValues, "value")
		showDetectedFiles(policySecrets, "secret")

		manifest := manifest.Manifest{
			Manifests: policyManifest,
			Values:    policyValues,
			Secrets:   policySecrets,
		}

		if len(globalInlineValues) > 0 {
			manifest.ValuesInline = append(manifest.ValuesInline, globalInlineValues)
		}

		if c.spec.Policies[i].ValuesInline != nil {
			policyInlineValues, err := parseValuesInline(*c.spec.Policies[i].ValuesInline)
			if err != nil {
				errs = append(errs, fmt.Errorf("parsing inline values for policy %q: %s", c.spec.Policies[i].Name, err))
				continue
			}

			manifest.ValuesInline = append(manifest.ValuesInline, policyInlineValues)
		}

		manifests = append(manifests, manifest)
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("policies errors: %s", errors.Join(errs...))
	}

	return manifests, nil
}

// parseValuesInline returns a list of inline values defined in the compose file
func parseValuesInline(data map[string]any) (string, error) {

	valuesInline, err := yaml.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("marshaling inline values: %s", err)
	}

	return string(valuesInline), nil
}
