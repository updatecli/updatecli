package gitgeneric

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	transportHttp "github.com/go-git/go-git/v5/plumbing/transport/http"
)

// IsSimilarBranch checks that the last commits of the two branches are similar then return
// true if it's the case
func IsSimilarBranch(a, b, workingDir string) (bool, error) {

	gitRepository, err := git.PlainOpen(workingDir)
	if err != nil {
		return false, err
	}

	refA, err := gitRepository.Reference(plumbing.NewBranchReferenceName(a), true)
	if err != nil {
		logrus.Errorf("reference %q - %s", a, err)
		return false, err
	}

	refB, err := gitRepository.Reference(plumbing.NewBranchReferenceName(b), true)

	if err != nil {
		logrus.Errorf("reference %q - %s", b, err)
		return false, err
	}

	if refA.Hash().String() == refB.Hash().String() {
		return true, nil
	}

	return false, nil

}

func GetChangedFiles(workingDir string) ([]string, error) {
	gitRepository, err := git.PlainOpen(workingDir)
	if err != nil {
		return []string{}, err
	}

	gitWorktree, err := gitRepository.Worktree()
	if err != nil {
		return []string{}, err
	}

	gitStatus, err := gitWorktree.Status()
	if err != nil {
		return []string{}, err
	}

	// gitStatus is a list of changed (as per git definition) files in the provided workingDir
	// We only retrieve the list of changed files without checking the file status (https://pkg.go.dev/github.com/go-git/go-git/v5@v5.4.2?utm_source=gopls#StatusCode)
	// as we assume all changes have been done by the target and shall be added, even if already added
	filesChanged := []string{}
	for changedFile := range gitStatus {
		filesChanged = append(filesChanged, changedFile)
	}

	return filesChanged, nil
}

// Add run `git add`.
func Add(files []string, workingDir string) error {

	logrus.Debugf("stage: git-add\n\n")

	for _, file := range files {
		logrus.Debugf("adding file: %q\n", file)

		// Strip working directory from file path if it's provided as an absolute path
		if filepath.IsAbs(file) {
			relativeFilePath, err := filepath.Rel(workingDir, file)
			if err != nil {
				return err
			}
			logrus.Debugf("absolute file path detected: %q converted to relative path %q\n", file, relativeFilePath)
			file = relativeFilePath
		}

		r, err := git.PlainOpen(workingDir)
		if err != nil {
			return err
		}

		w, err := r.Worktree()
		if err != nil {
			return err
		}

		_, err = w.Add(file)
		if err != nil {
			return err
		}
	}
	return nil
}

// Checkout create and then uses a temporary git branch.
func Checkout(branch, remoteBranch, workingDir string) error {

	logrus.Debugf("stage: git-checkout\n\n")

	logrus.Debugf("checkout branch %q, based on %q to directory %q",
		remoteBranch,
		branch,
		workingDir)

	r, err := git.PlainOpen(workingDir)
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		logrus.Debugln(err)
		return err
	}

	b := bytes.Buffer{}

	err = w.Pull(&git.PullOptions{
		Force:    true,
		Progress: &b,
	})

	if err != nil &&
		err != git.ErrNonFastForwardUpdate &&
		err != git.NoErrAlreadyUpToDate {
		logrus.Debugln(err)
		return err
	}

	// If remoteBranch already exist, use it
	// otherwise use the one define in the spec
	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(remoteBranch),
		Create: false,
		Keep:   false,
		Force:  true,
	})

	if err == plumbing.ErrReferenceNotFound {
		// Checkout source branch without creating it yet

		logrus.Debugf("branch '%v' doesn't exist, creating it from branch '%v'", remoteBranch, branch)

		err = w.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName(branch),
			Create: false,
			Keep:   false,
			Force:  true,
		})

		if err != nil {
			logrus.Debugf("branch: '%v' - \n\t%v", branch, err)
			return err
		}

		// Checkout locale branch without creating it yet
		err = w.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName(remoteBranch),
			Create: true,
			Keep:   false,
			Force:  true,
		})

		logrus.Debugf("branch %q successfully created", remoteBranch)

		if err != nil {
			return err
		}

	} else if err != plumbing.ErrReferenceNotFound && err != nil {
		logrus.Debugln(err)
		return err
	} else {

		// Means that a local branch named remoteBranch already exist
		// so we want to be sure that the local branch is
		// aligned with the remote one.

		remote, err := r.Remote("origin")
		if err != nil {
			return err
		}
		refs, err := remote.List(&git.ListOptions{})

		if err != nil {
			return err
		}

		if !exists(
			plumbing.NewBranchReferenceName(remoteBranch),
			refs) {
			logrus.Debugf("No remote name %q", remoteBranch)
			return nil
		}

		remoteBranchRef := plumbing.NewRemoteReferenceName("origin", remoteBranch)

		remoteRef, err := r.Reference(remoteBranchRef, true)

		if err != nil {
			return err
		}

		err = w.Reset(&git.ResetOptions{
			Commit: remoteRef.Hash(),
			Mode:   git.HardReset,
		})

		if err != nil {
			logrus.Debugln(err)
			return err
		}

		err = w.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName(remoteBranch),
			Create: false,
			Keep:   false,
			Force:  true,
		})

		if err != nil {
			logrus.Debugln(err)
			return err
		}

	}

	return nil
}

