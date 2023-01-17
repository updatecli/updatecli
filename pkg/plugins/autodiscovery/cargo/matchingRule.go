package cargo

import (
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// MatchingRule allows to specifies rules to identify manifest
type MatchingRule struct {
	// Path specifies a Cargo crate path pattern, the pattern requires to match all of name, not just a subpart of the path.
	Path string
}

type MatchingRules []MatchingRule

// isMatchingIgnoreRule checks if a specific file content matches the "ignore" rule
func (m MatchingRules) isMatchingIgnoreRule(rootDir, relativePath string) bool {
	result := false

	if len(m) > 0 {
		for _, rule := range m {
			logrus.Infof("Rule: %q\n", rule.Path)
			logrus.Infof("Relative Found: %q\n", relativePath)
			switch filepath.IsAbs(rule.Path) {
			case true:
				match, err := filepath.Match(rule.Path, filepath.Join(rootDir, relativePath))
				if err != nil {
					logrus.Errorf("%s - %q", err, rule.Path)
				}
				if match {
					result = true
				}
			case false:
				match, err := filepath.Match(rule.Path, relativePath)
				if err != nil {
					logrus.Errorf("%s - %q", err, rule.Path)
				}
				if match {
					result = true
				}
			}
		}
	}

	return result
}

// isMatchingOnlyRule checks if a specific file content matches the "only" rule
func (m MatchingRules) isMatchingOnlyRule(rootDir, relativePath string) bool {
	result := false
	if len(m) > 0 {
		for _, rule := range m {
			switch filepath.IsAbs(rule.Path) {
			case true:
				match, err := filepath.Match(rule.Path, filepath.Join(rootDir, relativePath))
				if err != nil {
					logrus.Errorf("%s - %q", err, rule.Path)
				}
				if match {
					result = true
				}
			case false:
				match, err := filepath.Match(rule.Path, relativePath)
				if err != nil {
					logrus.Errorf("%s - %q", err, rule.Path)
				}
				if match {
					result = true
				}
			}
		}
	}

	return result
}
