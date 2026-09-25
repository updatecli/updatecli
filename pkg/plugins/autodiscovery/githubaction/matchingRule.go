package githubaction

import (
	"fmt"
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
)

// MatchingRule defines a rule to select actions and Docker images.
// A rule matches when every field it sets matches.
// Each rule must set at least one field.
type MatchingRule struct {
	// "path" defines a workflow or composite action file path pattern.
	//
	// remark:
	//   * the pattern must match the whole path, not just a substring.
	//   * the pattern follows the Go filepath.Match syntax, such as "*" or "?".
	//
	Path string
	// "actions" defines the actions and Docker images to match, keyed by name.
	//
	// remark:
	//   * for an action, the key is the action name, such as "actions/checkout".
	//   * for a Docker image, the key is the image reference as written in the workflow, tag included,
	//     such as "docker://alpine:3.18" for a step or "alpine:3.18" for a job container.
	//   * an empty value matches any version.
	//   * otherwise the value is a semantic version constraint, such as ">=1.0.0".
	//   * when the version or the constraint cannot be parsed, the value must equal the reference,
	//     such as a git branch, a git tag, or a Docker image tag.
	//
	Actions map[string]string
}

// MatchingRules defines a list of rules.
// The list matches when at least one of its rules matches.
type MatchingRules []MatchingRule

// Validate checks that each matching rule has at least one non-empty field.
// Returns an error if any rule has no valid fields specified.
func (m MatchingRules) Validate() error {
	for i, rule := range m {
		if rule.Path == "" && len(rule.Actions) == 0 {
			return fmt.Errorf("rule %d has no valid fields (path or actions must be specified)", i+1)
		}
	}
	return nil
}

// isMatchingRules checks if a specific file content matches the "only" rule
func (m MatchingRules) isMatchingRules(rootDir, filePath, chartName, chartVersion string) bool {
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
				Checks if Actions is matching the policy constraint.
			*/

			if len(rule.Actions) > 0 {
				match := false

			outAction:
				for repository, reference := range rule.Actions {

					if chartName == repository {
						if reference == "" {
							match = true
							break outAction
						}

						v, err := semver.NewVersion(chartVersion)
						if err != nil {
							match = chartVersion == reference
							logrus.Debugf("%q - %s", chartVersion, err)
							break outAction
						}

						c, err := semver.NewConstraint(reference)
						if err != nil {
							match = chartVersion == reference
							logrus.Debugf("%q %s", err, reference)
							break outAction
						}

						match = c.Check(v)
						break outAction
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
