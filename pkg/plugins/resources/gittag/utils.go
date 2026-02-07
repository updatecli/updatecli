package gittag

import (
	"fmt"
	"sort"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
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

	refs, err := remote.List(&git.ListOptions{})
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

	// To align the behavior with `git ls-remote --refs --tags`
	// sort the tags list in lexicographical order before returning
	sort.Slice(tagsList, func(i, j int) bool {
		return tagsList[i] < tagsList[j]
	})

	return tagsList, results, nil
}

// listRemoteDirectoryTags lists all tags from a local git repository
func (gt *GitTag) listRemoteDirectoryTags(path string) ([]string, map[string]string, error) {

	results := make(map[string]string)
	tagsList := make([]string, 0)

	refs, err := gt.nativeGitHandler.TagRefs(path)
	if err != nil {
		return nil, nil, fmt.Errorf("retrieving tag refs: %w", err)
	}

	for _, ref := range refs {
		results[ref.Name] = ref.Hash
		tagsList = append(tagsList, ref.Name)
	}

	//// To align the behavior with `git ls-remote --refs --tags`
	//// sort the tags list in lexicographical order before returning
	//sort.Slice(tagsList, func(i, j int) bool {
	//	return tagsList[i] < tagsList[j]
	//})

	return tagsList, results, nil
}
