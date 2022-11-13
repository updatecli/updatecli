package dockercompose

import (
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// MatchingRule allows to specifies rules to identify manifest
type MatchingRule struct {
	// Path specifies a Helm chart path pattern, the pattern requires to match all of name, not just a substring.
	Path string
	// Services specifies a list of docker compose services
	Services []string
}

type MatchingRules []MatchingRule

// isMatchingRule tests that all defined rule are matching and return true if it's the case otherwise return false
func (m MatchingRules) isMatchingRule(rootDir string, filePath string, services map[string]service) bool {
	// Test if the ignore rule based on path is respected

	var ruleResults []bool
	var finaleResult bool

	if len(m) > 0 {
		for _, matchingRule := range m {
			var match bool
			var err error

			// Only check if path rule defined
			if matchingRule.Path != "" {
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
			// All services defined in the rule must be matched
			if len(matchingRule.Services) > 0 {
				match := false
				for _, ms := range matchingRule.Services {
					for serviceName := range services {
						if ms == serviceName {
							logrus.Debugf("service %q matching rule %q", filePath, matchingRule.Path)
							match = true
							break
						}
					}
				}
				ruleResults = append(ruleResults, match)

			}
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
