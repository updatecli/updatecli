package dockercompose

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
	// "archs" defines the image architectures to match.
	//
	// remark:
	//   * an architecture must be identical to one of the entries.
	//   * the architecture is the middle part of the "platform" key of a service,
	//     so "linux/amd64" gives "amd64". Entries must use that form, such as "amd64".
	//   * a service without an architecture in its "platform" key is not checked against "archs".
	//
	// example:
	//   * archs: ["amd64", "arm64"]
	//
	Archs []string
	// "path" defines a Docker Compose file path pattern.
	//
	// remark:
	//   * the pattern must match the whole path, not just a substring.
	//   * the pattern follows the Go filepath.Match syntax, such as "*" or "?".
	//
	Path string
	// "services" defines the Docker Compose service names to match.
	//
	// remark:
	//   * a service name must be identical to one of the entries.
	//
	Services []string
	// "images" defines the container images to match.
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
		if rule.Path == "" && len(rule.Archs) == 0 && len(rule.Services) == 0 && len(rule.Images) == 0 {
			return fmt.Errorf("rule %d has no valid fields (path, archs, services, or images must be specified)", i+1)
		}
	}
	return nil
}

// isMatchingRule tests that all defined rule are matching and return true if it's the case otherwise return false
func (m MatchingRules) isMatchingRule(rootDir, filePath, service, image, arch string) bool {
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
			if len(matchingRule.Archs) > 0 && arch != "" {
				match := false
				for _, a := range matchingRule.Archs {
					if a == arch {
						logrus.Debugf("arch %q matching rule %q", service, a)
						match = true
						break
					}
				}
				ruleResults = append(ruleResults, match)
			}

			// Only check if service rule defined.
			if len(matchingRule.Services) > 0 && service != "" {
				match := false
				for _, ms := range matchingRule.Services {
					if ms == service {
						logrus.Debugf("service %q matching rule %q", service, ms)
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
