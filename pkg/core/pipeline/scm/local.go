package scm

import (
	"context"
	"crypto/sha256"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
	"github.com/updatecli/updatecli/pkg/plugins/utils/redact"
)

// Local wraps an scm handler so a pipeline runs against an existing local checkout,
// such as the repository a manifest author is working in. UsesLocalCheckout tells which scms
// describe that repository.
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

// localCheckout remembers, for the whole run, which repository the local checkout stands for,
// so every scm can be compared with it.
var localCheckout = struct {
	sync.Mutex
	// directory is the checkout the fields below were resolved for.
	directory string
	// remotes maps the repository key of each remote of the checkout to the remote name.
	remotes map[string]string
	// branch is the branch checked out in the checkout, empty in detached HEAD.
	branch string
	// branchKnown reports whether branch could be read.
	branchKnown bool
	// reported holds the scm URLs already logged, so each is logged once although scms are
	// rebuilt before every resource.
	reported map[string]bool
	// reportedBranches holds the scm URL and branch pairs already logged as not matching the
	// branch of the checkout.
	reportedBranches map[string]bool
}{}

// resetLocalCheckout forgets the repository the local checkout stands for.
func resetLocalCheckout() {
	localCheckout.Lock()
	defer localCheckout.Unlock()

	localCheckout.directory = ""
	localCheckout.remotes = nil
	localCheckout.branch = ""
	localCheckout.branchKnown = false
	localCheckout.reported = nil
	localCheckout.reportedBranches = nil
}

// UsesLocalCheckout reports whether handler describes the repository of the local checkout
// in directory, by comparing its URL with every remote of the checkout. Such an scm uses the
// local checkout, while any other scm is cloned as usual.
//
// Every remote counts, not only origin: in a fork, origin is usually the fork while a
// manifest describes the upstream repository, another remote of the same checkout.
//
// Only the repository is compared, not the branch: a manifest usually targets a branch such
// as main while its author works on a short lived branch. Two scms of the same repository on
// different branches therefore both use the local checkout, whatever branch it is on. When
// the branch of an scm differs from the one checked out, it is only logged at debug level.
//
// A checkout without any remote can't be matched, so every scm with a URL is cloned. An scm
// without a URL, having nothing to clone, uses the local checkout. Each scm URL is logged once.
func UsesLocalCheckout(handler ScmHandler, directory string) bool {
	scmURL := handler.GetURL()
	key := repositoryKey(scmURL)

	localCheckout.Lock()
	defer localCheckout.Unlock()

	if localCheckout.directory != directory {
		localCheckout.directory = directory
		localCheckout.remotes = remoteKeys(directory)
		localCheckout.branch, localCheckout.branchKnown = currentBranch(directory)
		localCheckout.reported = map[string]bool{}
		localCheckout.reportedBranches = map[string]bool{}

		if len(localCheckout.remotes) == 0 {
			log.Warningf("the local checkout %q has no git remote, so no scm can be matched to it: every scm is cloned, read only. Add the remote the manifest describes to use the local checkout",
				directory)
		}
	}

	if key == "" {
		reportBranch(handler, scmURL, directory)
		return true
	}

	firstReport := !localCheckout.reported[scmURL]
	localCheckout.reported[scmURL] = true

	if remote, found := localCheckout.remotes[key]; found {
		if firstReport {
			log.Infof("scm repository %q uses the local checkout %q, as its remote %q", redact.URL(scmURL), directory, remote)
		}
		reportBranch(handler, scmURL, directory)
		return true
	}

	if firstReport {
		log.Infof("scm repository %q matches no remote of the local checkout %q: it is cloned, read only, and its targets run in dry run",
			redact.URL(scmURL), directory)
	}

	return false
}

