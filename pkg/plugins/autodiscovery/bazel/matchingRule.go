package bazel

import (
	"fmt"
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
)

// MatchingRule defines a rule to select Bazel modules.
// A rule matches when every field it sets matches.
// Each rule must set at least one field.
type MatchingRule struct {
	// "path" defines a "MODULE.bazel" path pattern.
	//
	// remark:
	//   * the pattern must match the whole path, not just a substring.
	//   * the pattern follows the Go filepath.Match syntax, such as "*" or "?".
	//
	Path string
	// "modules" defines the Bazel modules to match, keyed by module name as written in "MODULE.bazel".
	//
	// remark:
	//   * an empty value matches any version.
	//   * otherwise the value is a semantic version constraint, such as ">=1.0.0".
	//   * when the version or the constraint cannot be parsed, the value must equal the version.
	//   * a module name cannot be empty.
	//
	// example:
	//   ```
	//   - modules:
	//       # match any version of this module
	//       rules_go:
	//       # match only the versions of this module satisfying the constraint
	//       gazelle: "1.x"
	//   ```
	//
	Modules map[string]string
}

// MatchingRules defines a list of rules.
// The list matches when at least one of its rules matches.
type MatchingRules []MatchingRule

// Validate checks that each matching rule has at least one non-empty field.
// Returns an error if any rule has no valid fields specified.
// Also validates that Modules map doesn't contain empty keys (module names).
func (m MatchingRules) Validate() error {
	for i, rule := range m {
		if rule.Path == "" && len(rule.Modules) == 0 {
			return fmt.Errorf("rule %d has no valid fields (path or modules must be specified)", i+1)
		}
		// Validate that Modules map doesn't contain empty keys
		if len(rule.Modules) > 0 {
			for moduleName := range rule.Modules {
				if moduleName == "" {
					return fmt.Errorf("rule %d contains empty module name in modules map", i+1)
				}
			}
		}
	}
	return nil
}

// isMatchingRules checks for each matchingRule if parameters are matching rules and then return true or false.
func (m MatchingRules) isMatchingRules(rootDir, filePath, moduleName, moduleVersion string) bool {
	var ruleResults []bool

	if len(m) > 0 {
		for _, rule := range m {
			// Check if rule.Path is matching. Path accepts wildcard path
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

			// Checks if module is matching the module constraint.
			// If both module constraint is empty and no moduleName have been provided then we
			// assume the rule is matching
			//
			// Otherwise we checks both that version and module name are matching.
			// Version matching uses semantic versioning constraints if possible otherwise
			// just compare the version rule and the module version.
			if len(rule.Modules) > 0 {
				if moduleName != "" {
					match := false
					for ruleModuleName, ruleModuleVersion := range rule.Modules {
						if moduleName == ruleModuleName {
							if ruleModuleVersion == "" {
								match = true
								break
							}

							v, err := semver.NewVersion(moduleVersion)
							if err != nil {
								match = moduleVersion == ruleModuleVersion
								logrus.Debugf("%q - %s", moduleVersion, err)
								break
							}

							c, err := semver.NewConstraint(ruleModuleVersion)
							if err != nil {
								match = moduleVersion == ruleModuleVersion
								logrus.Debugf("%q %s", err, ruleModuleVersion)
								break
							}

							match = c.Check(v)
							break
						}
					}
					ruleResults = append(ruleResults, match)
				}
			}

			// If at least one rule is failing then we return false
			isAllMatching := true
			for i := range ruleResults {
				if !ruleResults[i] {
					isAllMatching = false
					break
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
