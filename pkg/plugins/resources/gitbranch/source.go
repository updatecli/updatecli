package gitbranch

import (
	"context"
	"errors"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
)

// Source returns the latest git tag based on create time
func (gb *GitBranch) Source(_ context.Context, pathResolver pathresolver.Resolver, resultSource *result.Source) error {
	var err error

	gb.directory = pathResolver.RepositoryDir()
	if gb.spec.URL != "" && !gb.lsRemote {
		gb.directory, err = gb.clone()
		if err != nil {
			return err
		}

	} else if gb.spec.Path != "" {
		gb.directory = pathResolver.JoinManifest(gb.spec.Path)
	}

	if gb.directory == "" && !gb.lsRemote {
		return fmt.Errorf("unknown Git working directory. Did you specify one of `spec.URL`, `scmid` or a `spec.path`?")
	}

	refs, err := gb.branchRefs()

	if err != nil {
		return fmt.Errorf("retrieving branches: %w", err)
	}

	// published holds every branch, values the ones within the age window
	var published, values []string
	for _, ref := range refs {
		published = append(published, ref.Name)
		if !gb.spec.Age.Matches(ref.When) {
			log.Debugf("ignoring branch %q, dated %s, as outside of the age window", ref.Name, ref.When)
			continue
		}
		values = append(values, ref.Name)
	}

	gb.foundVersion, err = age.Search(&gb.versionFilter, published, values)
	if err != nil {
		/*
			Every branch matching the version filter is still cooling down, which is an
			expected state of the age filter rather than a failure, so the source is skipped
			instead.
		*/
		if errors.Is(err, age.ErrNoVersionMatchingAge) {
			log.Debugf("%s", err)
			resultSource.Result = result.SKIPPED
			resultSource.Description = "no git branch matching the version filter matches the age filter yet"
			return nil
		}
		return fmt.Errorf("filtering branches: %w", err)
	}
	value := gb.foundVersion.GetVersion()

	if len(value) == 0 {
		return fmt.Errorf("no Git Branch found matching pattern %q", gb.versionFilter.Pattern)
	}

	if gb.spec.Key == "hash" {
		for i := range refs {
			if refs[i].Name == value {
				value = refs[i].Hash
			}
		}
	}

	resultSource.Result = result.SUCCESS
	resultSource.Information = value
	resultSource.Description = fmt.Sprintf("git branch %q found matching pattern %q", value, gb.versionFilter.Pattern)

	return nil
}
