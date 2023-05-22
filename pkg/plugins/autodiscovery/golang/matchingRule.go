package golang

import (
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
)

// MatchingRule allows to specifies rules to identify manifest
type MatchingRule struct {
	// Path specifies a go.mod path pattern, the pattern requires to match all of name, not just a substring.
	Path string
	// Modules specifies a list of module pattern.
	Modules map[string]string
	// GoVersions specifies a list of version pattern.
	GoVersion string
}

type MatchingRules []MatchingRule

// isMatchingRules checks for each matchingRule if parameters are matching rules and then return true or false.
func (m MatchingRules) isMatchingRules(rootDir, filePath, goVersion, moduleName, moduleVersion string) bool {
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
						if moduleName == ruleModuleName {
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
