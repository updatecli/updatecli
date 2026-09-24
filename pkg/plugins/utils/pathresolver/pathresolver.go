package pathresolver

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

// Resolver tells a resource where its relative paths resolve from and, when Updatecli
// works from an SCM checkout, the boundary those paths must stay within.
//
// Keeping the two apart lets a manifest resolve paths against its own directory while
// paths inside an SCM checkout stay contained.
type Resolver struct {
	// BaseDir is the directory relative paths resolve against.
	// An empty value means the process working directory.
	BaseDir string
	// Boundary is the directory a resolved path must stay within.
	// An empty value means there is no boundary, which is the case for a local run
	// without an SCM checkout.
	Boundary string
	// ManifestDir is where relative paths resolve from when the SCM checkout does not
	// apply to them: the directory of the manifest, or the process working directory
	// when empty. It is set with or without an SCM.
	ManifestDir string
}

// ScmDirectoryGetter is the part of an scm handler a path resolver needs: the directory
// where the repository was checked out. It is declared here instead of imported from the
// pipeline so that this package has no internal dependencies and every resource can use it.
type ScmDirectoryGetter interface {
	GetDirectory() (directory string)
}

// New builds the path resolver handed to a resource.
//
// With an scm, the checkout directory is both where relative paths resolve from and the
// boundary they must stay within. Without one there is no boundary to enforce, and paths
// resolve from baseDir: the directory of the manifest that declared them, or the process
// working directory when baseDir is empty.
func New(scmHandler ScmDirectoryGetter, baseDir string) Resolver {
	if scmHandler == nil {
		return Resolver{BaseDir: baseDir, ManifestDir: baseDir}
	}

	scmDirectory := scmHandler.GetDirectory()

	return Resolver{BaseDir: scmDirectory, Boundary: scmDirectory, ManifestDir: baseDir}
}

// Resolve turns a user provided path into the path Updatecli must read from or write to.
//
// http:// and https:// locations are returned unchanged, since they are fetched over the
// network.
//
// An absolute path is returned unchanged when there is no boundary, and rejected otherwise.
// Within an SCM checkout, a path that resolves outside of it almost always means an upstream
// value (such as a {{ source }} output) was injected into a source, condition or target path.
// The pipeline must then fail instead of reading from or writing to a location an attacker
// chose, so the path is rejected, not clamped. See GHSA-hj4x-hm4v-7wpw.
func (r Resolver) Resolve(path string) (string, error) {
	resolvedPath := r.Join(path)

	if r.Boundary == "" || isRemoteLocation(path) {
		return resolvedPath, nil
	}

	if filepath.IsAbs(path) {
		return "", fmt.Errorf(
			"absolute path %q is not allowed: files must stay within the working directory %q",
			path, r.Boundary)
	}

	relPath, err := filepath.Rel(r.Boundary, resolvedPath)
	if err != nil {
		return "", fmt.Errorf(
			"unable to verify that path %q stays within the working directory %q: %w",
			path, r.Boundary, err)
	}

	if relPath == ".." || strings.HasPrefix(relPath, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf(
			"path %q escapes the working directory %q",
			path, r.Boundary)
	}

	return resolvedPath, nil
}

// ResolveAll resolves every path of a list, such as the "files" attribute shared by the
// file based resources. It returns on the first path it cannot resolve.
func (r Resolver) ResolveAll(paths []string) ([]string, error) {
	resolvedPaths := make([]string, len(paths))

	for i := range paths {
		resolvedPath, err := r.Resolve(paths[i])
		if err != nil {
			return nil, err
		}

		resolvedPaths[i] = resolvedPath
	}

	return resolvedPaths, nil
}

// Join resolves a path against the base directory without enforcing the boundary.
//
// It is meant for locations Updatecli only uses to find something, such as a git
// repository, a chart directory or the working directory of a shell command. Those can
// legitimately sit outside of an SCM checkout. Files Updatecli reads from or writes to go
// through Resolve instead.
//
// It follows filepath.Join semantics, so an empty path resolves to the base directory
// itself. A caller for which an empty value means "unset" must test for it before calling
// Join.
func (r Resolver) Join(path string) string {
	if r.BaseDir == "" || isRemoteLocation(path) || filepath.IsAbs(path) {
		return path
	}

	return filepath.Join(r.BaseDir, path)
}

// JoinRooted is like Join, except that within an SCM checkout an absolute path is rooted at
// the checkout: "/Dockerfile" becomes "<checkout>/Dockerfile".
//
// It is for the resources that always joined their path under the checkout: an absolute
// path then can neither read a host file nor change which file an existing manifest reads.
func (r Resolver) JoinRooted(path string) string {
	if r.Boundary == "" || isRemoteLocation(path) || !filepath.IsAbs(path) {
		return r.Join(path)
	}

	return filepath.Join(r.BaseDir, path)
}

// JoinManifest resolves a path against the manifest directory, ignoring any SCM checkout.
//
// It is meant for settings documented as overriding the scm, such as the "path" of the git
// resources. They resolve from the manifest directory, or from the process working
// directory by default.
func (r Resolver) JoinManifest(path string) string {
	return Resolver{BaseDir: r.ManifestDir}.Join(path)
}

// Dir returns the directory relative paths resolve against, falling back to the process
// working directory when the resolver does not define one.
func (r Resolver) Dir() string {
	if r.BaseDir != "" {
		return r.BaseDir
	}

	workingDirectory, err := os.Getwd()
	if err != nil {
		logrus.Debugln("fail getting current working directory")
		return "."
	}

	return workingDirectory
}

// RepositoryDir returns the git repository a git resource reads from when it names neither
// a url nor a path: the SCM checkout, or else the process working directory.
//
// Unlike Dir, it never falls back to the manifest directory. Git opens a repository from its
// root only, and a manifest usually sits deeper in the repository, such as in updatecli.d.
func (r Resolver) RepositoryDir() string {
	return Resolver{BaseDir: r.Boundary}.Dir()
}

// isRemoteLocation reports whether a path is fetched over the network rather than read
// from the filesystem.
func isRemoteLocation(path string) bool {
	return strings.HasPrefix(path, "https://") || strings.HasPrefix(path, "http://")
}
