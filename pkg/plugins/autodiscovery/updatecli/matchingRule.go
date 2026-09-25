package updatecli

import (
	"fmt"
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
)

// MatchingRule defines a rule to select Updatecli policies.
// A rule matches when every field it sets matches.
// Each rule must set at least one field.
type MatchingRule struct {
	// "path" defines an Updatecli compose file path pattern.
	//
	// remark:
	//   * the pattern must match the whole path, not just a substring.
	//   * the pattern follows the Go filepath.Match syntax, such as "*" or "?".
	//
	Path string
	// "policies" defines the policies to match, keyed by policy name.
	//
	// remark:
	//   * the policy name is the OCI reference without its tag or digest.
	//   * an empty value matches any version.
	//   * otherwise the value is a semantic version constraint, such as ">=1.0.0".
	//   * when the version or the constraint cannot be parsed, the value must equal the version.
	//
	Policies map[string]string
}

// MatchingRules defines a list of rules.
// The list matches when at least one of its rules matches.
type MatchingRules []MatchingRule

// Validate checks that each matching rule has at least one non-empty field.
// Returns an error if any rule has no valid fields specified.
func (m MatchingRules) Validate() error {
	for i, rule := range m {
		if rule.Path == "" && len(rule.Policies) == 0 {
			return fmt.Errorf("rule %d has no valid fields (path or policies must be specified)", i+1)
		}
	}
	return nil
}

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
