package gittag

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

// Condition checks that a git tag exists
func (gt *GitTag) Condition(source string, scm scm.ScmHandler) (pass bool, message string, err error) {

	if gt.spec.Path != "" && scm != nil {
		logrus.Warningf("Path setting value %q is overriding the scm configuration (value %q)",
			gt.spec.Path,
			scm.GetDirectory())
	}

	if gt.spec.URL != "" && scm != nil {
		logrus.Warningf("URL setting value %q is overriding the scm configuration (value %q)",
			gt.spec.URL,
			scm.GetURL())
	}

	if gt.spec.URL != "" {
		gt.directory, err = gt.clone()
		if err != nil {
			return false, "", fmt.Errorf("cloning git repository: %w", err)
		}

	} else if gt.spec.Path != "" {
		gt.directory = gt.spec.Path
	} else if scm != nil {
		gt.directory = scm.GetDirectory()
	}

	// If source input is empty, then it means that it was disabled by the user with `disablesourceinput: true`
	if source != "" {
		logrus.Infof("Source input value detected: using it as spec.versionfilter.pattern")
		gt.versionFilter.Pattern = source
	}

	err = gt.Validate()
	if err != nil {
		return false, "", err
	}

	if gt.directory == "" {
		return false, "", fmt.Errorf("Unkownn Git working directory. Did you specify one of `URL`, `scmID`, or `spec.path`?")
	}

	tags, err := gt.nativeGitHandler.Tags(gt.directory)
	if err != nil {
		return false, "", err
	}

	if len(tags) == 0 {
		return false, "no git tag found", nil
	}

	gt.foundVersion, err = gt.versionFilter.Search(tags)
	if err != nil {
		return false, "", err
	}
	tag := gt.foundVersion.GetVersion()

	if len(tag) == 0 {
		return false, fmt.Sprintf("no git tag matching pattern %q, found", gt.versionFilter.Pattern), nil
	}

	return true, fmt.Sprintf("git tag matching %q found\n", gt.versionFilter.Pattern), nil

}
