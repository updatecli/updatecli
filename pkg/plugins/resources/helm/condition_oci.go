package helm

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Condition checks if a Helm chart version exists on a OCI registry
// It assumes that not being able to retrieve the OCI digest, means, the helm chart doesn't exist.
func (c *Chart) OCICondition(source string, scm scm.ScmHandler, resultCondition *result.Condition) error {

	refName := filepath.Join(strings.TrimPrefix(c.spec.URL, "oci://"), c.spec.Name)
	switch c.spec.Version == "" {
	case true:
		refName += ":" + source
	case false:
		refName += ":" + c.spec.Version
	}

	ref, err := name.ParseReference(refName)
	if err != nil {
		return fmt.Errorf("invalid artifact %s: %w", refName, err)
	}

	_, err = remote.Head(ref, c.options...)
	if err != nil {
		if strings.Contains(err.Error(), "unexpected status code 404") {
			return fmt.Errorf("the OCI Helm chart %s doesn't exist", ref.Name())
		}
		return err
	}

	resultCondition.Pass = true
	resultCondition.Result = result.SUCCESS
	resultCondition.Description = fmt.Sprintf("The OCI Helm chart %s exists and is available", ref.Name())

	return nil
}
