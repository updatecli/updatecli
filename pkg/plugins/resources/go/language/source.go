package language

import (
	"context"
	"errors"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
)

// Source returns the latest go module version
func (l *Language) Source(ctx context.Context, pathResolver pathresolver.Resolver, resultSource *result.Source) error {

	var err error

	switch l.Spec.Age.IsZero() {
	case true:
		_, err = l.versions(ctx)
		if err != nil {
			return fmt.Errorf("searching golang version: %w", err)
		}
	case false:
		_, err = l.getTagsFromRepository()
		if err != nil {
			/*
				Every version matching the version filter is still cooling down, which is an
				expected state of the age filter rather than a failure, so the source is
				skipped instead.
			*/
			if errors.Is(err, age.ErrNoVersionMatchingAge) {
				log.Debugf("%s", err)
				resultSource.Result = result.SKIPPED
				resultSource.Description = "no Golang version matching the version filter matches the age filter yet"
				return nil
			}
			return fmt.Errorf("searching golang version from repository: %w", err)
		}
	}

	resultSource.Information = l.Version.GetVersion()

	if resultSource.Information == "" {
		return fmt.Errorf("no Golang version found matching pattern %q",
			l.Spec.VersionFilter.Pattern,
		)
	}

	resultSource.Result = result.SUCCESS
	resultSource.Description = fmt.Sprintf("Golang version %s found",
		l.Version.GetVersion(),
	)

	return nil

}
