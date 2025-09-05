package golang

import (
	"path/filepath"
	"regexp"

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
)

// MatchingRule allows to specifies rules to identify manifest
// to update based on file path, module name, module version, go version and replace directive.
type MatchingRule struct {
	// Path specifies a go.mod path pattern
	// The pattern syntax is:
	// pattern:
	//     { term }
	// term:
	//     '*'         matches any sequence of non-Separator characters
	//     '?'         matches any single non-Separator character
	//     '[' [ '^' ] { character-range } ']'
	//                 character class (must be non-empty)
	//     c           matches character c (c != '*', '?', '\\', '[')
	//     '\\' c      matches character c
	// example:
	//   * 'go.mod' matches 'go.mod' in the currrent directory
	//   * '*/go.mod' matches 'go.mod' in any first level subdirectory
	Path string
	// Modules specifies a list of module pattern.
	// The module accepts regular expression for module name and semantic versioning constraint for module version.
	// If module version is empty then any version is matching.
	// Example:
	//   * 'github.com/updatecli/updatecli': '' matches any version of the module github.com/updatecli/updatecli
	//   * 'github.com/updatecli/updatecli': '1.0.0' matches only version 1.0.0 of the module github.com/updatecli/updatecli
	//   * 'github.com/updatecli/updatecli': '>=1.0.0' matches any version greater than or equal to 1.0.0 of the module github.com/updatecli/updatecli
	//   * 'github.com/.*': '>=1.0.0' matches any version greater than or equal to 1.0.0 of any module hosted under github.com
	//   * 'github\.com\/updatecli\/updatecli': '>=1.0.0' matches any version greater than or equal to 1.
	Modules map[string]string
	// GoVersions specifies a list of version pattern.
	// The version constraint must be a valid semantic version constraint.
	// If GoVersion is empty then any version is matching.
	// Example:
	//   * '1.19.*' matches any 1.19.x version
	//  * '>=1.20.0' matches any version greater than or equal to 1.20.0
	//   * '<1.20.0' matches any version strictly less than 1.20.0
	//   * '*' matches any version
	GoVersion string
	// Replace indicates if the module is a replace directive.
	// If Replace is nil then any module is matching.
	// If Replace is true then only module with a replace directive is matching.
	// If Replace is false then only module without a replace directive is matching.
	Replace *bool
}

type MatchingRules []MatchingRule

// isMatchingRules checks for each matchingRule if parameters are matching rules and then return true or false.
func (m MatchingRules) isMatchingRules(rootDir, filePath, goVersion, moduleName, moduleVersion string, isReplaced bool) bool {
	var ruleResults []bool

	if len(m) > 0 {
		for _, rule := range m {

			if rule.Replace != nil {
				match := false
				if *rule.Replace == isReplaced {
					match = true
				}
				ruleResults = append(ruleResults, match)
			}

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
				If both module constraint is empty and no modulename have been provided then we
				assume the rule is matching

				Otherwise we checks both that version and module name are matching.
				Version matching uses semantic versioning constraints if possible otherwise
				just compare the version rule and the module version.
			*/

			if len(rule.Modules) > 0 {
				if moduleName != "" {
					match := false
				outModule:
					for ruleModuleName, ruleModuleVersion := range rule.Modules {

						moduleMatch, err := regexp.MatchString(ruleModuleName, moduleName)
						if err != nil {
							logrus.Debugf("%q - %s", ruleModuleName, err)
							break outModule
						}

						if moduleMatch {
							if ruleModuleVersion == "" {
								match = true
								break outModule
							}

							v, err := semver.NewVersion(moduleVersion)
							if err != nil {
								match = moduleVersion == ruleModuleVersion
								logrus.Debugf("%q - %s", moduleVersion, err)
								break outModule
							}

							c, err := semver.NewConstraint(ruleModuleVersion)
							if err != nil {
								match = moduleVersion == ruleModuleVersion
								logrus.Debugf("%q %s", err, ruleModuleVersion)
								break outModule
							}

							match = c.Check(v)
							break outModule
						}
					}
					ruleResults = append(ruleResults, match)
				}

			}

			/*
				Checks if the goVersion is matching the rule
				The version constraint must be a valid semantic version constraint.
			*/
			if rule.GoVersion != "" {
				if goVersion == "" {
					ruleResults = append(ruleResults, false)
					goto goVersionDone
				}

				v, err := semver.NewVersion(goVersion)
				if err != nil {
					logrus.Errorln(err)
					ruleResults = append(ruleResults, false)
					goto goVersionDone
				}

				c, err := semver.NewConstraint(rule.GoVersion)
				if err != nil {
					logrus.Errorln(err)
					ruleResults = append(ruleResults, false)
					goto goVersionDone
				}

				ruleResults = append(ruleResults, c.Check(v))
			}

		goVersionDone:

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

func (m MatchingRules) isGoVersionOnly() bool {
	// If there is no rule then we assume it is not only go version
	if len(m) == 0 {
		return false
	}

	goOnlyFound := 0
	for _, rule := range m {
		if rule.GoVersion != "" && len(rule.Modules) == 0 {
			goOnlyFound++
		}
	}

	if goOnlyFound == len(m) && goOnlyFound > 0 {
		return true
	}

	return false
}

func (m MatchingRules) isGoModuleOnly() bool {
	// If there is no rule then we assume it is not only go version
	if len(m) == 0 {
		return false
	}

	moduleOnly := 0
	for _, rule := range m {
		if rule.GoVersion == "" && len(rule.Modules) > 0 {
			moduleOnly++
		}
	}

	if moduleOnly == len(m) && moduleOnly > 0 {
		return true
	}

	return false
}
