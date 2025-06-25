package gitgeneric

import (
	"bytes"
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
	// ErrNoBranchFound is the error message when no branch is found
	ErrNoBranchFound = "no branch found"
	// ErrNoGitTagFound is the error message when no tag is found
	ErrNoGitTagFound = "no tag found"
)

type GitHandler interface {
	Add(files []string, workingDir string) error
	Checkout(username, password, branch, remoteBranch, workingDir string, forceReset bool) error
	Clone(username, password, URL, workingDir string, withSubmodules *bool) error
	Commit(user, email, message, workingDir string, signingKey string, passphrase string) error
	GetChangedFiles(workingDir string) ([]string, error)
	GetLatestCommitHash(workingDir string) (string, error)
	IsSimilarBranch(a, b, workingDir string) (bool, error)
	IsLocalBranchPublished(baseBranch, workingBranch, username, password, workingDir string) (bool, error)
	NewTag(tag, message, workingDir string) (bool, error)
	NewBranch(branch, workingDir string) (bool, error)
	Pull(username, password, gitRepositoryPath, branch string, singleBranch, forceReset bool) error
	Push(username string, password string, workingDir string, force bool) (bool, error)
	PushTag(tag string, username string, password string, workingDir string, force bool) error
	PushBranch(branch string, username string, password string, workingDir string, force bool) error
	RemoteURLs(workingDir string) (map[string]string, error)
	SanitizeBranchName(branch string) string
	Tags(workingDir string) (tags []string, err error)
	TagHashes(workingDir string) (hashes []string, err error)
	TagRefs(workingDir string) (refs []DatedTag, err error)
	Branches(workingDir string) (branches []string, err error)
	BranchRefs(workingDir string) (branches []DatedBranch, err error)
	IsForceReset() bool
}

type GoGit struct {
	// ForceReset is set to true if the branch has been reset to the latest commit of the based branch.
	ForceReset bool
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
		return false, fmt.Errorf("checking if branch %q and %q are similar - %s", baseBranch, workingBranch, err)
	}

	gitRepository, err := git.PlainOpen(workingDir)
	if err != nil {
		return false, fmt.Errorf("opening %q git directory: %s", workingDir, err)
	}

	workingBranchReferenceName := workingBranch
	//
	rem, err := gitRepository.Remote(DefaultRemoteReferenceName)
	if err != nil {
		return false, fmt.Errorf("remote %q - %s", DefaultRemoteReferenceName, err)
	}

	listOptions := git.ListOptions{}

	if !isAuthEmpty(&auth) {
		listOptions.Auth = &auth
	}
	remoteRefs, err := rem.List(&listOptions)
	if err != nil {
		return false, fmt.Errorf("listing remote %q - %s", DefaultRemoteReferenceName, err)
	}

	var workingBranchRef *plumbing.Reference
	var remoteRef *plumbing.Reference

	workingBranchRef, err = gitRepository.Reference(plumbing.NewBranchReferenceName(workingBranchReferenceName), true)
	if err != nil {
		return false, fmt.Errorf("reference %q - %s", workingBranchReferenceName, err)
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
		return false, fmt.Errorf("opening %q git directory: %s", workingDir, err)
	}

	refA, err := gitRepository.Reference(plumbing.NewBranchReferenceName(a), true)
	if err != nil {
		return false, fmt.Errorf("reference %q - %s", a, err)
	}

	refB, err := gitRepository.Reference(plumbing.NewBranchReferenceName(b), true)

	if err != nil {
		return false, fmt.Errorf("reference %q - %s", b, err)
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
		return []string{}, fmt.Errorf("opening %q git directory: %s", workingDir, err)
	}

	gitWorktree, err := gitRepository.Worktree()
	if err != nil {
		return []string{}, fmt.Errorf("opening %q git worktree: %s", workingDir, err)
	}

	gitStatus, err := gitWorktree.Status()
	if err != nil {
		return []string{}, fmt.Errorf("getting git status: %s", err)
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

// GetLatestCommitHash returns the latest commit hash from the working directory
func (g GoGit) GetLatestCommitHash(workingDir string) (string, error) {
	gitRepository, err := git.PlainOpen(workingDir)
	if err != nil {
		return "", fmt.Errorf("opening %q git directory: %s", workingDir, err)
	}

	head, err := gitRepository.Head()
	if err != nil {
		return "", fmt.Errorf("getting HEAD: %s", err)
	}

	return head.Hash().String(), nil
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
				return fmt.Errorf("failed to convert absolute file path %q to relative path: %w", file, err)
			}
			logrus.Debugf("absolute file path detected: %q converted to relative path %q\n", file, relativeFilePath)
			file = relativeFilePath
		}

		r, err := git.PlainOpen(workingDir)
		if err != nil {
			return fmt.Errorf("opening %q git directory: %w", workingDir, err)
		}

		w, err := r.Worktree()
		if err != nil {
			return fmt.Errorf("opening %q git worktree: %w", workingDir, err)
		}

		_, err = w.Add(file)
		if err != nil {
			return fmt.Errorf("adding file %q: %w", file, err)
		}
	}
	return nil
}

