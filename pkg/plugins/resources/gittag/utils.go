package gittag

import (
	"fmt"
	"sort"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
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

	listOptions := &git.ListOptions{
		// An annotated tag is advertised with the hash of its tag object; only its peeled
		// entry carries the commit, which is what local listing and pinned references use.
		PeelingOption: git.AppendPeeled,
	}

	if gt.spec.Username != "" && gt.spec.Password != "" {
		listOptions.Auth = &http.BasicAuth{
			Username: gt.spec.Username,
			Password: gt.spec.Password,
		}
	}

	refs, err := remote.List(listOptions)
	if err != nil {
		return nil, nil, err
	}

	tagsList := make([]string, 0, len(refs))
	peeled := make(map[string]string)
	for _, ref := range refs {
		if !ref.Name().IsTag() {
			continue
		}
		if name, isPeeled := strings.CutSuffix(ref.Name().Short(), "^{}"); isPeeled {
			peeled[name] = ref.Hash().String()
			continue
		}
		results[ref.Name().Short()] = ref.Hash().String()
		tagsList = append(tagsList, ref.Name().Short())
	}
	for name, hash := range peeled {
		if _, found := results[name]; found {
			results[name] = hash
		}
	}

	// Sort the tags list in lexicographical order before returning
	// to align the behavior with `git ls-remote --refs --tags`
	sort.Slice(tagsList, func(i, j int) bool {
		return tagsList[i] < tagsList[j]
	})

	return tagsList, results, nil
}

// listRemoteDirectoryTags lists the tags of a local git repository which were created
// inside the provided age window.
//
// directory is the repository to read from when the spec names neither a URL nor a path.
// published holds every tag of the repository, whatever its age, so that callers can tell
// a tag still cooling down from a version filter matching no tag.
func (gt *GitTag) listRemoteDirectoryTags(directory string, tagAge age.Spec, pathResolver pathresolver.Resolver) (tagsList []string, results map[string]string, published []string, err error) {
	if gt.nativeGitHandler == nil {
		return nil, nil, nil, fmt.Errorf("nativeGitHandler is not initialized")
	}

	gt.directory = directory

	results = make(map[string]string)
	tagsList = make([]string, 0)

	if gt.spec.URL != "" {
		gt.directory, err = gt.clone()
		if err != nil {
			return nil, nil, nil, fmt.Errorf("cloning repository: %w", err)
		}
	}
	if gt.spec.Path != "" {
		gt.directory = pathResolver.JoinManifest(gt.spec.Path)
	}

	if gt.directory == "" {
		return nil, nil, nil, fmt.Errorf("unknown Git working directory. Did you specify one of `URL`, `scmID`, or `spec.path`?")
	}

	refs, err := gt.nativeGitHandler.TagRefs(gt.directory)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("retrieving tag refs: %w", err)
	}

	for _, ref := range refs {
		published = append(published, ref.Name)
		if !tagAge.Matches(ref.When) {
			log.Debugf("ignoring tag %q, dated %s, as outside of the age window", ref.Name, ref.When)
			continue
		}
		results[ref.Name] = ref.Hash
		tagsList = append(tagsList, ref.Name)
	}

	return tagsList, results, published, nil
}
