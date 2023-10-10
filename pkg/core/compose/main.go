package compose

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/engine/manifest"
	"github.com/updatecli/updatecli/pkg/core/registry"
)

// Compose is a struct that contains a compose object
type Compose struct {
	// spec contains the compose spec
	spec Spec
}

// New creates a new Compose object
func New(filename string) (Compose, error) {
	var c Compose

	spec, err := LoadFile(filename)
	if err != nil {
		return c, err
	}

	c.spec = *spec

	return c, nil
}

// GetPolicies returns a list of policies defined in the compose file
func (c *Compose) GetPolicies(disableTLS bool) ([]manifest.Manifest, error) {
	var manifests []manifest.Manifest
	var errs []error

	if err := c.spec.Environments.SetEnv(); err != nil {
		errs = append(errs, err)
	}

	if err := c.spec.Env_files.SetEnv(); err != nil {
		errs = append(errs, err)
	}

	for i := range c.spec.Policies {
		if c.spec.Policies[i].Policy != "" {
			policyManifest, policyValues, policySecrets, err := registry.Pull(c.spec.Policies[i].Policy, disableTLS)
			if err != nil {
				errs = append(errs, fmt.Errorf("pulling policy %q: %s", c.spec.Policies[i].Policy, err))
				continue
			}

			policyManifest = append(policyManifest, c.spec.Policies[i].Config...)
			policyValues = append(policyValues, c.spec.Policies[i].Values...)
			policySecrets = append(policySecrets, c.spec.Policies[i].Secrets...)

			manifests = append(manifests, manifest.Manifest{
				Manifests: policyManifest,
				Values:    policyValues,
				Secrets:   policySecrets,
			})
		}
	}

	if len(errs) > 0 {
		return manifests, fmt.Errorf("policies errors: %s", errs)
	}

	return manifests, nil
}
