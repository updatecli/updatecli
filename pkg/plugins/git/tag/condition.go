package tag

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/scm"
	"github.com/updatecli/updatecli/pkg/plugins/git/generic"
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

	err = t.VersionFilter.Search(tags)
	if err != nil {
		return false, err
	}
	tag := t.foundVersion.ParsedVersion

	if len(tag) == 0 {
		err = fmt.Errorf("no git tag matching pattern %q, found", t.VersionFilter.Pattern)
		return false, err
	}

	if strings.Compare(tag, t.VersionFilter.Pattern) == 0 {
		logrus.Printf("%s git tag %q matching\n", result.SUCCESS, t.VersionFilter.Pattern)
		return true, nil
	}

	logrus.Printf("%s git tag %q not matching %q\n",
		result.FAILURE,
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

	err = t.VersionFilter.Search(tags)
	if err != nil {
		return false, err
	}
	tag := t.foundVersion.ParsedVersion

	if tag == t.VersionFilter.Pattern {
		logrus.Printf("%s Git Tag %q matching\n", result.SUCCESS, t.VersionFilter.Pattern)
		return true, nil
	}
	logrus.Printf("%s Git Tag %q not matching %q\n",
		result.FAILURE,
		t.VersionFilter.Pattern,
		tag)
	return false, nil
}
