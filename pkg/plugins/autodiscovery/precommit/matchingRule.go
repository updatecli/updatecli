package precommit

import (
	"fmt"
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
)

// MatchingRule defines a rule to select pre-commit hook repositories.
// A rule matches when every field it sets matches.
// Each rule must set at least one field.
type MatchingRule struct {
	// "path" defines a ".pre-commit-config.yaml" path pattern.
	//
	// remark:
	//   * the pattern must match the whole path, not just a substring.
	//   * the pattern follows the Go filepath.Match syntax, such as "*" or "?".
	//
	Path string
	// "repos" defines the hook repositories to match, keyed by repository URL.
	//
	// remark:
	//   * an empty value matches any revision.
	//   * otherwise the value is a semantic version constraint, such as ">=1.0.0".
	//   * when the revision or the constraint cannot be parsed, the value must equal the revision.
	//
	Repos map[string]string
}

// MatchingRules defines a list of rules.
// The list matches when at least one of its rules matches.
type MatchingRules []MatchingRule

// Validate checks that each matching rule has at least one non-empty field.
// Returns an error if any rule has no valid fields specified.
func (m MatchingRules) Validate() error {
	for i, rule := range m {
		if rule.Path == "" && len(rule.Repos) == 0 {
			return fmt.Errorf("rule %d has no valid fields (path or repos must be specified)", i+1)
		}
	}
	return nil
}

// isMatchingRules checks if a specific file content matches the "only" rule
func (m MatchingRules) isMatchingRules(rootDir, filePath, repoName, repoVersion string) bool {
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

			if len(rule.Repos) > 0 {
				match := false

			outPackage:
				for rulePackageName, rulePackageVersion := range rule.Repos {

					if repoName == rulePackageName {
						if rulePackageVersion == "" {
							match = true
							break outPackage
						}

						v, err := semver.NewVersion(repoVersion)
						if err != nil {
							match = repoVersion == rulePackageVersion
							logrus.Debugf("%q - %s", repoVersion, err)
							break outPackage
						}

						c, err := semver.NewConstraint(rulePackageVersion)
						if err != nil {
							match = repoVersion == rulePackageVersion
							logrus.Debugf("%q %s", err, rulePackageVersion)
							break outPackage
						}

						match = c.Check(v)
						break outPackage
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
