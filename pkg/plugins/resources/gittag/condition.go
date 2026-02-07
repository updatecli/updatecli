package gittag

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

// Condition checks that a git tag exists
func (gt *GitTag) Condition(source string, scm scm.ScmHandler) (pass bool, message string, err error) {

	err = gt.Validate()
	if err != nil {
		return false, "", err
	}

	if gt.spec.URL != "" && scm != nil {
		logrus.Warningf("URL setting value %q is overriding the scm configuration (value %q)",
			gt.spec.URL,
			scm.GetURL())
	}

	var tags map[string]string
	var tagsList []string

	if gt.spec.URL != "" {
		tagsList, tags, err = gt.listRemoteURLTags()
		if err != nil {
			return false, "", fmt.Errorf("listing remote tags: %w", err)
		}
	} else {
		if scm != nil {
			gt.directory = scm.GetDirectory()
		}

		if gt.spec.Path != "" {
			gt.directory = gt.spec.Path
		}

		if gt.spec.Path != "" && scm != nil {
			logrus.Warningf("Path setting value %q is overriding the scm configuration (value %q)",
				gt.spec.Path,
				scm.GetDirectory())
		}

		if gt.directory == "" {
			return false, "", fmt.Errorf("unkownn Git working directory. Did you specify one of `URL`, `scmID`, or `spec.path`?")
		}
		if gt.spec.Path != "" {
			gt.directory = gt.spec.Path
		}
		tagsList, tags, err = gt.listRemoteDirectoryTags(gt.directory)
		if err != nil {
			return false, "", fmt.Errorf("listing local tags: %w", err)
		}
	}

	// If source input is empty, then it means that it was disabled by the user with `disablesourceinput: true`
	if source != "" {
		logrus.Infof("Source input value detected: using it as spec.versionfilter.pattern")
		gt.versionFilter.Pattern = source
	}

	if len(tagsList) == 0 {
		return false, "no tags found", nil
	}

	// If spec.Tag is provided, use it for exact match checking (priority over source)
	// If source is provided and spec.Tag is empty, use source for exact match
	// Otherwise, fall back to versionFilter pattern matching
	tag := source
	if gt.spec.Tag != "" {
		tag = gt.spec.Tag
	}

	// If tag is specified (either via spec.Tag or source), check for exact match
	if tag != "" {

		if _, ok := tags[tag]; ok {
			return true, fmt.Sprintf("git tag %q found", tag), nil
		}

		// Tag not found
		return false, fmt.Sprintf("no git tag found matching %q", tag), nil
	}

	// Fall back to versionFilter pattern matching (existing behavior)
	gt.foundVersion, err = gt.versionFilter.Search(tagsList)
	if err != nil {
		return false, "", err
	}
	foundTag := gt.foundVersion.GetVersion()

	if len(foundTag) == 0 {
		return false, fmt.Sprintf("no git tag matching pattern %q, found", gt.versionFilter.Pattern), nil
	}

	return true, fmt.Sprintf("git tag matching %q found\n", gt.versionFilter.Pattern), nil
}
