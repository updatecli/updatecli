package gitgeneric

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	transportHttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/sirupsen/logrus"
)

type DatedBranch struct {
	When time.Time
	Name string
	Hash string
}

// BranchRefs return a list of git granch ordered by latest commit time
func (g GoGit) BranchRefs(workingDir string) (branches []DatedBranch, err error) {
	r, err := git.PlainOpen(workingDir)
	if err != nil {
		logrus.Errorf("opening %q git directory: %s", workingDir, err)
		return branches, fmt.Errorf("opening %q git directory: %s", workingDir, err)

	}

	branchrefs, err := r.Branches()
	if err != nil {
		return branches, fmt.Errorf("listing branches: %s", err)
	}

	err = branchrefs.ForEach(func(tagRef *plumbing.Reference) error {
		revision := plumbing.Revision(tagRef.Name().String())
		tagCommitHash, err := r.ResolveRevision(revision)
		if err != nil {
			return fmt.Errorf("resolving revision %q: %s", revision, err)
		}

		commit, err := r.CommitObject(*tagCommitHash)
		if err != nil {
			return fmt.Errorf("getting commit object for %q: %s", tagRef.Name().Short(), err)
		}

		branches = append(
			branches,
			DatedBranch{
				Name: tagRef.Name().Short(),
				When: commit.Committer.When,
				Hash: commit.Hash.String(),
			},
		)

		return nil
	})
	if err != nil {
		return []DatedBranch{}, fmt.Errorf("listing branches: %s", err)
	}

	if len(branches) == 0 {
		return []DatedBranch{}, fmt.Errorf(ErrNoBranchFound)
	}

	// Sort tags by time
	sort.Slice(branches, func(i, j int) bool {
		return branches[i].When.Before(branches[j].When)
	})

	logrus.Debugf("got branch refs: %v", branches)

	if len(branches) == 0 {
		return branches, fmt.Errorf(ErrNoBranchFound)
	}

	return branches, err

}

// Branches return a list of git branches name ordered by latest commit time
func (g GoGit) Branches(workingDir string) (branches []string, err error) {
	branchRefs, err := g.BranchRefs(workingDir)
	if err != nil {
		return branches, err
	}
	// Extract the list of tags names (ordered by time)
	for _, branchRef := range branchRefs {
		branches = append(branches, branchRef.Name)
	}

	logrus.Debugf("got branches: %v", branches)

	if len(branches) == 0 {
		return branches, fmt.Errorf(ErrNoBranchFound)
	}

	return branches, err

}

// NewBranch create a tag then return a boolean to indicate if
// the tag was created or not.
func (g GoGit) NewBranch(branch, workingDir string) (bool, error) {

	r, err := git.PlainOpen(workingDir)

	if err != nil {
		return false, fmt.Errorf("opening %q git directory: %s", workingDir, err)
	}

	// Retrieve local branch
	head, err := r.Head()
	if err != nil {
		return false, err
	}

	if !head.Name().IsBranch() {
		return false, fmt.Errorf("Branch %q is not a branch", head.Name())
	}

	// Create a new plumbing.HashReference object with the name of the branch
	// and the hash from the HEAD. The reference name should be a full reference
	// name and not an abbreviated one, as is used on the git cli.
	//
	// For tags we should use `refs/tags/%s` instead of `refs/heads/%s` used
	// for branches.
	refName := plumbing.NewBranchReferenceName(branch)
	ref := plumbing.NewHashReference(refName, head.Hash())

	// The created reference is saved in the storage.
	err = r.Storer.SetReference(ref)

	if err != nil {
		return false, fmt.Errorf("create git branch: %s", err)
	}
	return true, nil
}

// PushBranch publish a single branch created locally
func (g GoGit) PushBranch(branch string, username string, password string, workingDir string, force bool) error {

	auth := transportHttp.BasicAuth{
		Username: username, // anything except an empty string
		Password: password,
	}

	r, err := git.PlainOpen(workingDir)

	if err != nil {
		return fmt.Errorf("opening %q git directory: %s", workingDir, err)
	}

	logrus.Debugf("Pushing git branch: %q", branch)

	// By default don't force push
	refspec := config.RefSpec("refs/heads/" + branch + ":refs/heads/" + branch)

	if force {
		refspec = config.RefSpec("+refs/heads/" + branch + ":refs/heads/" + branch)
	}

	po := &git.PushOptions{
		RemoteName: "origin",
		Progress:   os.Stdout,
		RefSpecs:   []config.RefSpec{refspec},
		Auth:       &auth,
	}

	err = r.Push(po)

	if err != nil {
		if err == git.NoErrAlreadyUpToDate {
			logrus.Info("origin remote was up to date, no push done")
			return nil
		}
		return fmt.Errorf("push to remote origin: %s", err)
	}

	return nil
}

// isBranchCommonAncestor checks if a new branch has common ancestor with the latest
// commit of the based branch.
func isBranchCommonAncestor(newBranch, basedBranch plumbing.ReferenceName, gitRepositoryPath string) (bool, error) {
	r, err := git.PlainOpen(gitRepositoryPath)

	if err != nil {
		return false, fmt.Errorf("opening %q git directory: %s", gitRepositoryPath, err)
	}

	newBranchRefName, err := r.Reference(newBranch, true)
	if err != nil {
		return false, fmt.Errorf("getting new branch reference %q: %s", newBranch, err)
	}

	basedBranchRefName, err := r.Reference(basedBranch, true)
	if err != nil {
		return false, fmt.Errorf("getting based branch reference %q: %s", basedBranch, err)
	}

	newBranchRefCommit, err := r.CommitObject(newBranchRefName.Hash())
	if err != nil {
		return false, fmt.Errorf("getting commit object for new branch %q: %s", newBranch, err)
	}

	basedBranchRefCommit, err := r.CommitObject(basedBranchRefName.Hash())
	if err != nil {
		return false, fmt.Errorf("getting commit object for based branch %q: %s", basedBranch, err)
	}

	return basedBranchRefCommit.IsAncestor(newBranchRefCommit)
}

// resetNewBranchToBaseBranch reset the new branch to the latest commit of the based branch
// if they don't have common ancestor
func resetNewBranchToBaseBranch(newBranch, basedBranch plumbing.ReferenceName, gitRepositoryPath string) (bool, error) {

	ok, err := isBranchCommonAncestor(newBranch, basedBranch, gitRepositoryPath)
	if err != nil {
		return false, fmt.Errorf("checking if %q has common ancestor with %q - %s", newBranch, basedBranch, err)
	}

	// Nothing to do if the new branch has common ancestor with the based branch
	if ok {
		return false, nil
	}

	logrus.Warningf("branch %q diverged from %q, resetting it to %q", newBranch, basedBranch, basedBranch)

	repository, err := git.PlainOpen(gitRepositoryPath)
	if err != nil {
		logrus.Errorf("opening %q git directory: %s", gitRepositoryPath, err)
		return false, fmt.Errorf("opening %q git directory: %s", gitRepositoryPath, err)
	}

	ref, err := repository.Reference(basedBranch, true)
	if err != nil {
		return false, fmt.Errorf("reference %q - %s", basedBranch, err)
	}

	worktree, err := repository.Worktree()
	if err != nil {
		return false, fmt.Errorf("loading work tree: %s", err)
	}
	err = worktree.Reset(&git.ResetOptions{
		Commit: ref.Hash(),
		Mode:   git.HardReset,
	})

	if err != nil {
		return false, fmt.Errorf("resetting branch %q to %q - %s", newBranch, basedBranch, err)
	}
	return true, nil
}
