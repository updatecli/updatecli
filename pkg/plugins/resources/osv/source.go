package osv

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source returns either the lowest version without known vulnerabilities or the IDs of the
// known vulnerabilities, depending on the spec key.
func (o *Osv) Source(ctx context.Context, workingDir string, resultSource *result.Source) error {
	if o.spec.Version == "" {
		return errors.New("osv source requires a version")
	}

	if o.spec.Key == KeyIDs {
		groups, err := o.knownVulnerabilities(ctx, o.spec.Version)
		if err != nil {
			return err
		}

		ids := make([]string, 0, len(groups))
		for _, group := range groups {
			logrus.Infof("  * %s", group)
			ids = append(ids, group.ID)
		}

		resultSource.Information = strings.Join(ids, ",")
		resultSource.Result = result.SUCCESS
		resultSource.Description = fmt.Sprintf("%d known vulnerabilities found for version %q of %s", len(ids), o.spec.Version, o.packageLabel())

		return nil
	}

	fixed, err := o.fixedVersion(ctx)
	if err != nil {
		/*
			No published version fixes every known vulnerability yet, which is an expected state
			rather than a failure, so the source is skipped instead.
		*/
		if errors.Is(err, ErrNoFixedVersion) {
			resultSource.Result = result.SKIPPED
			resultSource.Description = fmt.Sprintf("%s for version %q of %s", err, o.spec.Version, o.packageLabel())
			return nil
		}

		return err
	}

	resultSource.Information = fixed
	resultSource.Result = result.SUCCESS
	if fixed == o.spec.Version {
		resultSource.Description = fmt.Sprintf("version %q of %s has no known vulnerabilities", fixed, o.packageLabel())
	} else {
		resultSource.Description = fmt.Sprintf("version %q of %s fixes the known vulnerabilities of version %q", fixed, o.packageLabel(), o.spec.Version)
	}

	return nil
}
