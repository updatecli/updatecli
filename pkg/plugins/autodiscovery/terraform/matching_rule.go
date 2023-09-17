package terraform

import (
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
)

// MatchingRule allows to specifies rules to identify manifest
type MatchingRule struct {
	// Path specifies a go.mod path pattern, the pattern requires to match all of name, not just a substring.
	Path string
	// Modules specifies a list of module pattern.
	Providers map[string]string
}

type MatchingRules []MatchingRule

// isMatchingRules checks for each matchingRule if parameters are matching rules and then return true or false.
func (m MatchingRules) isMatchingRules(rootDir, filePath, providerName, providerVersion string) bool {
	var ruleResults []bool

	if len(m) > 0 {
		for _, rule := range m {
			/*
			 Check if rule.Path is matching. Path accepts wildcard path
			*/
			if rule.Path != "" {
				if filepath.IsAbs(rule.Path) {
					filePath = filepath.Join(rootDir, filePath)
				}

				match, err := filepath.Match(rule.Path, filePath)
				if err != nil {
					logrus.Errorf("%s - %q", err, rule.Path)
					continue
				}
				ruleResults = append(ruleResults, match)
				if match {
					logrus.Debugf("file path %q matching rule %q", filePath, rule.Path)
				}
			}

			/*
				Checks if provider is matching the provider constraint.
				If both provider constraint is empty and no providerName have been provided then we
				assume the rule is matching

				Otherwise we checks both that version and provider name are matching.
				Version matching uses semantic versioning constraints if possible otherwise
				just compare the version rule and the module version.
			*/

			if len(rule.Providers) > 0 {
				if providerName != "" {
					match := false
					for ruleProviderName, ruleProviderVersion := range rule.Providers {
						if providerName == ruleProviderName {
							if ruleProviderVersion == "" {
								match = true
								break
							}

							v, err := semver.NewVersion(providerVersion)
							if err != nil {
								match = providerVersion == ruleProviderVersion
								logrus.Debugf("%q - %s", providerVersion, err)
								break
							}

							c, err := semver.NewConstraint(ruleProviderVersion)
							if err != nil {
								match = providerVersion == ruleProviderVersion
								logrus.Debugf("%q %s", err, ruleProviderVersion)
								break
							}

							match = c.Check(v)
							break
						}
					}
					ruleResults = append(ruleResults, match)
				}
			}

			/*
				If at least one rule is failing then we return false
			*/
			isAllMatching := true
			for i := range ruleResults {
				if !ruleResults[i] {
					isAllMatching = false
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
