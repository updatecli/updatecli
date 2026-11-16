package language

import (
	"errors"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/core/tmp"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
)

const (
	// GolangGitRepository is the default git repository for Golang
	GolangGitRepository = "https://github.com/golang/go.git"
)

type TagInfo struct {
	Name          string
	Hash          string
	CommitterWhen time.Time
}

// getTagsFromRepository returns tags with their resolved commit date.
// To avoid fetching the entire git history, it only fetches the latest commit for each tag and uses its commit date as the tag release date.
// To improve performance across Updatecli runs, it also clones the repository locally on disk.
func (l *Language) getTagsFromRepository() ([]string, error) {

	workingDir := path.Join(tmp.Directory, "github", "golang", "go")

	repo, err := git.PlainClone(workingDir, false, &git.CloneOptions{
		URL:   GolangGitRepository,
		Depth: 1,
		Tags:  git.AllTags,
	})

	if err != nil {
		if err != git.ErrRepositoryAlreadyExists {
			return nil, err
		}

		repo, err = git.PlainOpen(workingDir)
		if err != nil {
			return nil, err
		}
	}

	err = repo.Fetch(&git.FetchOptions{
		Tags:  git.AllTags,
		Force: true,
		Depth: 1,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return nil, err
	}

	refs, err := repo.Tags()
	if err != nil {
		return nil, err
	}

	tags := []datedVersion{}

	err = refs.ForEach(func(ref *plumbing.Reference) error {
		if !ref.Name().IsTag() {
			return nil
		}

		tagName := ref.Name().Short()
		commitHash := ref.Hash()

		// Annotated tag: ref hash points to a tag object; peel to target commit.
		if tagObj, err := repo.TagObject(ref.Hash()); err == nil {
			commitHash = tagObj.Target
		}

		commit, err := repo.CommitObject(commitHash)
		if err != nil {
			// skip tags that don't resolve to commits.
			return nil
		}

		version, err := semver.StrictNewVersion(strings.TrimPrefix(tagName, "go"))
		if err != nil {
			log.Debugf("ignoring version %q from repository due to invalid semantic version format: %q\n", tagName, err)
			return nil
		}

		tags = append(tags, datedVersion{version: version, releaseDate: commit.Committer.When})

		return nil
	})

	if err != nil {
		return nil, err
	}

	return l.searchDatedVersions(tags)
}

// datedVersion is a Golang version with the date of the commit it was tagged on.
type datedVersion struct {
	version     *semver.Version
	releaseDate time.Time
}

// searchDatedVersions sets the version matching both the version filter and the age filter,
// and returns the versions matching the age filter.
//
// When the age filter discards every version matching the version filter, it returns these
// versions along with age.ErrNoVersionMatchingAge, so that callers can skip rather than fail.
func (l *Language) searchDatedVersions(tags []datedVersion) ([]string, error) {
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].version.LessThan(tags[j].version)
	})

	// published holds every version, versions the ones within the age window
	published := []string{}
	versions := []string{}
	for _, tag := range tags {
		published = append(published, tag.version.Original())

		if !l.Spec.Age.Matches(tag.releaseDate) {
			log.Debugf("ignoring version %q from repository because it doesn't match the age filter (released on %s)\n", tag.version.Original(), tag.releaseDate)
			continue
		}

		versions = append(versions, tag.version.Original())
	}

	var err error
	l.Version, err = age.Search(&l.versionFilter, published, versions)
	if err != nil {
		if errors.Is(err, age.ErrNoVersionMatchingAge) {
			return versions, err
		}
		return nil, err
	}

	return versions, nil
}
