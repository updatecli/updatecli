package gitgeneric

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git/sign"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	transportHttp "github.com/go-git/go-git/v5/plumbing/transport/http"
)

const (
	DefaultRemoteReferenceName = "origin"
)

type GitHandler interface {
	Add(files []string, workingDir string) error
	Checkout(username, password, branch, remoteBranch, workingDir string, forceReset bool) error
	Clone(username, password, URL, workingDir string) error
	Commit(user, email, message, workingDir string, signingKey string, passphrase string) error
	GetChangedFiles(workingDir string) ([]string, error)
	IsSimilarBranch(a, b, workingDir string) (bool, error)
	IsLocalBranchPublished(baseBranch, workingBranch, username, password, workingDir string) (bool, error)
	NewTag(tag, message, workingDir string) (bool, error)
	NewBranch(branch, workingDir string) (bool, error)
	Push(username string, password string, workingDir string, force bool) error
	PushTag(tag string, username string, password string, workingDir string, force bool) error
	PushBranch(branch string, username string, password string, workingDir string, force bool) error
	RemoteURLs(workingDir string) (map[string]string, error)
	SanitizeBranchName(branch string) string
	Tags(workingDir string) (tags []string, err error)
	TagHashes(workingDir string) (hashes []string, err error)
	TagRefs(workingDir string) (refs []DatedTag, err error)
	Branches(workingDir string) (branches []string, err error)
}

type GoGit struct {
}

/*
IsLocalBranchPublished checks if the local working branch, contains any changes which should be published.

  - `git diff` is not yet supported by the go-git library.
    https://github.com/go-git/go-git/issues/36

We retrieve:
  - the latest commit from the local working branch
  - the latest commit from the base branch
  - the latest commit from the remote working branch if it exists.

Based on those three information we decide if local changes should be published.

If local working branch == local base branch and no remote working branch
=> This means that base branch and working are similar so nothing need to be pushed

If local working branch == local base branch != remote working branch
=> This means that at some point the remote branch diverge from base branch

If local working branch == local base branch == remote working branch
=> This means that everything is up to date

returns true if both the remote and local reference are similar.
*/
func (g GoGit) IsLocalBranchPublished(baseBranch, workingBranch, username, password, workingDir string) (bool, error) {

	logrus.Debugln("Checking if local changes have been done that should be published")

	auth := transportHttp.BasicAuth{
		Username: username, // anything excepted an empty string
		Password: password,
	}

	// Check if base branch and working branch have the same reference
	matching, err := g.IsSimilarBranch(baseBranch, workingBranch, workingDir)
	if err != nil {
		return false, err
	}

	gitRepository, err := git.PlainOpen(workingDir)
	if err != nil {
		return false, err
	}

	workingBranchReferenceName := workingBranch
	//
	rem, err := gitRepository.Remote(DefaultRemoteReferenceName)
	if err != nil {
		logrus.Errorf("reference %q - %s", workingBranchReferenceName, err)
		return false, err
	}

	listOptions := git.ListOptions{}
	if !isAuthEmpty(&auth) {
		listOptions.Auth = &auth
	}

	remoteRefs, err := rem.List(&listOptions)
	if err != nil {
		return false, err
	}

	var workingBranchRef *plumbing.Reference
	var remoteRef *plumbing.Reference

	workingBranchRef, err = gitRepository.Reference(plumbing.NewBranchReferenceName(workingBranchReferenceName), true)
	if err != nil {
		return false, err
	}

	// Looking for remote branch hash
	for id := range remoteRefs {
		if remoteRefs[id].Name().IsBranch() {
			// We are looking for the remote branch matching the local one
			if remoteRefs[id].Name().Short() != workingBranchReferenceName {
				continue
			}

			remoteRef = remoteRefs[id]

			// Local branch matches the remote one which means that local one is already
			// up to date. Nothing else to do
			if remoteRef.Hash().String() == workingBranchRef.Hash().String() {
				return true, nil
			}
			break
		}
	}

	if remoteRef == nil && matching {
		/*
			if there is no remote branch, and the local working
			branch equal the base branch then it means that nothing changed
		*/
		return true, nil
	}

	/*
		There is no way to git log between two branches at the moment using go-git
		https://github.com/go-git/go-git/issues/69
	*/

	return false, nil
}

// IsSimilarBranch checks that the last commits of the two branches are similar then return
// true if it's the case
func (g GoGit) IsSimilarBranch(a, b, workingDir string) (bool, error) {

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
		logrus.Debugf("no changes detected between branches %q and %q",
			a, b)
		return true, nil
	}

	return false, nil
}

