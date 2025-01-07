package gitgeneric

import (
	"fmt"
	"io"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// ReadFileFromRevision reads a file from a git repository at a given revision.
func ReadFileFromRevision(repoPath, revision, filePath string) ([]byte, error) {

	repo, err := git.PlainOpen(repoPath)

	if err != nil {
		return nil, fmt.Errorf("open git repository at %q: %w", repoPath, err)
	}

	h, err := repo.ResolveRevision(plumbing.Revision(revision))
	if err != nil {
		return nil, fmt.Errorf("resolve revision %q: %w", revision, err)
	}

	if h == nil {
		return nil, fmt.Errorf("revision not found: %s", revision)
	}

	obj, err := repo.Object(plumbing.AnyObject, *h)
	if err != nil {
		return nil, fmt.Errorf("get object %q: %w", *h, err)
	}

	blob, err := resolve(obj, filePath)
	if err != nil {
		return nil, fmt.Errorf("resolve %q: %w", filePath, err)
	}

	r, err := blob.Reader()
	if err != nil {
		return nil, fmt.Errorf("open reader: %w", err)
	}

	defer func() {
		err = r.Close()
	}()

	content, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read content: %w", err)
	}

	return content, nil
}

// resolve blob at given path from obj. obj can be a commit, tag, tree, or blob.
func resolve(obj object.Object, path string) (*object.Blob, error) {
	switch o := obj.(type) {
	case *object.Commit:
		t, err := o.Tree()
		if err != nil {
			return nil, fmt.Errorf("get tree: %w", err)
		}
		return resolve(t, path)
	case *object.Tag:
		target, err := o.Object()
		if err != nil {
			return nil, fmt.Errorf("get object: %w", err)
		}
		return resolve(target, path)
	case *object.Tree:
		file, err := o.File(path)
		if err != nil {
			return nil, fmt.Errorf("get file %q: %w", path, err)
		}
		return &file.Blob, nil
	case *object.Blob:
		return o, nil
	default:
		return nil, object.ErrUnsupportedObject
	}
}
