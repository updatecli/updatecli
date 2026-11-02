package scm

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"

	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
)

// Local wraps an scm handler so a pipeline runs against an existing local checkout,
// such as the repository a manifest author is working in.
//
// Resources keep their scm, so paths resolve and are checked against the repository
// root exactly as they would be in the cloned checkout. Nothing touches the git state of
// the checkout: no clone, pull, checkout, commit, push or cleanup ever happens, so
// uncommitted work is safe.
//
// Every ScmHandler method needs a deliberate local behavior: only methods describing the
// repository are forwarded to the wrapped scm, whose own checkout never exists.
type Local struct {
	inner            ScmHandler
	directory        string
	nativeGitHandler gitgeneric.GitHandler
	// baseline maps the files already changed when a resource starts to a hash of their
	// content, so GetChangedFiles leaves out the author's own uncommitted work. It is nil
	// until Checkout records it.
	baseline map[string]string
}

var _ ScmHandler = (*Local)(nil)

// NewLocal returns a handler using directory as the checkout of inner.
func NewLocal(inner ScmHandler, directory string) *Local {
	return &Local{inner: inner, directory: directory, nativeGitHandler: &gitgeneric.GoGit{}}
}

// IsLocal reports whether handler runs against a local checkout.
//
// Resources running git commands themselves instead of through the handler, such as the
// gitbranch and gittag targets, use it to leave the local checkout untouched.
func IsLocal(handler ScmHandler) bool {
	_, ok := handler.(*Local)
	return ok
}

// LocalDirectory returns the local checkout containing path: the root of its git repository.
func LocalDirectory(path string) (string, error) {
	return (&gitgeneric.GoGit{}).RepositoryRoot(path)
}

// GetDirectory returns the local checkout.
func (l *Local) GetDirectory() string {
	return l.directory
}

// Clone does nothing: the local checkout already exists.
func (l *Local) Clone() (string, error) {
	log.Debugf("local scm: not cloning, using %q", l.directory)
	return l.directory, nil
}

// Checkout keeps the current branch and uncommitted changes. As every resource checks out its
// scm before running, it records the files already changed, so GetChangedFiles only reports
// what the resource changes afterwards.
func (l *Local) Checkout() error {
	files, err := l.nativeGitHandler.GetChangedFiles(l.directory)
	if err != nil {
		log.Debugf("local scm: reading the changes of %q: %s", l.directory, err)
		l.baseline = nil
		return nil
	}

	l.baseline = make(map[string]string, len(files))
	for _, file := range files {
		l.baseline[file] = contentHash(filepath.Join(l.directory, file))
	}

	return nil
}

// Clean does nothing: the local checkout must never be removed.
func (l *Local) Clean() error {
	return nil
}

// CleanWorkingBranch does nothing: no working branch is ever created.
func (l *Local) CleanWorkingBranch() (bool, error) {
	return false, nil
}

// Add does nothing: changes are left for the author to review.
func (l *Local) Add(files []string) error {
	log.Debugf("local scm: not staging %d file(s)", len(files))
	return nil
}

// Commit does nothing: changes are left for the author to review.
func (l *Local) Commit(_ context.Context, _ string) error {
	log.Debugf("local scm: not committing in %q", l.directory)
	return nil
}

// Push does nothing and reports that no branch was reset.
func (l *Local) Push() (bool, error) {
	return false, nil
}

// PushTag does nothing.
func (l *Local) PushTag(tag string) error {
	log.Debugf("local scm: not pushing tag %q", tag)
	return nil
}

// PushBranch does nothing.
func (l *Local) PushBranch(branch string) error {
	log.Debugf("local scm: not pushing branch %q", branch)
	return nil
}

// GetChangedFiles returns the files git sees as changed in workingDir, leaving out the ones
// already changed, with the same content, when Checkout ran.
func (l *Local) GetChangedFiles(workingDir string) ([]string, error) {
	files, err := l.nativeGitHandler.GetChangedFiles(workingDir)
	if err != nil || l.baseline == nil {
		return files, err
	}

	changed := []string{}
	for _, file := range files {
		before, found := l.baseline[file]
		if !found || before != contentHash(filepath.Join(workingDir, file)) {
			changed = append(changed, file)
		}
	}

	return changed, nil
}

// contentHash returns a hash of the content of file, or an empty string when it can't be
// read, such as a deleted file.
func contentHash(file string) string {
	content, err := os.ReadFile(file)
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%x", sha256.Sum256(content))
}

// IsRemoteBranchUpToDate reports true, as there is never anything to push.
func (l *Local) IsRemoteBranchUpToDate() (bool, error) {
	return true, nil
}

// IsRemoteWorkingBranchExist reports true, so --existing-only never skips a target:
// the local checkout plays the part of the working branch.
func (l *Local) IsRemoteWorkingBranchExist() (bool, error) {
	return true, nil
}

// GetBranches returns the branches of the wrapped scm.
func (l *Local) GetBranches() (sourceBranch, workingBranch, targetBranch string) {
	return l.inner.GetBranches()
}

// GetURL returns the URL of the wrapped scm.
func (l *Local) GetURL() string {
	return l.inner.GetURL()
}

// Summary describes the wrapped scm and the local checkout used in its place.
func (l *Local) Summary() string {
	return fmt.Sprintf("%s (local checkout %s)", l.inner.Summary(), l.directory)
}