// reportBranch logs at debug level, once per scm URL and branch, that an scm using the local
// checkout targets another branch than the one checked out. It is expected while working on a
// short lived branch, so it never warns. The caller holds the localCheckout lock.
func reportBranch(handler ScmHandler, scmURL, directory string) {
	if !localCheckout.branchKnown {
		return
	}

	scmBranch, _, _ := handler.GetBranches()
	if scmBranch == "" || scmBranch == localCheckout.branch {
		return
	}

	reportKey := scmURL + "\x00" + scmBranch
	if localCheckout.reportedBranches[reportKey] {
		return
	}
	localCheckout.reportedBranches[reportKey] = true

	if localCheckout.branch == "" {
		log.Debugf("scm repository %q targets branch %q, but the local checkout %q is not on a branch: its files are used",
			redact.URL(scmURL), scmBranch, directory)
		return
	}

	log.Debugf("scm repository %q targets branch %q, but the local checkout %q is on branch %q: its files are used",
		redact.URL(scmURL), scmBranch, directory, localCheckout.branch)
}

// currentBranch returns the branch checked out in directory, or an empty branch in detached
// HEAD. It reports false when HEAD can't be read.
func currentBranch(directory string) (string, bool) {
	repository, err := git.PlainOpenWithOptions(directory, &git.PlainOpenOptions{DetectDotGit: true, EnableDotGitCommonDir: true})
	if err != nil {
		log.Debugf("opening the local checkout %q: %s", directory, err)
		return "", false
	}

	// HEAD is read without resolving it, so a branch without any commit yet is still found.
	head, err := repository.Reference(plumbing.HEAD, false)
	if err != nil {
		log.Debugf("reading the HEAD of the local checkout %q: %s", directory, err)
		return "", false
	}

	if head.Type() == plumbing.SymbolicReference && head.Target().IsBranch() {
		return head.Target().Short(), true
	}

	return "", true
}

// remoteKeys maps the repository key of each remote of the checkout in directory to the
// remote name. When several remotes share a repository, origin wins, then the first name in
// alphabetical order, so the name logged is stable.
func remoteKeys(directory string) map[string]string {
	remotes, err := (&gitgeneric.GoGit{}).RemoteURLs(directory)
	if err != nil {
		log.Debugf("reading the remotes of the local checkout %q: %s", directory, err)
		return nil
	}

	names := slices.Sorted(maps.Keys(remotes))
	if i := slices.Index(names, "origin"); i > 0 {
		names = append([]string{"origin"}, slices.Delete(names, i, i+1)...)
	}

	keys := map[string]string{}
	for _, name := range names {
		key := repositoryKey(remotes[name])
		if _, found := keys[key]; key != "" && !found {
			keys[key] = name
		}
	}

	return keys
}

// repositoryKey reduces a repository URL to "host/path", so the https and ssh forms of the
// same repository compare equal. It returns an empty string for an empty URL.
func repositoryKey(rawURL string) string {
	key := strings.TrimSpace(rawURL)

	hasScheme := false
	if i := strings.Index(key, "://"); i >= 0 {
		key = key[i+len("://"):]
		hasScheme = true
	}

	host, path, _ := strings.Cut(key, "/")

	// Credentials or an ssh user: "user:password@host" or "git@host:owner/repo".
	if i := strings.LastIndex(host, "@"); i >= 0 {
		host = host[i+1:]
	}

	if h, rest, found := strings.Cut(host, ":"); found {
		if hasScheme {
			// "host:port": the port differs between https and ssh, so it is dropped.
			host = h
		} else {
			// scp-like ssh form "host:owner/repo".
			host = h
			path = strings.TrimPrefix(rest+"/"+path, "/")
			path = strings.TrimSuffix(path, "/")
		}
	}

	path = strings.TrimSuffix(strings.TrimSuffix(path, "/"), ".git")
	path = strings.Trim(path, "/")

	// Azure DevOps spells the same repository "dev.azure.com/org/project/_git/repo" over https
	// and "ssh.dev.azure.com:v3/org/project/repo" over ssh.
	switch strings.ToLower(host) {
	case "ssh.dev.azure.com":
		host = "dev.azure.com"
		path = strings.TrimPrefix(path, "v3/")
	case "dev.azure.com":
		path = strings.Replace(path, "/_git/", "/", 1)
	}

	if host == "" && path == "" {
		return ""
	}

	return strings.ToLower(host + "/" + path)
}
