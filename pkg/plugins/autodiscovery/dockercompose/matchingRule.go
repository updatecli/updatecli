package dockercompose

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

// MatchingRule allows to specifies rules to identify manifest
type MatchingRule struct {
	// Arch specifies a list of docker image architecture
	Archs []string
	// Path specifies a Helm chart path pattern, the pattern requires to match all of name, not just a substring.
	Path string
	// Services specifies a list of docker compose services
	Services []string
	// Image specifies a list of docker image
	Images []string
}

type MatchingRules []MatchingRule

// Validate checks that each matching rule has at least one non-empty field.
// Returns an error if any rule has no valid fields specified.
func (m MatchingRules) Validate() error {
	for i, rule := range m {
		if rule.Path == "" && len(rule.Archs) == 0 && len(rule.Services) == 0 && len(rule.Images) == 0 {
			return fmt.Errorf("rule %d has no valid fields (path, archs, services, or images must be specified)", i+1)
		}
	}
	return nil
}

// isMatchingRule tests that all defined rule are matching and return true if it's the case otherwise return false
func (m MatchingRules) isMatchingRule(rootDir, filePath, service, image, arch string) bool {
	// Test if the ignore rule based on path is respected

	var ruleResults []bool

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
				for _, a := range matchingRule.Archs {
					if a == arch {
						logrus.Debugf("arch %q matching rule %q", service, a)
						match = true
						break
					}
				}
				ruleResults = append(ruleResults, match)
			}

			// Only check if service rule defined.
			if len(matchingRule.Services) > 0 && service != "" {
				match := false
				for _, ms := range matchingRule.Services {
					if ms == service {
						logrus.Debugf("service %q matching rule %q", service, ms)
						match = true
						break
					}
				}
				ruleResults = append(ruleResults, match)
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

			if allMatchingRule {
				return true
			}
		}

	}

	return false
}
