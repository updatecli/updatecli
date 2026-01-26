package bazel

import (
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
)

// MatchingRule allows specifying rules to identify manifests
type MatchingRule struct {
	// `path` specifies a `MODULE.bazel` path pattern, the pattern requires to match all of name, not just a substring.
	Path string
	/*
		`modules` specifies a map of modules, the key is module name as seen in the `MODULE.bazel`,
		the value is an optional semver version constraint.

		examples:
		```
		- modules:
		  # Ignoring module updates for this module
		  rules_go:
		  # Ignore module updates for this version
		  gazelle: "1.x"
		```
	*/
	Modules map[string]string
}

type MatchingRules []MatchingRule

// isMatchingRules checks for each matchingRule if parameters are matching rules and then return true or false.
func (m MatchingRules) isMatchingRules(rootDir, filePath, moduleName, moduleVersion string) bool {
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
				Checks if module is matching the module constraint.
				If both module constraint is empty and no moduleName have been provided then we
				assume the rule is matching

				Otherwise we checks both that version and module name are matching.
				Version matching uses semantic versioning constraints if possible otherwise
				just compare the version rule and the module version.
			*/

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

			/*
				If at least one rule is failing then we return false
			*/
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
