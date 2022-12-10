package helm

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Condition checks if a Helm chart version exists on a OCI registry
// We assume that if we can't retrieve the oci digest, then it means
// it doesn't exist.
func (c *Chart) OCICondition(source string) (bool, error) {

	refName := filepath.Join(strings.TrimPrefix(c.spec.URL, "oci://"), c.spec.Name)
	switch c.spec.Version == "" {
	case true:
		refName += ":" + source
	case false:
		refName += ":" + c.spec.Version
	}

	ref, err := name.ParseReference(refName)
	if err != nil {
		return false, fmt.Errorf("invalid artifact %s: %w", refName, err)
	}

	_, err = remote.Head(ref, c.options...)
	if err != nil {
		if strings.Contains(err.Error(), "unexpected status code 404") {
			logrus.Infof("%s The OCI Helm chart %s doesn't exist.",
				result.FAILURE,
				ref.Name(),
			)
			return false, nil
		}
		return false, err
	}

	logrus.Infof("%s The OCI Helm chart %s exists and is available.",
		result.SUCCESS,
		ref.Name(),
	)

	return true, err
}
