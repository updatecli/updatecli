package gittag

import (
	"fmt"
	"sort"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

// listRemoteURLTags lists all tags from a remote git repository
func (gt *GitTag) listRemoteURLTags() ([]string, map[string]string, error) {
	results := make(map[string]string)

	remote := git.NewRemote(nil, &config.RemoteConfig{
		Name: "origin",
		URLs: []string{
			gt.spec.URL,
		},
	})

	listOptions := &git.ListOptions{}

	if gt.spec.Username != "" && gt.spec.Password != "" {
		listOptions.Auth = &http.BasicAuth{
			Username: gt.spec.Username,
			Password: gt.spec.Password,
		}
	}

	refs, err := remote.List(listOptions)
	if err != nil {
		return nil, nil, fmt.Errorf("listing remote tags: %w", err)
	}

	tagsList := make([]string, 0, len(refs))
	for _, ref := range refs {
		if !ref.Name().IsTag() {
			continue
		}
		results[ref.Name().Short()] = ref.Hash().String()
		tagsList = append(tagsList, ref.Name().Short())
	}

	// Sort the tags list in lexicographical order before returning
	// to align the behavior with `git ls-remote --refs --tags`
	sort.Slice(tagsList, func(i, j int) bool {
		return tagsList[i] < tagsList[j]
	})

	return tagsList, results, nil
}

// listRemoteDirectoryTags lists all tags from a local git repository
func (gt *GitTag) listRemoteDirectoryTags(workingDir string) ([]string, map[string]string, error) {
	if gt.nativeGitHandler == nil {
		return nil, nil, fmt.Errorf("nativeGitHandler is not initialized")
	}

	gt.directory = workingDir

	var err error

	results := make(map[string]string)
	tagsList := make([]string, 0)

	if gt.spec.URL != "" {
		gt.directory, err = gt.clone()
		if err != nil {
			return nil, nil, fmt.Errorf("cloning repository: %w", err)
		}
	}
	if gt.spec.Path != "" {
		gt.directory = gt.spec.Path
	}

	if gt.directory == "" {
		return nil, nil, fmt.Errorf("unknown Git working directory. Did you specify one of `URL`, `scmID`, or `spec.path`?")
	}

	refs, err := gt.nativeGitHandler.TagRefs(gt.directory)
	if err != nil {
		return nil, nil, fmt.Errorf("retrieving tag refs: %w", err)
	}

	for _, ref := range refs {
		results[ref.Name] = ref.Hash
		tagsList = append(tagsList, ref.Name)
	}

	return tagsList, results, nil
}
