package tag

import (
	"fmt"

	"github.com/olblak/updateCli/pkg/core/scm"
	"github.com/olblak/updateCli/pkg/plugins/git"
	"github.com/olblak/updateCli/pkg/plugins/git/generic"
	"github.com/sirupsen/logrus"
)

// Target create a tag if needed from a local git repository, without pushing the tag
func (t *Tag) Target(source string, dryRun bool) (changed bool, err error) {
	if len(t.VersionFilter.Pattern) == 0 {
		t.VersionFilter.Pattern = source
	}

	if len(t.Path) == 0 {
		logrus.Errorf("At least path settings required")
	}

	err = t.Validate()

	if err != nil {
		logrus.Errorln(err)
		return changed, err
	}

	tags, err := generic.Tags(t.Path)

	if err != nil {
		logrus.Errorln(err)
		return false, err
	}

	existingTag, err := t.VersionFilter.Search(tags)
	if err != nil {
		return false, err
	}

	// A matching git tag has been found
	if len(existingTag) != 0 {
		logrus.Printf("\u2714 git tag %q already exist, nothing else todo", existingTag)
		return changed, nil
	}

	newTag := t.VersionFilter.Pattern

	logrus.Printf("\u2714 git tag %q not found, will create it", newTag)

	if dryRun {
		return changed, err
	}

	changed, err = generic.NewTag(newTag, t.Message, t.Path)

	if err != nil {
		return changed, err
	}
	logrus.Printf("\u2714 git tag %q created", newTag)

	scm := git.Git{
		Directory: t.Path,
	}

	err = scm.PushTag(newTag)

	if err != nil {
		logrus.Errorf("Git push tag error: %s", err)
		return changed, err
	}

	logrus.Printf("\u2714 git tag %q pushed", newTag)

	return changed, err

}

// TargetFromSCM create and push a git tag based on the SCM configuration
func (t *Tag) TargetFromSCM(source string, scm scm.Scm, dryRun bool) (changed bool, files []string, message string, err error) {

	if len(t.VersionFilter.Pattern) == 0 {
		t.VersionFilter.Pattern = source
	}

	err = t.Validate()

	if err != nil {
		logrus.Errorln(err)
		return changed, files, message, err
	}

	if len(t.Path) > 0 {
		logrus.Warningf("Path setting value %q ignored as it conflicts with %q from scm configuration",
			t.Path,
			scm.GetDirectory())
	}
	path := scm.GetDirectory()

	tags, err := generic.Tags(path)

	if err != nil {
		logrus.Errorln(err)
		return changed, files, message, err
	}

	existingTag, err := t.VersionFilter.Search(tags)
	if err != nil {
		return changed, files, message, err
	}

	// A matching git tag has been found
	if len(existingTag) != 0 {
		logrus.Printf("\u2714 git tag %q already exist, nothing else todo",
			existingTag)
		return changed, files, message, err
	}

	newTag := t.VersionFilter.Pattern

	logrus.Printf("\u2714 git tag %q not found, creating it", newTag)

	if dryRun {
		return changed, files, message, err
	}

	changed, err = generic.NewTag(newTag, t.Message, path)
	if err != nil {
		return changed, files, message, err
	}
	logrus.Printf("\u2714 git tag %q created", newTag)

	err = scm.PushTag(newTag)

	if err != nil {
		logrus.Errorf("Git push tag error: %s", err)
		return changed, files, message, err
	}

	logrus.Printf("\u2714 git tag %q pushed", newTag)

	message = fmt.Sprintf("Git tag %q pushed", newTag)

	return changed, files, message, err
}