func exists(ref plumbing.ReferenceName, refs []*plumbing.Reference) bool {
	for _, ref2 := range refs {
		if ref.String() == ref2.Name().String() {
			return true
		}
	}
	return false
}

// Commit run `git commit`.
func Commit(user, email, message, workingDir string) error {

	logrus.Debugf("stage: git-commit\n\n")

	r, err := git.PlainOpen(workingDir)
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	status, err := w.Status()
	if err != nil {
		return err
	}
	logrus.Debugf("status: %q\n", status)

	commit, err := w.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  user,
			Email: email,
			When:  time.Now(),
		},
	})
	if err != nil {
		return err
	}
	obj, err := r.CommitObject(commit)
	if err != nil {
		return err
	}
	logrus.Debugf("git commit object:\n%s\n", obj)

	return nil

}

// Clone run `git clone`.
func Clone(username, password, URL, workingDir string) error {

	logrus.Debugf("stage: git-clone\n\n")

	var repo *git.Repository

	auth := transportHttp.BasicAuth{
		Username: username, // anything except an empty string
		Password: password,
	}

	var b bytes.Buffer

	b.WriteString(fmt.Sprintf("cloning git repository: %s in %s\n", URL, workingDir))
	repo, err := git.PlainClone(workingDir, false, &git.CloneOptions{
		URL:               URL,
		Auth:              &auth,
		Progress:          &b,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})

	if err == git.ErrRepositoryAlreadyExists {
		b.Reset()

		repo, err = git.PlainOpen(workingDir)
		if err != nil {
			return err
		}

		w, err := repo.Worktree()
		if err != nil {
			return err
		}

		status, err := w.Status()
		if err != nil {
			return err
		}
		b.WriteString(status.String())

		err = w.Pull(&git.PullOptions{
			Auth:     &auth,
			Force:    true,
			Progress: &b,
		})

		logrus.Debugln(b.String())

		b.Reset()

		if err != nil &&
			err != git.ErrNonFastForwardUpdate &&
			err != git.NoErrAlreadyUpToDate {
			logrus.Debugln(err)
			return err
		}

	} else if err != nil &&
		err != git.NoErrAlreadyUpToDate {
		return err
	}

	remotes, err := repo.Remotes()

	if err != nil {
		return err
	}

	b.WriteString("fetching remote branches")
	for _, r := range remotes {

		err := r.Fetch(&git.FetchOptions{
			Auth:     &auth,
			Progress: &b,
			RefSpecs: []config.RefSpec{"refs/*:refs/*"},
			Force:    true,
		})

		logrus.Debugln(b.String())

		b.Reset()
		if err != nil &&
			err != git.NoErrAlreadyUpToDate &&
			err != git.ErrBranchExists {
			return err
		}
	}

	return err
}

// Push run `git push`.
func Push(username string, password string, workingDir string, force bool) error {

	logrus.Debugf("stage: git-push\n\n")

	auth := transportHttp.BasicAuth{
		Username: username, // anything excepted an empty string
		Password: password,
	}

	r, err := git.PlainOpen(workingDir)
	if err != nil {
		return err
	}

	// Retrieve local branch
	head, err := r.Head()
	if err != nil {
		return err
	}

	if !head.Name().IsBranch() {
		return fmt.Errorf("not pushing from a branch")
	}

	localBranch := strings.TrimPrefix(head.Name().String(), "refs/heads/")
	localRefSpec := head.Name().String()

	// By default don't force push
	refspec := config.RefSpec(fmt.Sprintf("%s:refs/heads/%s",
		localRefSpec,
		localBranch))

	if force {
		refspec = config.RefSpec(fmt.Sprintf("+%s:refs/heads/%s",
			localRefSpec,
			localBranch))
	}

	if err := refspec.Validate(); err != nil {
		return err
	}

	b := bytes.Buffer{}

	// Only push one branch at a time
	err = r.Push(&git.PushOptions{
		Auth:     &auth,
		Progress: &b,
		RefSpecs: []config.RefSpec{refspec},
	})

	logrus.Debugln(b.String())

	if err != nil {
		return err
	}

	return nil
}

