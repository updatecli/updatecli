package updatecli

import (
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
)

// MatchingRule allows to specifies rules to identify manifest
type MatchingRule struct {
	// Path specifies a Updatecli compose filepath pattern, the pattern requires to match all of name, not just a subpart of the path.
	Path string
	// Policies specifies a Updatecli policy
	Policies map[string]string
}

type MatchingRules []MatchingRule

// isMatchingRules checks if a specific file content matches the "only" rule
func (m MatchingRules) isMatchingRules(rootDir, filePath, policyName, policyVersion string) bool {
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
				Checks if policy is matching the policy constraint.
			*/

			if len(rule.Policies) > 0 {
				if policyName != "" {
					match := false
				outPolicy:
					for rulePolicyName, rulePolicyVersion := range rule.Policies {
						if policyName == rulePolicyName {
							if rulePolicyVersion == "" {
								match = true
								break outPolicy
							}

							v, err := semver.NewVersion(policyVersion)
							if err != nil {
								match = policyVersion == rulePolicyVersion
								logrus.Debugf("%q - %s", policyVersion, err)
								break outPolicy
							}

							c, err := semver.NewConstraint(rulePolicyVersion)
							if err != nil {
								match = policyVersion == rulePolicyVersion
								logrus.Debugf("%q %s", err, rulePolicyVersion)
								break outPolicy
							}

							match = c.Check(v)
							break outPolicy
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
