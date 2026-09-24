package nomad

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

// MatchingRule defines a rule to select container images.
// A rule matches when every field it sets matches.
// Each rule must set at least one field.
type MatchingRule struct {
	// "path" defines a Nomad file path pattern.
	//
	// remark:
	//   * the pattern must match the whole path, not just a substring.
	//   * the pattern follows the Go filepath.Match syntax, such as "*" or "?".
	//
	Path string
	// "jobs" defines the Nomad job names to match.
	//
	// remark:
	//   * a job matches when its name equals one of the values.
	//
	Jobs []string
	// "images" defines the container image names to match.
	//
	// remark:
	//   * an image matches when its name starts with one of the values.
	//
	Images []string
}

// MatchingRules defines a list of rules.
// The list matches when at least one of its rules matches.
type MatchingRules []MatchingRule

// Validate checks that each matching rule has at least one non-empty field.
// Returns an error if any rule has no valid fields specified.
func (m MatchingRules) Validate() error {
	for i, rule := range m {
		if rule.Path == "" && len(rule.Jobs) == 0 && len(rule.Images) == 0 {
			return fmt.Errorf("rule %d has no valid fields (path, jobs, or images must be specified)", i+1)
		}
	}
	return nil
}

// isMatchingRule tests that all defined rule are matching and return true if it's the case otherwise return false
func (m MatchingRules) isMatchingRule(rootDir, filePath, job, image string) bool {
	// Test if the ignore rule based on path is respected

	var ruleResults []bool

	if len(m) > 0 {
		for _, matchingRule := range m {
			var match bool
			var err error

			// Only check if path rule defined
			if matchingRule.Path != "" && filePath != "" {
				if filepath.IsAbs(matchingRule.Path) {
					filePath = filepath.Join(rootDir, filePath)
				}

				match, err = filepath.Match(matchingRule.Path, filePath)
				if err != nil {
					logrus.Errorf("%s - %q", err, matchingRule.Path)
				}
				ruleResults = append(ruleResults, match)
				if match {
					logrus.Debugf("file path %q matching rule %q", filePath, matchingRule.Path)
				}
			}

			// Only check if service rule defined.
			if len(matchingRule.Jobs) > 0 && job != "" {
				match := false
				for _, ms := range matchingRule.Jobs {
					if ms == job {
						logrus.Debugf("service %q matching rule %q", job, ms)
						match = true
						break
					}
				}
				ruleResults = append(ruleResults, match)
			}
			// Only check if image rule defined.
			if len(matchingRule.Images) > 0 && image != "" {
				match := false
				for _, i := range matchingRule.Images {
					if strings.HasPrefix(image, i) {
						logrus.Debugf("image %q matching rule %q", image, i)
						match = true
						break
					}
				}
				ruleResults = append(ruleResults, match)
			}

			allMatchingRule := true
			for i := range ruleResults {
				if !ruleResults[i] {
					allMatchingRule = false
					break
				}
			}

			if allMatchingRule {
				return true
			}
		}

	}

	return false
}
