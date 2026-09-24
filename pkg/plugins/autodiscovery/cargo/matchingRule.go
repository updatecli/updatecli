package cargo

import (
	"fmt"
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
)

// MatchingRule defines a rule to select crates.
// A rule matches when every field it sets matches.
// Each rule must set at least one field.
type MatchingRule struct {
	// "path" defines a "Cargo.toml" path pattern.
	//
	// remark:
	//   * the pattern must match the whole path, not just a substring.
	//   * the pattern follows the Go filepath.Match syntax, such as "*" or "?".
	//
	Path string `yaml:",omitempty"`
	// "crates" defines the crates to match, keyed by crate name.
	//
	// remark:
	//   * an empty value matches any version.
	//   * otherwise the value is a semantic version constraint, such as ">=1.0.0".
	//   * when the version or the constraint cannot be parsed, the value must equal the version.
	//
	Crates map[string]string `yaml:",omitempty"`
	// "registries" defines the Cargo registry names to match.
	//
	// remark:
	//   * the registry name is the "registry" key of a dependency in "Cargo.toml".
	//   * a registry name must be identical to one of the entries.
	//
	Registries []string `yaml:",omitempty"`
}

// MatchingRules defines a list of rules.
// The list matches when at least one of its rules matches.
type MatchingRules []MatchingRule

// Validate checks that each matching rule has at least one non-empty field.
// Returns an error if any rule has no valid fields specified.
func (m MatchingRules) Validate() error {
	for i, rule := range m {
		if rule.Path == "" && len(rule.Crates) == 0 && len(rule.Registries) == 0 {
			return fmt.Errorf("rule %d has no valid fields (path, crates, or registries must be specified)", i+1)
		}
	}
	return nil
}

// isMatchingRules checks if a specific file content matches the "only" rule
func (m MatchingRules) isMatchingRules(rootDir, filePath, registry, crateName, crateVersion string) bool {
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
				Checks if the groupid is matching the policy constraint.
			*/

			if len(rule.Registries) > 0 {
				match := false

			outRegistry:
				for i := range rule.Registries {

					if registry == rule.Registries[i] {
						match = true
						break outRegistry
					}
				}
				ruleResults = append(ruleResults, match)
			}

			/*
				Checks if Chart Name is matching the policy constraint.
			*/

			if len(rule.Crates) > 0 {
				match := false

			outCrate:
				for ruleCrateName, ruleCrateVersion := range rule.Crates {

					if crateName == ruleCrateName {
						if ruleCrateVersion == "" {
							match = true
							break outCrate
						}

						v, err := semver.NewVersion(crateVersion)
						if err != nil {
							match = crateVersion == ruleCrateVersion
							logrus.Debugf("%q - %s", crateVersion, err)
							break outCrate
						}

						c, err := semver.NewConstraint(ruleCrateVersion)
						if err != nil {
							match = crateVersion == ruleCrateVersion
							logrus.Debugf("%q %s", err, ruleCrateVersion)
							break outCrate
						}

						match = c.Check(v)
						break outCrate
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
