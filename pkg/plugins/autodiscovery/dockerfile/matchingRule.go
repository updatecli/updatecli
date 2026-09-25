package dockerfile

import (
	"fmt"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// MatchingRule defines a rule to select container images.
// A rule matches when every field it sets matches.
// Each rule must set at least one field.
type MatchingRule struct {
	// "archs" defines the image architectures to match.
	//
	// remark:
	//   * an architecture must be identical to one of the entries.
	//   * the architecture is the value of the "--platform" flag of a "FROM" instruction, such as "linux/amd64".
	//
	Archs []string
	// "path" defines a Dockerfile path pattern.
	//
	// remark:
	//   * the pattern must match the whole path, not just a substring.
	//   * the pattern follows the Go filepath.Match syntax, such as "*" or "?".
	//
	Path string
	// "images" defines the container images to match.
	//
	// remark:
	//   * an image name, without its tag, must be identical to one of the entries.
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
		if rule.Path == "" && len(rule.Archs) == 0 && len(rule.Images) == 0 {
			return fmt.Errorf("rule %d has no valid fields (path, archs, or images must be specified)", i+1)
		}
	}
	return nil
}

// isMatchingRule tests that all defined rule are matching and return true if it's the case otherwise return false
func (m MatchingRules) isMatchingRule(rootDir, filePath, image, arch string) bool {
	// Test if the ignore rule are respected
	var ruleResults []bool

	if len(m) > 0 {
		for _, rule := range m {
			var match bool
			var err error

			// Only check if path rule defined
			if rule.Path != "" {
				if filepath.IsAbs(rule.Path) {
					filePath = filepath.Join(rootDir, filePath)
				}

				match, err = filepath.Match(rule.Path, filePath)
				if err != nil {
					logrus.Errorf("%s - %q", err, rule.Path)
					continue
				}
				ruleResults = append(ruleResults, match)
				if match {
					logrus.Debugf("file path %q matching rule %q", filePath, rule.Path)
				}
			}

			// Only check if service rule defined.
			if len(rule.Archs) > 0 {
				match := false
			archs:
				for _, a := range rule.Archs {
					if a == arch {
						logrus.Debugf("arch %q matching rule %q", arch, a)
						match = true
						break archs
					}
				}
				ruleResults = append(ruleResults, match)
			}

			// Only check if image rule defined.
			if len(rule.Images) > 0 {
				match := false
			images:
				for _, i := range rule.Images {
					if image == i {
						logrus.Debugf("image %q matching rule %q", image, i)
						match = true
						break images
					}
				}
				ruleResults = append(ruleResults, match)
			}

			if len(ruleResults) == 0 {
				return false
			}

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
