package golang

import (
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
)

// MatchingRule defines a rule to select Go modules or Go versions.
// A rule matches when every field it sets matches.
// Each rule must set at least one field.
type MatchingRule struct {
	// "path" defines a go.mod path pattern.
	//
	// remark:
	//   * the pattern must match the whole path, not just a substring.
	//   * the pattern follows the Go filepath.Match syntax, such as "*" or "?".
	//
	// example:
	//   * path: go.mod
	//   * path: "*/go.mod"
	//
	Path string
	// "modules" defines the Go modules to match, keyed by module name.
	//
	// remark:
	//   * the key is a regular expression matched against the module name.
	//   * the expression is not anchored, so it also matches module names that contain it.
	//   * an empty value matches any version.
	//   * otherwise the value is a semantic version constraint, such as ">=1.0.0".
	//   * when the version or the constraint cannot be parsed, the value must equal the version.
	//
	// example:
	//   * "github.com/updatecli/updatecli": "" matches any version of the module.
	//   * "github.com/updatecli/updatecli": "1.0.0" matches only version 1.0.0 of the module.
	//   * "github.com/updatecli/updatecli": ">=1.0.0" matches version 1.0.0 or later of the module.
	//   * "github.com/.*": ">=1.0.0" matches version 1.0.0 or later of any module hosted on github.com.
	//
	Modules map[string]string
	// "goversion" defines a Go version constraint to match.
	//
	// remark:
	//   * the value must be a valid semantic version constraint.
	//   * when unset, any Go version matches.
	//
	// example:
	//   * goversion: "1.19.*"
	//   * goversion: ">=1.20.0"
	//   * goversion: "<1.20.0"
	//   * goversion: "*"
	//
	GoVersion string
	// "replace" defines whether the module must come from a replace directive.
	//
	// default:
	//   unset, any module matches.
	//
	// remark:
	//   * true matches only modules with a replace directive.
	//   * false matches only modules without a replace directive.
	//
	Replace *bool
}

// MatchingRules defines a list of rules.
// The list matches when at least one of its rules matches.
type MatchingRules []MatchingRule

// Validate checks that each matching rule has at least one non-empty field.
// Returns an error if any rule has no valid fields specified.
func (m MatchingRules) Validate() error {
	for i, rule := range m {
		if rule.Path == "" && len(rule.Modules) == 0 && rule.GoVersion == "" && rule.Replace == nil {
			return fmt.Errorf("rule %d has no valid fields (path, modules, goversion, or replace must be specified)", i+1)
		}
	}
	return nil
}

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
