package gittag

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/git"
	"github.com/updatecli/updatecli/pkg/plugins/git/generic"
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

	err = t.VersionFilter.Search(tags)
	if err != nil {
		return false, err
	}
	existingTag := t.foundVersion.ParsedVersion

	// A matching git tag has been found
	if len(existingTag) != 0 {
		logrus.Printf("%s git tag %q already exist, nothing else todo", result.SUCCESS, existingTag)
		return changed, nil
	}

	newTag := t.VersionFilter.Pattern

	logrus.Printf("%s git tag %q not found, will create it", result.ATTENTION, newTag)

	if dryRun {
		return changed, err
	}

	changed, err = generic.NewTag(newTag, t.Message, t.Path)

	if err != nil {
		return changed, err
	}
	logrus.Printf("%s git tag %q created", result.ATTENTION, newTag)

	scm := git.Git{
		Directory: t.Path,
	}

	err = scm.PushTag(newTag)

	if err != nil {
		logrus.Errorf("Git push tag error: %s", err)
		return changed, err
	}

	logrus.Printf("%s git tag %q pushed", result.ATTENTION, newTag)

	return changed, err

}

// TargetFromSCM create and push a git tag based on the SCM configuration
func (t *Tag) TargetFromSCM(source string, scm scm.ScmHandler, dryRun bool) (changed bool, files []string, message string, err error) {

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

	err = t.VersionFilter.Search(tags)
	if err != nil {
		return changed, files, message, err
	}
	existingTag := t.foundVersion.ParsedVersion

	// A matching git tag has been found
	if len(existingTag) != 0 {
		logrus.Printf("%s git tag %q already exist, nothing else todo",
			result.SUCCESS,
			existingTag)
		return changed, files, message, err
	}

	newTag := t.VersionFilter.Pattern

	logrus.Printf("%s git tag %q not found, creating it", result.ATTENTION, newTag)

	if dryRun {
		return changed, files, message, err
	}

	changed, err = generic.NewTag(newTag, t.Message, path)
	if err != nil {
		return changed, files, message, err
	}
	logrus.Printf("%s git tag %q created", result.ATTENTION, newTag)

	err = scm.PushTag(newTag)

	if err != nil {
		logrus.Errorf("Git push tag error: %s", err)
		return changed, files, message, err
	}

	logrus.Printf("%s git tag %q pushed", result.ATTENTION, newTag)

	message = fmt.Sprintf("Git tag %q pushed", newTag)

	return changed, files, message, err
}
