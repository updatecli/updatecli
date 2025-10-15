package npm

import (
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
)

// MatchingRule allows to specifies rules to identify manifest
type MatchingRule struct {
	// Path specifies a package.json path pattern, the pattern requires to match all of name, not just a substring.
	Path string
	// Packages specifies the list of NPM packages to check
	Packages map[string]string
	// VersionConstraint indicates either the matching rule should match any version constraint or not.
	VersionConstraint bool
}

type MatchingRules []MatchingRule

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

			if rule.VersionConstraint {
				isVersionConstraint := isVersionConstraintSpecified(
					packageName,
					packageVersion)

				ruleResults = append(ruleResults, isVersionConstraint)
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