// Checkout create and then uses a temporary git branch.
func (g *GoGit) Checkout(username, password, basedBranch, newBranch, gitRepositoryPath string, forceReset bool) error {

	logrus.Debugf("checkout git branch %q, based on %q",
		newBranch,
		basedBranch)

	repository, err := git.PlainOpen(gitRepositoryPath)
	if err != nil {
		return fmt.Errorf("opening %q git directory: %w", gitRepositoryPath, err)
	}

	worktree, err := repository.Worktree()
	if err != nil {
		return fmt.Errorf("opening %q git worktree: %w", gitRepositoryPath, err)
	}

	auth := transportHttp.BasicAuth{
		Username: username, // anything except an empty string
		Password: password,
	}

	// Checkout source branch without creating it yet
	// If newBranch already exist, use it
	// otherwise use the one define in the spec
	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(newBranch),
		Create: false,
		Keep:   false,
		Force:  true,
	})

	switch err {
	case plumbing.ErrReferenceNotFound:
		// Means that the new branch doesn't exist

		logrus.Debugf("new branch %q doesn't exist, creating it from branch %q", newBranch, basedBranch)

		// First we need to checkout the based branch
		err = worktree.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName(basedBranch),
			Create: false,
			Keep:   false,
			Force:  true,
		})

		if err != nil {
			return fmt.Errorf("checking out branch %q - %s", basedBranch, err)
		}

		// Then we create the new branch
		err = worktree.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName(newBranch),
			Create: true,
			Keep:   false,
			Force:  true,
		})
		if err != nil {
			return fmt.Errorf("checking out branch %q - %s", newBranch, err)
		}

		logrus.Debugf("branch %q successfully created", newBranch)

	case nil:
		// Means that a local branch named remoteBranch already exist
		// so we want to be sure that the local branch is
		// aligned with the remote one.
		b := bytes.Buffer{}

		// Todo in a separated pullrequest, we should validate that we can remove the pull operation from the checkout function
		// as we are already doing it in the clone function.

		// Today the checkout function is call when Updatecli is started to clone git repositories
		// then after each resource execution that depends on a git repository.
		// For large repository, the pull can take a long time so ideally we would like to only do it once
		// when Updatecli is started.
		pullOptions := git.PullOptions{
			Force:    true,
			Progress: &b,
		}

		if !isAuthEmpty(&auth) {
			pullOptions.Auth = &auth
		}

		err = worktree.Pull(&pullOptions)
		if b.String() != "" {
			logrus.Debugln(b.String())
		}
		b.Reset()
		if err != nil &&
			err != git.ErrNonFastForwardUpdate &&
			err != git.NoErrAlreadyUpToDate {
			logrus.Debugln(err)
			return err
		}

		if forceReset {
			logrus.Debugf("Checking if branch %q diverged from %q:", newBranch, basedBranch)
			// If the newBranch diverged from the basedBranch, we need to reset it
			resetBranch, err := resetNewBranchToBaseBranch(
				plumbing.NewBranchReferenceName(newBranch),
				plumbing.NewBranchReferenceName(basedBranch),
				gitRepositoryPath,
			)
			g.ForceReset = resetBranch
			if err != nil {
				return err
			}
			if resetBranch {
				logrus.Debugf("\tbranch %q reset to %q", newBranch, basedBranch)
			} else {
				logrus.Debugf("\tall good,branch %q is ahead of %q", newBranch, basedBranch)
			}
		}

	default:
		return fmt.Errorf("checking out branch %q - %s", newBranch, err)
	}

	return nil
}

