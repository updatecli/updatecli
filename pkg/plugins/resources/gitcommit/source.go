package gitcommit

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
)

// Source returns the latest commit hash for the configured Git branch.
func (gc *GitCommit) Source(_ context.Context, workingDir string, resultSource *result.Source) error {
	var err error

	gc.directory = workingDir
	if gc.spec.URL != "" {
		gc.directory, err = gc.clone()
		if err != nil {
			return fmt.Errorf("cloning Git repository: %w", err)
		}
	} else if gc.spec.Path != "" {
		gc.directory = gc.spec.Path
	}

	if gc.directory == "" {
		return fmt.Errorf("unknown Git working directory. Did you specify one of `spec.URL`, `scmid` or `spec.path`?")
	}

	branch := gc.spec.Branch
	if branch == "" {
		branch = "HEAD"
	}

	// Without an age filter the branch tip is the answer, so the history is left alone.
	if gc.spec.Age.IsZero() {
		hash, err := gc.nativeGitHandler.GetCommitHash(gc.directory, gc.spec.Branch)
		if err != nil {
			return fmt.Errorf("retrieving latest Git commit: %w", err)
		}

		resultSource.Result = result.SUCCESS
		resultSource.Information = hash
		resultSource.Description = fmt.Sprintf("Git commit %q found for branch %q", hash, branch)

		return nil
	}

	commit, err := gc.nativeGitHandler.SearchCommit(gc.directory, gc.spec.Branch, gc.spec.Age.Matches)
	if err != nil {
		/*
			The branch does have a history but the age filter discarded every commit of
			it, which means the commit we would have returned is still cooling down.
			That's an expected state of the filter rather than a lookup failure, so the
			source is skipped instead. A history truncated by `depth` ends up here too,
			hence the reminder.
		*/
		if errors.Is(err, gitgeneric.ErrNoCommitFound) {
			logrus.Debugf("%s", age.ErrNoVersionMatchingAge)
			logrus.Warningf("no commit of branch %q matched the age filter %s", branch, endOfHistory(err, gc.spec.Depth))

			resultSource.Result = result.SKIPPED
			resultSource.Description = "no git commit matches the age filter yet"

			return nil
		}

		return fmt.Errorf("retrieving latest Git commit: %w", err)
	}

	resultSource.Result = result.SUCCESS
	resultSource.Information = commit.Hash
	resultSource.Description = fmt.Sprintf("Git commit %q found for branch %q, committed on %s and matching the age filter",
		commit.Hash, branch, commit.When.Format(time.RFC3339))

	return nil
}

// endOfHistory explains why a history walk ran out of commits, telling a branch which
// genuinely holds no matching commit from one whose older commits were never fetched.
func endOfHistory(err error, depth *int) string {
	if !errors.Is(err, gitgeneric.ErrTruncatedHistory) {
		return "anywhere in its history"
	}

	if depth == nil || *depth == 0 {
		return "before the end of the shallow history available. A matching commit may exist further back, in commits which were never fetched"
	}

	return fmt.Sprintf("before the end of the %d commit(s) fetched by `spec.depth`. A matching commit may exist further back, consider raising it", *depth)
}