func (g GoGit) GetChangedFiles(workingDir string) ([]string, error) {
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
func (g GoGit) Add(files []string, workingDir string) error {

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
func (g GoGit) Checkout(username, password, branch, remoteBranch, workingDir string, forceReset bool) error {

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

	//// Retrieve local branch
	//head, err := r.Head()
	//if err != nil {
	//	return err
	//}

	//if head.Name().IsBranch() {
	//	if head.Name().Short() == branch {
	//		return nil
	//	}
	//}

	b := bytes.Buffer{}

	auth := transportHttp.BasicAuth{
		Username: username, // anything except an empty string
		Password: password,
	}

	pullOptions := git.PullOptions{
		Force:    true,
		Progress: &b,
	}

	if !isAuthEmpty(&auth) {
		pullOptions.Auth = &auth
	}

	err = w.Pull(&pullOptions)

	logrus.Debugln(b.String())
	b.Reset()

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

	switch err {
	case nil:
		/*
			Means that a local branch named remoteBranch already exist
			so we want to be sure that the local branch is
			aligned with the remote one.
		*/

		remote, err := r.Remote(DefaultRemoteReferenceName)
		if err != nil {
			return err
		}

		listOptions := git.ListOptions{}

		if !isAuthEmpty(&auth) {
			listOptions.Auth = &auth
		}

		refs, err := remote.List(&listOptions)

		if err != nil {
			return err
		}

		if !g.exists(
			plumbing.NewBranchReferenceName(remoteBranch),
			refs) {
			logrus.Debugf("No remote name %q", remoteBranch)
			return nil
		}

		remoteBranchRef := plumbing.NewRemoteReferenceName(DefaultRemoteReferenceName, remoteBranch)

		remoteRef, err := r.Reference(remoteBranchRef, true)

		if err != nil {
			return err
		}

		if forceReset {
			err = w.Reset(&git.ResetOptions{
				Commit: remoteRef.Hash(),
				Mode:   git.HardReset,
			})

			if err != nil {
				logrus.Debugln(err)
				return err
			}
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

	case plumbing.ErrReferenceNotFound:
		/*
			Checkout source branch without creating it yet
		*/

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

	default:
		logrus.Debugln(err)
		return err
	}

	return nil
}

func (g GoGit) exists(ref plumbing.ReferenceName, refs []*plumbing.Reference) bool {
	for _, ref2 := range refs {
		if ref.String() == ref2.Name().String() {
			return true
		}
	}
	return false
}

// Commit run `git commit`.
func (g GoGit) Commit(user, email, message, workingDir string, signingKey string, passphrase string) error {

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

	commitOptions := git.CommitOptions{
		// Several plugin
		// We assume that updatecli is working from a clean worktree and can add all files that need to be tracked by git
		// Hence why we run git commit -A
		All: true,
		Author: &object.Signature{
			Name:  user,
			Email: email,
			When:  time.Now(),
		},
	}

	logrus.Debugf("status: %q\n", status)

	if len(signingKey) > 0 {
		key, err := sign.GetCommitSignKey(signingKey, passphrase)
		if err != nil {
			return err
		}
		commitOptions.SignKey = key
	}

	commit, err := w.Commit(message, &commitOptions)
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
func (g GoGit) Clone(username, password, URL, workingDir string) error {

	logrus.Debugf("stage: git-clone\n\n")

	var repo *git.Repository

	auth := transportHttp.BasicAuth{
		Username: username, // anything except an empty string
		Password: password,
	}

	var b bytes.Buffer
	cloneOptions := git.CloneOptions{
		URL:               URL,
		Progress:          &b,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	}

	if !isAuthEmpty(&auth) {
		cloneOptions.Auth = &auth
	}

	b.WriteString(fmt.Sprintf("cloning git repository: %s in %s\n", URL, workingDir))
	repo, err := git.PlainClone(workingDir, false, &cloneOptions)

	logrus.Debugln(b.String())
	b.Reset()

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

		pullOptions := git.PullOptions{
			Force:    true,
			Progress: &b,
		}

		if !isAuthEmpty(&auth) {
			pullOptions.Auth = &auth
		}

		err = w.Pull(&pullOptions)

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

		fetchOptions := git.FetchOptions{
			Progress: &b,
			RefSpecs: []config.RefSpec{"refs/*:refs/*"},
			Force:    true,
		}
		if !isAuthEmpty(&auth) {
			fetchOptions.Auth = &auth
		}

		err := r.Fetch(&fetchOptions)

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
func (g GoGit) Push(username string, password string, workingDir string, force bool) error {

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

	pushOptions := git.PushOptions{
		Auth:     &auth,
		Progress: &b,
		RefSpecs: []config.RefSpec{refspec},
	}

	if !isAuthEmpty(&auth) {
		pushOptions.Auth = &auth
	}

	// Only push one branch at a time
	err = r.Push(&pushOptions)

	logrus.Debugln(b.String())
	b.Reset()

	if err != nil {
		return err
	}

	return nil
}

// SanitizeBranchName replace wrong character in the branch name
func (g GoGit) SanitizeBranchName(branch string) string {

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

// TagHashes returns a list of the commit hashes for git tags ordered by creation time
func (g GoGit) TagHashes(workingDir string) (hashes []string, err error) {
	refs, err := g.TagRefs(workingDir)
	if err != nil {
		logrus.Errorf("problem finding tag references for %q, err: %s", workingDir, err)
		return hashes, err
	}

	// Extract the list of tag hashes (ordered by time)
	for _, ref := range refs {
		hashes = append(hashes, ref.Hash)
	}

	logrus.Debugf("got tags: %v", hashes)

	if len(hashes) == 0 {
		err = errors.New("no tag found")
		return hashes, err
	}

	return hashes, err

}

// Tags returns a list of the names for git tags ordered by creation time
func (g GoGit) Tags(workingDir string) (names []string, err error) {
	refs, err := g.TagRefs(workingDir)
	if err != nil {
		logrus.Errorf("problem finding tag references for %q, err: %s", workingDir, err)
		return names, err
	}

	// Extract the list of tag names (ordered by time)
	for _, ref := range refs {
		names = append(names, ref.Name)
	}

	logrus.Debugf("got tags: %v", names)

	if len(names) == 0 {
		err = errors.New("no tag found")
		return names, err
	}

	return names, err
}

type DatedTag struct {
	When time.Time
	Name string
	Hash string
}

// TagRefs returns a list of git tags ordered by creation time
func (g GoGit) TagRefs(workingDir string) (tags []DatedTag, err error) {
	r, err := git.PlainOpen(workingDir)
	if err != nil {
		logrus.Errorf("opening %q git directory err: %s", workingDir, err)
		return tags, err
	}
	tagrefs, err := r.Tags()
	if err != nil {
		return tags, err
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
				Name: tagRef.Name().Short(),
				Hash: tagRef.Hash().String(),
				When: commit.Committer.When,
			},
		)

		return nil
	})
	if err != nil {
		return tags, err
	}

	if len(listOfDatedTags) == 0 {
		err = errors.New("no tag found")
		return []DatedTag{}, err
	}

	// Sort tags by time
	sort.Slice(listOfDatedTags, func(i, j int) bool {
		return listOfDatedTags[i].When.Before(listOfDatedTags[j].When)
	})

	return listOfDatedTags, err
}

// NewTag create a tag then return a boolean to indicate if
// the tag was created or not.
func (g GoGit) NewTag(tag, message, workingDir string) (bool, error) {

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
func (g GoGit) PushTag(tag string, username string, password string, workingDir string, force bool) error {

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

	b := bytes.Buffer{}
	po := &git.PushOptions{
		RemoteName: DefaultRemoteReferenceName,
		Progress:   &b,
		RefSpecs:   []config.RefSpec{refspec},
	}

	if !isAuthEmpty(&auth) {
		po.Auth = &auth
	}

	err = r.Push(po)

	logrus.Debugln(b.String())
	b.Reset()

	if err != nil {
		if err == git.NoErrAlreadyUpToDate {
			logrus.Infof("%q remote was up to date, no push done", DefaultRemoteReferenceName)
			return nil
		}
		logrus.Infof("push to remote %q error: %s", DefaultRemoteReferenceName, err)
		return err
	}

	return nil
}

// RemoteURLs returns the list of remote URLs
func (g GoGit) RemoteURLs(workingDir string) (map[string]string, error) {
	remoteList := make(map[string]string)

	gitRepository, err := git.PlainOpenWithOptions(workingDir, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return remoteList, err
	}

	list, err := gitRepository.Remotes()
	if err != nil {
		return remoteList, err
	}

	for _, r := range list {
		remoteList[r.Config().Name] = r.Config().URLs[0]
	}

	return remoteList, nil
}

// isAuthEmpty return true if no username/password are defined
func isAuthEmpty(auth *transportHttp.BasicAuth) bool {
	return auth.Username == "" && auth.Password == ""
}