func (g *GoGit) IsForceReset() bool {
	// If ForceReset is set to true, it means that the branch has been reset
	// to the latest commit of the based branch.
	return g.ForceReset
}

// Commit run `git commit`.
func (g GoGit) Commit(user, email, message, workingDir string, signingKey string, passphrase string) error {

	logrus.Debugf("stage: git-commit\n\n")

	r, err := git.PlainOpen(workingDir)
	if err != nil {
		return fmt.Errorf("opening %q git directory: %w", workingDir, err)
	}

	w, err := r.Worktree()
	if err != nil {
		return fmt.Errorf("opening %q git worktree: %w", workingDir, err)
	}

	status, err := w.Status()
	if err != nil {
		return fmt.Errorf("getting git status: %w", err)
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
			return fmt.Errorf("getting commit sign key: %w", err)
		}
		commitOptions.SignKey = key
	}

	commit, err := w.Commit(message, &commitOptions)
	if err != nil {
		return fmt.Errorf("committing: %w", err)
	}
	obj, err := r.CommitObject(commit)
	if err != nil {
		return fmt.Errorf("getting commit object: %w", err)
	}
	logrus.Debugf("git commit object:\n%s\n", obj)

	return nil

}

// Clone run `git clone`.
func (g GoGit) Clone(username, password, URL, workingDir string, withSubmodules *bool) error {

	var repo *git.Repository

	auth := transportHttp.BasicAuth{
		Username: username, // anything except an empty string
		Password: password,
	}

	submodule := git.DefaultSubmoduleRecursionDepth
	if withSubmodules != nil && !*withSubmodules {
		submodule = git.NoRecurseSubmodules
	}

	var b bytes.Buffer
	cloneOptions := git.CloneOptions{
		URL:               URL,
		Progress:          &b,
		RecurseSubmodules: submodule,
	}

	if !isAuthEmpty(&auth) {
		cloneOptions.Auth = &auth
	}

	b.WriteString(fmt.Sprintf("cloning git repository: %s in %s\n", URL, workingDir))
	repo, err := git.PlainClone(workingDir, false, &cloneOptions)

	if b.String() != "" {
		logrus.Debugln(b.String())
	}
	b.Reset()

	if err == git.ErrRepositoryAlreadyExists {

		logrus.Debugf("repository already exists, trying to pull changes")
		b.Reset()

		repo, err = git.PlainOpen(workingDir)
		if err != nil {
			return fmt.Errorf("opening %q git directory: %w", workingDir, err)
		}

		w, err := repo.Worktree()
		if err != nil {
			return fmt.Errorf("opening %q git worktree: %w", workingDir, err)
		}

		status, err := w.Status()
		if err != nil {
			return fmt.Errorf("getting git status: %w", err)
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

		if b.String() != "" {
			logrus.Debugln(b.String())
		}
		b.Reset()

		if err != nil &&
			err != git.ErrNonFastForwardUpdate &&
			err != git.NoErrAlreadyUpToDate {
			return fmt.Errorf("pulling: %w", err)
		}

	} else if err != nil &&
		err != git.NoErrAlreadyUpToDate {
		// It's a common error to specify a password without a username
		// Considering that I don't see a use case for an empty username
		// I will return an error if a password is specified without a username
		// Feel free to open an issue if you think it's a bad idea.
		if username == "" && password != "" {
			return fmt.Errorf("cloning: %w\n\tIt appears that a password has been specified without a username, could it be related?", err)
		}
		return fmt.Errorf("cloning: %w", err)
	}

	remotes, err := repo.Remotes()

	if err != nil {
		return fmt.Errorf("listing remotes: %w", err)
	}

	logrus.Debugf("Fetching remote branches for resetting local ones")

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

		if b.String() != "" {
			logrus.Debugln(b.String())
		}
		b.Reset()

		if err != nil &&
			err != git.NoErrAlreadyUpToDate &&
			err != git.ErrBranchExists {
			return fmt.Errorf("fetching remote %q: %w", r.Config().Name, err)
		}
	}

	return nil
}