// SanitizeBranchName replace wrong character in the branch name
func SanitizeBranchName(branch string) string {

	removedCharacter := []string{
		":", "=", "+", "$", "&", "#", "!", "@", "*", " ",
	}

	replacedByUnderscore := []string{
		"/", "\\", "{", "}", "[", "]",
	}

	for _, character := range removedCharacter {
		branch = strings.ReplaceAll(branch, character, "")
	}
	for _, character := range replacedByUnderscore {
		branch = strings.ReplaceAll(branch, character, "_")
	}

	if len(branch) > 255 {
		return branch[0:255]
	}
	return branch
}

// Tags return a list of git tags ordered by creation time
func Tags(workingDir string) (tags []string, err error) {
	r, err := git.PlainOpen(workingDir)
	if err != nil {
		logrus.Errorf("opening %q git directory err: %s", workingDir, err)
		return tags, err
	}

	tagrefs, err := r.Tags()
	if err != nil {
		return tags, err
	}

	type DatedTag struct {
		when time.Time
		name string
	}
	listOfDatedTags := []DatedTag{}

	err = tagrefs.ForEach(func(tagRef *plumbing.Reference) error {
		revision := plumbing.Revision(tagRef.Name().String())
		tagCommitHash, err := r.ResolveRevision(revision)
		if err != nil {
			return err
		}

		commit, err := r.CommitObject(*tagCommitHash)
		if err != nil {
			return err
		}

		listOfDatedTags = append(
			listOfDatedTags,
			DatedTag{
				name: tagRef.Name().Short(),
				when: commit.Committer.When,
			},
		)

		return nil
	})
	if err != nil {
		return tags, err
	}

	if len(listOfDatedTags) == 0 {
		err = errors.New("no tag found")
		return []string{}, err
	}

	// Sort tags by time
	sort.Slice(listOfDatedTags, func(i, j int) bool {
		return listOfDatedTags[i].when.Before(listOfDatedTags[j].when)
	})

	// Extract the list of tags names (ordered by time)
	for _, datedTag := range listOfDatedTags {
		tags = append(tags, datedTag.name)
	}

	logrus.Debugf("got tags: %v", tags)

	if len(tags) == 0 {
		err = errors.New("no tag found")
		return tags, err
	}

	return tags, err

}

// NewTag create a tag then return a boolean to indicate if
// the tag was created or not.
func NewTag(tag, message, workingDir string) (bool, error) {

	r, err := git.PlainOpen(workingDir)

	if err != nil {
		logrus.Errorf("opening %q git directory err: %s", workingDir, err)
		return false, err
	}

	h, err := r.Head()
	if err != nil {
		logrus.Errorf("get HEAD error: %s", err)
		return false, err
	}

	_, err = r.CreateTag(tag, h.Hash(), &git.CreateTagOptions{
		Message: message,
	})
	if err != nil {
		logrus.Errorf("create git tag error: %s", err)
		return false, err
	}
	return true, nil
}

// PushTag publish a single tag created locally
func PushTag(tag string, username string, password string, workingDir string, force bool) error {

	auth := transportHttp.BasicAuth{
		Username: username, // anything except an empty string
		Password: password,
	}

	r, err := git.PlainOpen(workingDir)

	if err != nil {
		logrus.Errorf("opening %q git directory err: %s", workingDir, err)
		return err
	}

	logrus.Debugf("Pushing git Tag: %q", tag)

	// By default don't force push
	refspec := config.RefSpec("refs/tags/" + tag + ":refs/tags/" + tag)

	if force {
		refspec = config.RefSpec("+refs/tags/" + tag + ":refs/tags/" + tag)
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
		logrus.Infof("push to remote origin error: %s", err)
		return err
	}

	return nil
}
