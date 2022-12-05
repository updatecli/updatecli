package dockerfile

import (
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

// MatchingRule allows to specifies rules to identify manifest
type MatchingRule struct {
	// Arch specifies a list of docker image architecture
	Archs []string
	// Path specifies a Dockerfile path pattern, the pattern requires to match all of name, not just a substring.
	Path string
	// Image specifies a list of docker image
	Images []string
}

type MatchingRules []MatchingRule

// isMatchingRule tests that all defined rule are matching and return true if it's the case otherwise return false
func (m MatchingRules) isMatchingRule(rootDir, filePath, image, arch string) bool {
	// Test if the ignore rule are respected
	var ruleResults []bool
	var finaleResult bool

	if len(m) > 0 {
		for _, matchingRule := range m {
			var match bool
			var err error

			// Only check if path rule defined
			if matchingRule.Path != "" && filePath != "" {
				if filepath.IsAbs(matchingRule.Path) {
					filePath = filepath.Join(rootDir, filePath)
				}

				match, err = filepath.Match(matchingRule.Path, filePath)
				if err != nil {
					logrus.Errorf("%s - %q", err, matchingRule.Path)
				}
				ruleResults = append(ruleResults, match)
				if match {
					logrus.Debugf("file path %q matching rule %q", filePath, matchingRule.Path)
				}
			}

			// Only check if service rule defined.
			if len(matchingRule.Archs) > 0 && arch != "" {
				match := false
			archs:
				for _, a := range matchingRule.Archs {
					if a == arch {
						logrus.Debugf("arch %q matching rule %q", arch, a)
						match = true
						break archs
					}
				}
				ruleResults = append(ruleResults, match)
			}

			// Only check if image rule defined.
			if len(matchingRule.Images) > 0 && image != "" {
				match := false
			images:
				for _, i := range matchingRule.Images {
					if strings.HasPrefix(image, i) {
						logrus.Debugf("image %q matching rule %q", image, i)
						match = true
						break images
					}
				}
				ruleResults = append(ruleResults, match)
			}
		}

		if len(ruleResults) == 0 {
			return false
		}

		allMatchingRule := true
		for i := range ruleResults {
			if !ruleResults[i] {
				allMatchingRule = false
				break
			}
		}

		if allMatchingRule {
			finaleResult = true
		}
	}

	return finaleResult
}
