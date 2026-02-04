package woodpecker

import (
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

// MatchingRule allows to specify rules to identify manifests
type MatchingRule struct {
	// Path specifies a Woodpecker workflow path pattern, the pattern requires to match all of name, not just a substring.
	Path string
	// Images specifies a list of docker images
	Images []string
}

// MatchingRules is a slice of MatchingRule
type MatchingRules []MatchingRule

// isMatchingRule tests that all defined rules are matching and returns true if it's the case otherwise returns false
func (m MatchingRules) isMatchingRule(rootDir, filePath, image string) bool {
	if len(m) > 0 {
		for _, matchingRule := range m {
			var ruleResults []bool
			var match bool
			var err error

			// Only check if path rule defined
			if matchingRule.Path != "" && filePath != "" {
				checkPath := filePath
				if filepath.IsAbs(matchingRule.Path) {
					checkPath = filepath.Join(rootDir, filePath)
				}

				match, err = filepath.Match(matchingRule.Path, checkPath)
				if err != nil {
					logrus.Errorf("%s - %q", err, matchingRule.Path)
				}
				ruleResults = append(ruleResults, match)
				if match {
					logrus.Debugf("file path %q matching rule %q", checkPath, matchingRule.Path)
				}
			}

			// Only check if image rule defined.
			if len(matchingRule.Images) > 0 && image != "" {
				match := false
				for _, i := range matchingRule.Images {
					if strings.HasPrefix(image, i) {
						logrus.Debugf("image %q matching rule %q", image, i)
						match = true
						break
					}
				}
				ruleResults = append(ruleResults, match)
			}

			allMatchingRule := true
			for i := range ruleResults {
				if !ruleResults[i] {
					allMatchingRule = false
					break
				}
			}

			if allMatchingRule && len(ruleResults) > 0 {
				return true
			}
		}
	}

	return false
}
