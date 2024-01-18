package helm

import (
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
)

// MatchingRule allows to specifies rules to identify manifest
type MatchingRule struct {
	// Path specifies a Helm chart path pattern, the pattern requires to match all of name, not just a subpart of the path.
	Path string
	// Dependencies specifies a list of dependencies pattern.
	Dependencies map[string]string
	// Containers specifies a list of containers pattern.
	Containers map[string]string
}

type MatchingRules []MatchingRule

// isMatchingRules checks for each matchingRule if parameters are matching rules and then return true or false.
func (m MatchingRules) isMatchingRules(rootDir, filePath, dependencyName, dependencyVersion, containerName, containerTag string) bool {
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
				Checks if dependency is matching the dependency constraint.
				If both the dependency constraint is empty and no dependency name have been provided then we
				assume the rule is not matching

				Otherwise we checks both that version and module name are matching.
				Version matching uses semantic versioning constraints if possible otherwise
				just compare the version rule and the module version.
			*/

			if len(rule.Dependencies) > 0 {
				match := false
			outDependency:
				for ruleDependencyName, ruleDependencyVersion := range rule.Dependencies {
					if dependencyName == ruleDependencyName {
						if ruleDependencyVersion == "" {
							match = true
							break outDependency
						}

						v, err := semver.NewVersion(dependencyVersion)
						if err != nil {
							match = dependencyVersion == ruleDependencyVersion
							logrus.Debugf("%q - %s", dependencyVersion, err)
							break outDependency
						}

						c, err := semver.NewConstraint(ruleDependencyVersion)
						if err != nil {
							match = dependencyVersion == ruleDependencyVersion
							logrus.Debugf("%q %s", err, ruleDependencyVersion)
							break outDependency
						}

						match = c.Check(v)
						break outDependency
					}
				}
				ruleResults = append(ruleResults, match)

			}

			/*
				Checks if container is matching the container constraint.
				If the dependency constraint is empty and no dependency name have been provided then we
				assume the rule is not matching

				Otherwise we checks both that version and module name are matching.
				Version matching uses semantic versioning constraints if possible otherwise
				just compare the version rule and the module version.
			*/
			if len(rule.Containers) > 0 {
				match := false
			outContainer:
				for ruleContainerName, ruleContainerTag := range rule.Containers {
					if containerName == ruleContainerName {
						if ruleContainerTag == "" {
							match = true
							break outContainer
						}

						v, err := semver.NewVersion(containerTag)
						if err != nil {
							match = containerTag == ruleContainerTag
							logrus.Debugf("%q - %s", containerTag, err)
							break outContainer
						}

						c, err := semver.NewConstraint(ruleContainerTag)
						if err != nil {
							match = containerTag == ruleContainerTag
							logrus.Debugf("%q %s", err, ruleContainerTag)
							break outContainer
						}

						match = c.Check(v)
						break outContainer
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
	}

	return false
}
