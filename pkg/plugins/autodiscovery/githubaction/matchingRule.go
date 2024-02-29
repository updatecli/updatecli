package githubaction

import (
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
)

// MatchingRule allows to specifies rules to identify manifest
type MatchingRule struct {
	// Path specifies a Flux filepath pattern, the pattern requires to match all of name, not just a subpart of the path.
	Path string
	// Actions specifies the list of artifacts to check
	//
	// The key is the artifact name and the value is the artifact version
	//
	// An artifact can be a Helm Chart when used in the context of Helmrelease
	// or an OCIRepository when used in the context of OCIRepository
	//
	// If the value is empty, then the artifact name is enough to match
	// If the value is a valid semver constraint, then the artifact version must match the constraint
	Actions map[string]string
}

type MatchingRules []MatchingRule

// isMatchingRules checks if a specific file content matches the "only" rule
func (m MatchingRules) isMatchingRules(rootDir, filePath, chartName, chartVersion string) bool {
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
				Checks if Actions is matching the policy constraint.
			*/

			if len(rule.Actions) > 0 {
				match := false

			outAction:
				for repository, reference := range rule.Actions {

					if chartName == repository {
						if reference == "" {
							match = true
							break outAction
						}

						v, err := semver.NewVersion(chartVersion)
						if err != nil {
							match = chartVersion == reference
							logrus.Debugf("%q - %s", chartVersion, err)
							break outAction
						}

						c, err := semver.NewConstraint(reference)
						if err != nil {
							match = chartVersion == reference
							logrus.Debugf("%q %s", err, reference)
							break outAction
						}

						match = c.Check(v)
						break outAction
					}
				}
				ruleResults = append(ruleResults, match)
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
		return false
	}

	return false
}
