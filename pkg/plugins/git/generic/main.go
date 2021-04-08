package generic

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	transportHttp "github.com/go-git/go-git/v5/plumbing/transport/http"
)

// Add run `git add`.
func Add(files []string, workingDir string) error {

	for _, file := range files {
		logrus.Infof("Adding file: %s", file)

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

	logrus.Debugf("Checkout branch '%v', based on '%v' from directory '%v'",
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

		logrus.Debugf("Branch '%v' doesn't exist, creating it from branch '%v'", remoteBranch, branch)

		err = w.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName(branch),
			Create: false,
			Keep:   false,
			Force:  true,
		})

		if err != nil {
			logrus.Debugf("Branch: '%v' - \n\t%v", branch, err)
			return err
		}

		// Checkout locale branch without creating it yet
		err = w.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName(remoteBranch),
			Create: true,
			Keep:   false,
			Force:  true,
		})

		logrus.Debugf("Creating branch '%v'", remoteBranch)

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
			logrus.Debugf("No remote name '%v'", remoteBranch)
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

	logrus.Infof("Commit changes")

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
	logrus.Infof("%s", status)

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
	logrus.Infof("%s", obj)

	return nil

}

// Clone run `git clone`.
func Clone(username, password, URL, workingDir string) error {

	var repo *git.Repository

	auth := transportHttp.BasicAuth{
		Username: username, // anything except an empty string
		Password: password,
	}

	var b bytes.Buffer

	b.WriteString(fmt.Sprintf("Cloning git repository: %s in %s\n", URL, workingDir))
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

		logrus.Infof(b.String())

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

	b.WriteString("Fetching remote branches")
	for _, r := range remotes {

		err := r.Fetch(&git.FetchOptions{
			Auth:     &auth,
			Progress: &b,
			RefSpecs: []config.RefSpec{"refs/*:refs/*"},
			Force:    true,
		})

		logrus.Infof(b.String())

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
func Push(username, password, workingDir string) error {

	auth := transportHttp.BasicAuth{
		Username: username, // anything excepted an empty string
		Password: password,
	}

	logrus.Infof("Push changes")

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

	refspec := config.RefSpec(fmt.Sprintf("+%s:refs/heads/%s",
		localRefSpec,
		localBranch))

	if err := refspec.Validate(); err != nil {
		return err
	}

	b := bytes.Buffer{}

	// Only push one branch at a time
	err = r.Push(&git.PushOptions{
		Auth:     &auth,
		Progress: &b,
		RefSpecs: []config.RefSpec{
			refspec,
		},
	})

	fmt.Println(b.String())

	if err != nil {
		return err
	}

	logrus.Infof("")

	return nil
}

// SanitizeBranchName replace wrong character in the branch name
func SanitizeBranchName(branch string) string {

	removedCharacter := []string{
		":", "=", "+", "-", "$", "&", "#", "!", "@", "*", " ",
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
		logrus.Errorf("opening %q git directory err: %s", err)
		return tags, err
	}

	tagrefs, err := r.Tags()

	if err != nil {
		return tags, err
	}

	err = tagrefs.ForEach(func(t *plumbing.Reference) error {
		tags = append(tags, t.Name().Short())
		return nil
	})

	if err != nil {
		return tags, err
	}

	logrus.Debugf("Got tags: %v", tags)

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
		logrus.Errorf("opening %q git directory err: %s", err)
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
func PushTag(tag, username, password, workingDir string) error {

	auth := transportHttp.BasicAuth{
		Username: username, // anything except an empty string
		Password: password,
	}

	r, err := git.PlainOpen(workingDir)

	if err != nil {
		logrus.Errorf("opening %q git directory err: %s", err)
		return err
	}

	logrus.Debugf("Pushing git Tag: %q", tag)

	po := &git.PushOptions{
		RemoteName: "origin",
		Progress:   os.Stdout,
		RefSpecs: []config.RefSpec{
			config.RefSpec(
				config.RefSpec("+refs/tags/" + tag + ":refs/tags/" + tag))},
		Auth: &auth,
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