// Push run `git push`.
// The function returns a boolean to indicate if the push needed a force push or not.
// and an error if something went wrong.
func (g GoGit) Push(username string, password string, workingDir string, force bool) (bool, error) {

	logrus.Debugf("stage: git-push\n\n")

	auth := transportHttp.BasicAuth{
		Username: username, // anything excepted an empty string
		Password: password,
	}

	r, err := git.PlainOpen(workingDir)
	if err != nil {
		return false, fmt.Errorf("opening %q git directory: %s", workingDir, err)
	}

	// Retrieve local branch
	head, err := r.Head()
	if err != nil {
		return false, fmt.Errorf("getting HEAD: %s", err)
	}

	if !head.Name().IsBranch() {
		return false, fmt.Errorf("not pushing from a branch")
	}

	localBranch := strings.TrimPrefix(head.Name().String(), "refs/heads/")
	localRefSpec := head.Name().String()

	// By default don't force push
	refspec := config.RefSpec(fmt.Sprintf("%s:refs/heads/%s",
		localRefSpec,
		localBranch))

	if err := refspec.Validate(); err != nil {
		return false, fmt.Errorf("validating refspec: %s", err)
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
	if b.String() != "" {
		logrus.Debugln(b.String())
	}
	b.Reset()

	if err == nil || err == git.NoErrAlreadyUpToDate {
		return false, nil
	}

	// Even if push --force is allowed by the configuration,
	// we first want to see if the push is a fast forward or not.
	if strings.Contains(err.Error(), git.ErrNonFastForwardUpdate.Error()) {
		//if errors.Is(err, git.ErrNonFastForwardUpdate) {
		if !force {
			return false, fmt.Errorf("force push needed, use scm parameter \"force\" to true")
		}

		refspec = config.RefSpec(fmt.Sprintf("+%s:refs/heads/%s",
			localRefSpec,
			localBranch))

		pushOptions = git.PushOptions{
			Auth:     &auth,
			Progress: &b,
			RefSpecs: []config.RefSpec{refspec},
		}

		// Only push one branch at a time
		err = r.Push(&pushOptions)

		if b.String() != "" {
			logrus.Debugln(b.String())
		}
		b.Reset()

		if err != nil {
			return false, fmt.Errorf("push to remote origin: %s", err)
		}
		return true, nil
	}

	logrus.Errorf("%s", git.ErrNonFastForwardUpdate.Error())
	logrus.Errorf("push: %s", err)

	return false, err
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
		return hashes, fmt.Errorf("listing git tags in directory %q: %w", workingDir, err)
	}

	// Extract the list of tag hashes (ordered by time)
	for _, ref := range refs {
		hashes = append(hashes, ref.Hash)
	}

	logrus.Debugf("got tags: %v", hashes)

	if len(hashes) == 0 {
		return hashes, fmt.Errorf("%s", ErrNoGitTagFound)
	}

	return hashes, nil

}

// Tags returns a list of the names for git tags ordered by creation time
func (g GoGit) Tags(workingDir string) (names []string, err error) {
	refs, err := g.TagRefs(workingDir)
	if err != nil {
		return names, fmt.Errorf("listing git tags in directory %q: %w", workingDir, err)
	}

	// Extract the list of tag names (ordered by time)
	for _, ref := range refs {
		names = append(names, ref.Name)
	}

	logrus.Debugf("got tags: %v", names)

	if len(names) == 0 {
		return names, fmt.Errorf("%s", ErrNoGitTagFound)
	}

	return names, nil
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
		return tags, fmt.Errorf("opening %q git directory: %w", workingDir, err)
	}
	tagrefs, err := r.Tags()
	if err != nil {
		return tags, fmt.Errorf("listing tags: %w", err)
	}

	listOfDatedTags := []DatedTag{}

	err = tagrefs.ForEach(func(tagRef *plumbing.Reference) error {
		// cfr https://github.com/updatecli/updatecli/issues/1392
		// using reference hash instead of reference name

		revision := plumbing.Revision(tagRef.Hash().String())

		hash := ""

		obj, err := r.TagObject(tagRef.Hash())
		switch err {
		case nil:
			// Annotated Tag detected
			revision = plumbing.Revision(obj.Target.String())

			hash = obj.Target.String()

		case plumbing.ErrObjectNotFound:
			// Not an annotated tag object
			hash = tagRef.Hash().String()

		default:
			return fmt.Errorf("getting tag object for %q: %w", tagRef.Name().Short(), err)
		}

		tagCommitHash, err := r.ResolveRevision(revision)
		if err != nil {
			return fmt.Errorf("resolving revision %q: %w", revision, err)
		}

		commit, err := r.CommitObject(*tagCommitHash)
		if err != nil {
			return fmt.Errorf("getting commit object for %q: %w", tagRef.Name().Short(), err)
		}

		listOfDatedTags = append(
			listOfDatedTags,
			DatedTag{
				Name: tagRef.Name().Short(),
				Hash: hash,
				When: commit.Committer.When,
			},
		)

		return nil
	})
	if err != nil {
		return tags, fmt.Errorf("listing tags: %w", err)
	}

	if len(listOfDatedTags) == 0 {
		return []DatedTag{}, fmt.Errorf("%s", ErrNoGitTagFound)
	}

	// Sort tags by time
	sort.Slice(listOfDatedTags, func(i, j int) bool {
		return listOfDatedTags[i].When.Before(listOfDatedTags[j].When)
	})

	return listOfDatedTags, nil
}

