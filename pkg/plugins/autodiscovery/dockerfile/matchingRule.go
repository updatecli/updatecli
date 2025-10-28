package dockerfile

import (
	"path/filepath"

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

	if len(m) > 0 {
		for _, rule := range m {
			var match bool
			var err error

			// Only check if path rule defined
			if rule.Path != "" {
				if filepath.IsAbs(rule.Path) {
					filePath = filepath.Join(rootDir, filePath)
				}

				match, err = filepath.Match(rule.Path, filePath)
				if err != nil {
					logrus.Errorf("%s - %q", err, rule.Path)
					continue
				}
				ruleResults = append(ruleResults, match)
				if match {
					logrus.Debugf("file path %q matching rule %q", filePath, rule.Path)
				}
			}

			// Only check if service rule defined.
			if len(rule.Archs) > 0 {
				match := false
			archs:
				for _, a := range rule.Archs {
					if a == arch {
						logrus.Debugf("arch %q matching rule %q", arch, a)
						match = true
						break archs
					}
				}
				ruleResults = append(ruleResults, match)
			}

			// Only check if image rule defined.
			if len(rule.Images) > 0 {
				match := false
			images:
				for _, i := range rule.Images {
					if image == i {
						logrus.Debugf("image %q matching rule %q", image, i)
						match = true
						break images
					}
				}
				ruleResults = append(ruleResults, match)
			}

			if len(ruleResults) == 0 {
				return false
			}

			isAllMatching := true
			for i := range ruleResults {
				if !ruleResults[i] {
					isAllMatching = false
					break
				}
			}

			if isAllMatching {
				return true
			}
			ruleResults = []bool{}
		}
	}

	return false
}
