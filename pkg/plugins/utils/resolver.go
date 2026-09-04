package utils

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
// Those two ideas used to be a single "working directory" argument, which is why a
// relative path could only ever be understood as "relative to wherever updatecli was
// started from". Keeping them apart is what allows a manifest to resolve its own paths
// against its own directory without weakening the containment guarantee.
type Resolver struct {
	// BaseDir is the directory relative paths resolve against.
	// An empty value means the process working directory.
	BaseDir string
	// Boundary is the directory a resolved path must stay within.
	// An empty value means there is no boundary, which is the case for a local run
	// without an SCM checkout.
	Boundary string
}

// ScmDirectoryGetter is the only thing a resolver needs from an scm handler: where the
// repository was checked out. It is declared here rather than imported from the pipeline
// so this package stays a leaf, which is what lets every resource depend on it.
type ScmDirectoryGetter interface {
	GetDirectory() (directory string)
}

// NewResolver builds the path resolver handed to a resource.
//
// With an scm, the checkout directory is both where relative paths resolve from and the
// boundary they must stay within. Without one there is no boundary to enforce, and paths
// resolve from baseDir: the directory of the manifest that declared them, or the process
// working directory when baseDir is empty.
func NewResolver(scmHandler ScmDirectoryGetter, baseDir string) Resolver {
	if scmHandler == nil {
		return Resolver{BaseDir: baseDir}
	}

	scmDirectory := scmHandler.GetDirectory()

	return Resolver{BaseDir: scmDirectory, Boundary: scmDirectory}
}

// Resolve turns a user provided path into the path Updatecli must read from or write to.
//
// http:// and https:// locations are returned unchanged, they are fetched over the network
// rather than read from disk.
//
// An absolute path is returned unchanged when there is no boundary, and rejected otherwise.
// Rejecting instead of silently clamping is deliberate: within an SCM checkout, a path
// resolving outside of it almost always means an upstream value (such as a {{ source }}
// output) has been injected into a source/condition/target path, and the pipeline must fail
// rather than read from or write to an attacker chosen location. See GHSA-hj4x-hm4v-7wpw.
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
// It is meant for the locations Updatecli only needs in order to find something — a git
// repository, a chart directory, the working directory of a shell command — as opposed to
// the files it reads from or writes to, which go through Resolve. Those locations
// legitimately sit outside of an SCM checkout.
//
// It follows filepath.Join semantics, which means an empty path resolves to the base
// directory itself rather than staying empty. A caller for which an empty value means
// "unset" rather than "here" must therefore test for it before calling Join.
func (r Resolver) Join(path string) string {
	if r.BaseDir == "" || isRemoteLocation(path) || filepath.IsAbs(path) {
		return path
	}

	return filepath.Join(r.BaseDir, path)
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

// isRemoteLocation reports whether a path is fetched over the network rather than read
// from the filesystem.
func isRemoteLocation(path string) bool {
	return strings.HasPrefix(path, "https://") || strings.HasPrefix(path, "http://")
}