// NewTag create a tag then return a boolean to indicate if
// the tag was created or not.
func (g GoGit) NewTag(tag, message, workingDir string) (bool, error) {

	r, err := git.PlainOpen(workingDir)

	if err != nil {
		return false, fmt.Errorf("opening %q git directory: %w", workingDir, err)
	}

	h, err := r.Head()
	if err != nil {
		return false, fmt.Errorf("get HEAD: %w", err)
	}

	_, err = r.CreateTag(tag, h.Hash(), &git.CreateTagOptions{
		Message: message,
	})
	if err != nil {
		return false, fmt.Errorf("creating tag %q: %w", tag, err)
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
		return fmt.Errorf("opening %q git directory: %w", workingDir, err)
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
		return fmt.Errorf("push to remote %q: %s", DefaultRemoteReferenceName, err)
	}

	return nil
}

// RemoteURLs returns the list of remote URLs
func (g GoGit) RemoteURLs(workingDir string) (map[string]string, error) {
	remoteList := make(map[string]string)

	gitRepository, err := git.PlainOpenWithOptions(workingDir, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return remoteList, fmt.Errorf("opening %q git directory: %w", workingDir, err)
	}

	list, err := gitRepository.Remotes()
	if err != nil {
		return remoteList, fmt.Errorf("listing remotes: %w", err)
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

// Pull run `git pull` on the local HEAD.
func (g GoGit) Pull(username, password, gitRepositoryPath, branch string, singleBranch, forceReset bool) error {

	repository, err := git.PlainOpen(gitRepositoryPath)
	if err != nil {
		return fmt.Errorf("opening %q git directory: %w", gitRepositoryPath, err)
	}

	ref, err := repository.Head()
	if err != nil {
		return fmt.Errorf("get branch head: %s", err)
	}

	logrus.Debugf("Pulling git branch %q", ref.Name())

	worktree, err := repository.Worktree()
	if err != nil {
		return fmt.Errorf("opening %q git worktree: %w", gitRepositoryPath, err)
	}

	auth := transportHttp.BasicAuth{
		Username: username, // anything except an empty string
		Password: password,
	}

	b := bytes.Buffer{}

	pullOptions := git.PullOptions{
		RemoteName:  DefaultRemoteReferenceName,
		ReferenceName: plumbing.NewRemoteReferenceName(DefaultRemoteReferenceName, branch),
		Force:        forceReset,
		Progress:     &b,
		SingleBranch: singleBranch,
	}

	if branch != "" {
		pullOptions.ReferenceName = plumbing.NewBranchReferenceName(branch)
	}

	if !isAuthEmpty(&auth) {
		pullOptions.Auth = &auth
	}

	err = worktree.Pull(&pullOptions)
	if b.String() != "" {
		logrus.Debugln(b.String())
	}

	if err != nil &&
		err != git.ErrNonFastForwardUpdate &&
		err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("pulling remote branch %s", err)
	}

	ref, err = repository.Head()
	if err != nil {
		return fmt.Errorf("get branch head: %s", err)
	}

	logrus.Debugf("branch %q is now at %s", ref.Name(), ref.Hash())

	return nil
}
