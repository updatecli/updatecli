package language

import (
	"path"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/tmp"
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

	tags := []*semver.Version{}

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

		releaseDate, err := time.Parse(time.RFC3339, commit.Committer.When.Format(time.RFC3339))
		if err != nil {
			logrus.Debugf("ignoring version %q from repository due to invalid release date format: %q\n", tagName, err)
			return nil
		}

		if l.Spec.Age.Minimum != "" && l.Spec.Age.IsOlderThan(releaseDate, nil) {
			logrus.Debugf("ignoring version %q from repository because its age is below %q (released on %s)\n", tagName, l.Spec.Age.Minimum, releaseDate)
			return nil
		}

		if l.Spec.Age.Maximum != "" && l.Spec.Age.IsNewerThan(releaseDate, nil) {
			logrus.Debugf("ignoring version %q from repository because its age is above %q (released on %s)\n", tagName, l.Spec.Age.Maximum, releaseDate)
			return nil
		}

		version, err := semver.StrictNewVersion(strings.TrimPrefix(tagName, "go"))
		if err != nil {
			logrus.Debugf("ignoring version %q from repository due to invalid semantic version format: %q\n", tagName, err)
			return nil
		}

		tags = append(tags, version)

		return nil
	})

	if err != nil {
		return nil, err
	}

	sort.Sort(semver.Collection(tags))

	versions := []string{}
	for _, tag := range tags {
		versions = append(versions, tag.Original())
	}

	l.Version, err = l.versionFilter.Search(versions)
	if err != nil {
		return nil, err
	}

	return versions, nil
}
