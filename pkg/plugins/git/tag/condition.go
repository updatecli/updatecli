package tag

import (
	"fmt"
	"strings"

	"github.com/olblak/updateCli/pkg/core/scm"
	"github.com/olblak/updateCli/pkg/plugins/git/generic"
	"github.com/sirupsen/logrus"
)

// Condition checks that a git tag exists
func (t *Tag) Condition(source string) (bool, error) {
	if len(t.VersionFilter.Pattern) == 0 {
		t.VersionFilter.Pattern = source
	}

	err := t.Validate()
	if err != nil {
		logrus.Errorln(err)
		return false, err
	}

	tags, err := generic.Tags(t.Path)

	if err != nil {
		logrus.Errorln(err)
		return false, err
	}

	tag, err := t.VersionFilter.Search(tags)
	if err != nil {
		return false, err
	}

	if len(tag) == 0 {
		err = fmt.Errorf("no git tag matching pattern %q, found", t.VersionFilter.Pattern)
		return false, err
	}

	if strings.Compare(tag, t.VersionFilter.Pattern) == 0 {
		logrus.Printf("\u2714 git tag %q matching\n", t.VersionFilter.Pattern)
		return true, nil
	}

	logrus.Printf("\u2717 git tag %q not matching %q\n",
		t.VersionFilter.Pattern,
		tag)

	return false, nil
}

// ConditionFromSCM test if a tag exist from a git repository specific from SCM
func (t *Tag) ConditionFromSCM(source string, scm scm.Scm) (bool, error) {
	path := scm.GetDirectory()

	if len(t.VersionFilter.Pattern) == 0 {
		t.VersionFilter.Pattern = source
	}

	if len(t.Path) > 0 {
		logrus.Warningf("Path is defined and set to %q but is overridden by the scm definition %q",
			t.Path,
			path)
	}
	t.Path = path

	err := t.Validate()
	if err != nil {
		return false, err
	}

	tags, err := generic.Tags(t.Path)

	if err != nil {
		logrus.Errorln(err)
		return false, err
	}

	tag, err := t.VersionFilter.Search(tags)
	if err != nil {
		return false, err
	}

	if tag == t.VersionFilter.Pattern {
		logrus.Printf("\u2714 Git Tag %q matching\n", t.VersionFilter.Pattern)
		return true, nil
	}
	logrus.Printf("\u2717 Git Tag %q not matching %q\n",
		t.VersionFilter.Pattern,
		tag)
	return false, nil
}
