package helmfile

import (
	"fmt"
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
)

// MatchingRule allows to specifies rules to identify manifest
type MatchingRule struct {
	// Path specifies a Helmfile chart path pattern, the pattern requires to match all of name, not just a subpart of the path.
	Path string
	// Repositories specifies the list of Helm Chart repository to check
	Repositories []string
	// Charts specifies the list of Helm Chart repository to check
	Charts map[string]string
}

type MatchingRules []MatchingRule

// Validate checks that each matching rule has at least one non-empty field.
// Returns an error if any rule has no valid fields specified.
func (m MatchingRules) Validate() error {
	for i, rule := range m {
		if rule.Path == "" && len(rule.Repositories) == 0 && len(rule.Charts) == 0 {
			return fmt.Errorf("rule %d has no valid fields (path, repositories, or charts must be specified)", i+1)
		}
	}
	return nil
}

// isMatchingRules checks if a specific file content matches the "only" rule
func (m MatchingRules) isMatchingRules(rootDir, filePath, repositoryURL, chartName, chartVersion string) bool {
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
				Checks if Chart Repository is matching the policy constraint.
			*/

			if len(rule.Repositories) > 0 {
				match := false

			outRepository:
				for i := range rule.Repositories {

					if repositoryURL == rule.Repositories[i] {
						match = true
						break outRepository
					}
				}
				ruleResults = append(ruleResults, match)
			}

			/*
				Checks if Chart Name is matching the policy constraint.
			*/

			if len(rule.Charts) > 0 {
				match := false

			outChartName:
				for ruleReleaseName, ruleReleaseVersion := range rule.Charts {

					if chartName == ruleReleaseName {
						if ruleReleaseVersion == "" {
							match = true
							break outChartName
						}

						v, err := semver.NewVersion(chartVersion)
						if err != nil {
							match = chartVersion == ruleReleaseVersion
							logrus.Debugf("%q - %s", chartVersion, err)
							break outChartName
						}

						c, err := semver.NewConstraint(ruleReleaseVersion)
						if err != nil {
							match = chartVersion == ruleReleaseVersion
							logrus.Debugf("%q %s", err, ruleReleaseVersion)
							break outChartName
						}

						match = c.Check(v)
						break outChartName
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
