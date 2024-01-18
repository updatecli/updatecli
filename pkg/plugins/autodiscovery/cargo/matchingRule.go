package cargo

import (
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
)

// MatchingRule allows to specifies rules to identify manifest
type MatchingRule struct {
	// Path specifies a Cargo crate path pattern, the pattern requires to match all of name, not just a subpart of the path.
	Path string `yaml:",omitempty"`
	// Crates specifies the list of Cargo crates to check
	Crates map[string]string `yaml:",omitempty"`
	// Registries specifies the list of Cargo registries to check
	Registries []string `yaml:",omitempty"`
}

type MatchingRules []MatchingRule

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
