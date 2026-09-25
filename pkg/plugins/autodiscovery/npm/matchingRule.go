package npm

import (
	"fmt"
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
)

// MatchingRule defines a rule to select npm packages.
// A rule matches when every field it sets matches.
// Each rule must set at least one field.
type MatchingRule struct {
	// "path" defines a package.json path pattern.
	//
	// remark:
	//   * the pattern must match the whole path, not just a substring.
	//   * the pattern follows the Go filepath.Match syntax, such as "*" or "?".
	//
	Path string
	// "packages" defines the npm packages to match, keyed by package name.
	//
	// remark:
	//   * an empty value matches any version.
	//   * otherwise the value is a semantic version constraint, such as ">=1.0.0".
	//   * when the version or the constraint cannot be parsed, the value must equal the version.
	//
	Packages map[string]string
	// "hasversionconstraint" defines whether the package must be declared with a version constraint.
	//
	// example:
	//   * hasversionconstraint: true
	//
	HasVersionConstraint *bool
}

// MatchingRules defines a list of rules.
// The list matches when at least one of its rules matches.
type MatchingRules []MatchingRule

// Validate checks that each matching rule has at least one non-empty field.
// Returns an error if any rule has no valid fields specified.
func (m MatchingRules) Validate() error {
	for i, rule := range m {
		if rule.Path == "" && len(rule.Packages) == 0 && rule.HasVersionConstraint == nil {
			return fmt.Errorf("rule %d has no valid fields (path, packages, or hasVersionConstraint must be specified)", i+1)
		}
	}
	return nil
}

// isMatchingRules checks if a specific file content matches the "only" rule
func (m MatchingRules) isMatchingRules(rootDir, filePath, packageName, packageVersion string) bool {
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

			if len(rule.Packages) > 0 {
				match := false

			outPackage:
				for rulePackageName, rulePackageVersion := range rule.Packages {

					if packageName == rulePackageName {
						if rulePackageVersion == "" {
							match = true
							break outPackage
						}

						v, err := semver.NewVersion(packageVersion)
						if err != nil {
							match = packageVersion == rulePackageVersion
							logrus.Debugf("%q - %s", packageVersion, err)
							break outPackage
						}

						c, err := semver.NewConstraint(rulePackageVersion)
						if err != nil {
							match = packageVersion == rulePackageVersion
							logrus.Debugf("%q %s", err, rulePackageVersion)
							break outPackage
						}

						match = c.Check(v)
						break outPackage
					}
				}
				ruleResults = append(ruleResults, match)
			}

			if rule.HasVersionConstraint != nil {
				isVersionConstraint := isVersionConstraintSpecified(
					packageName,
					packageVersion)

				if *rule.HasVersionConstraint == isVersionConstraint {
					ruleResults = append(ruleResults, isVersionConstraint)
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
		return false
	}

	return false
}
