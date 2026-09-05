package gitgeneric

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/sirupsen/logrus"
)

// ErrNoCommitFound is returned when the history of a branch ends without any
// commit accepted by the caller's matcher.
var ErrNoCommitFound = errors.New("no commit found")

// ErrTruncatedHistory reports a history which ended on a commit whose parents were
// never fetched, as a shallow clone does. It accompanies ErrNoCommitFound so that a
// caller can tell "this branch has no matching commit" from "the commits which could
// have matched were never fetched".
var ErrTruncatedHistory = errors.New("truncated history")

// DatedCommit represents a git commit with its committer date.
type DatedCommit struct {
	When time.Time
	Hash string
}

// resolveBranchHash returns the hash referenced by a local or remote branch.
// An empty branch resolves the repository HEAD.
func resolveBranchHash(gitRepository *git.Repository, branch string) (plumbing.Hash, error) {
	if branch == "" {
		head, err := gitRepository.Head()
		if err != nil {
			return plumbing.ZeroHash, fmt.Errorf("getting HEAD: %s", err)
		}
		return head.Hash(), nil
	}

	references := []plumbing.ReferenceName{
		plumbing.NewBranchReferenceName(branch),
		plumbing.NewRemoteReferenceName(DefaultRemoteReferenceName, branch),
	}
	for _, referenceName := range references {
		reference, err := gitRepository.Reference(referenceName, true)
		if err == nil {
			return reference.Hash(), nil
		}
		if err != plumbing.ErrReferenceNotFound {
			return plumbing.ZeroHash, fmt.Errorf("getting branch %q: %s", branch, err)
		}
	}

	return plumbing.ZeroHash, fmt.Errorf("branch %q not found", branch)
}

// insertByCommitterTime inserts a commit into a list kept ordered by committer date,
// newest first.
func insertByCommitterTime(commits []*object.Commit, commit *object.Commit) []*object.Commit {
	index := sort.Search(len(commits), func(i int) bool {
		return commits[i].Committer.When.Before(commit.Committer.When)
	})

	commits = append(commits, nil)
	copy(commits[index+1:], commits[index:])
	commits[index] = commit

	return commits
}

/*
SearchCommit walks the history of a branch, newest committer date first, and returns
the first commit accepted by match.

The walk is lazy: it only loads the parents of the commits it visits, so a branch whose
tip already matches costs a single commit object rather than a full history read.

The history is walked by committer date rather than by the depth-first order which
go-git traverses by default: on a history containing merges, a depth-first walk yields
commits of a side branch before more recent ones of the mainline, so stopping on the
first match would not return the newest matching commit.

The history of a shallow repository ends on a commit whose parents were never fetched.
That truncation is indistinguishable from the end of a complete history here, so the
walk stops and reports ErrNoCommitFound wrapping ErrTruncatedHistory, leaving the
caller to explain it. Note that the boundary commit itself is matched before its
missing parents are looked up, which is why the walk is driven here rather than
through Repository.Log: the go-git iterators load the parents of a commit before
handing it over, so they would drop the oldest commit of a shallow clone.
*/
func (g GoGit) SearchCommit(workingDir, branch string, match func(when time.Time) bool) (DatedCommit, error) {
	gitRepository, err := git.PlainOpen(workingDir)
	if err != nil {
		return DatedCommit{}, fmt.Errorf("opening %q git directory: %s", workingDir, err)
	}

	hash, err := resolveBranchHash(gitRepository, branch)
	if err != nil {
		return DatedCommit{}, err
	}

	tip, err := object.GetCommit(gitRepository.Storer, hash)
	if err != nil {
		return DatedCommit{}, fmt.Errorf("getting commit %q: %s", hash, err)
	}

	// pending holds the commits left to visit, newest committer date first.
	pending := []*object.Commit{tip}
	seen := map[plumbing.Hash]bool{tip.Hash: true}
	truncated := false

	for len(pending) > 0 {
		commit := pending[0]
		pending = pending[1:]

		if match(commit.Committer.When) {
			return DatedCommit{
				Hash: commit.Hash.String(),
				When: commit.Committer.When,
			}, nil
		}

		logrus.Debugf("ignoring commit %q, dated %s, as rejected by the commit filter",
			commit.Hash.String(), commit.Committer.When)

		for _, parentHash := range commit.ParentHashes {
			if seen[parentHash] {
				continue
			}

			parent, err := object.GetCommit(gitRepository.Storer, parentHash)
			if err != nil {
				/*
					A missing parent means the history is truncated, which a shallow clone
					does on purpose, so this branch of the walk ends here instead of
					failing the whole lookup.
				*/
				if errors.Is(err, plumbing.ErrObjectNotFound) {
					logrus.Debugf("history of %q in %q is truncated at commit %q",
						branch, workingDir, commit.Hash.String())
					truncated = true
					continue
				}

				return DatedCommit{}, fmt.Errorf("getting commit %q: %s", parentHash, err)
			}

			seen[parentHash] = true
			pending = insertByCommitterTime(pending, parent)
		}
	}

	if truncated {
		return DatedCommit{}, fmt.Errorf("%w in the %w of %q", ErrNoCommitFound, ErrTruncatedHistory, branch)
	}

	return DatedCommit{}, fmt.Errorf("%w in the history of %q", ErrNoCommitFound, branch)
}
