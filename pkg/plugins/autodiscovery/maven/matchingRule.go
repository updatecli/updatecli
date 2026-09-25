package maven

import (
	"fmt"
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
)

// MatchingRule defines a rule to select Maven dependencies.
// A rule matches when every field it sets matches.
// Each rule must set at least one field.
type MatchingRule struct {
	// "path" defines a pom.xml path pattern.
	//
	// remark:
	//   * the pattern must match the whole path, not just a substring.
	//   * the pattern follows the Go filepath.Match syntax, such as "*" or "?".
	//
	Path string `yaml:",omitempty"`
	// "groupids" defines the Maven group IDs to match.
	//
	// remark:
	//   * a dependency matches when its group ID equals one of the values.
	//
	GroupIDs []string `yaml:",omitempty"`
	// "artifactids" defines the Maven artifacts to match, keyed by artifact ID.
	//
	// remark:
	//   * an empty value matches any version.
	//   * otherwise the value is a semantic version constraint, such as ">=1.0.0".
	//   * when the version or the constraint cannot be parsed, the value must equal the version.
	//
	ArtifactIDs map[string]string `yaml:",omitempty"`
}

// MatchingRules defines a list of rules.
// The list matches when at least one of its rules matches.
type MatchingRules []MatchingRule

// Validate checks that each matching rule has at least one non-empty field.
// Returns an error if any rule has no valid fields specified.
func (m MatchingRules) Validate() error {
	for i, rule := range m {
		if rule.Path == "" && len(rule.GroupIDs) == 0 && len(rule.ArtifactIDs) == 0 {
			return fmt.Errorf("rule %d has no valid fields (path, groupids, or artifactids must be specified)", i+1)
		}
	}
	return nil
}

// isMatchingRules checks if a specific file content matches the "only" rule
func (m MatchingRules) isMatchingRules(rootDir, filePath, groupid, artifactName, artifactVersion string) bool {
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

			if len(rule.GroupIDs) > 0 {
				match := false

			outGroupID:
				for i := range rule.GroupIDs {

					if groupid == rule.GroupIDs[i] {
						match = true
						break outGroupID
					}
				}
				ruleResults = append(ruleResults, match)
			}

			/*
				Checks if Chart Name is matching the policy constraint.
			*/

			if len(rule.ArtifactIDs) > 0 {
				match := false

			outArtifactName:
				for ruleArtifactName, ruleArtifactVersion := range rule.ArtifactIDs {

					if artifactName == ruleArtifactName {
						if ruleArtifactVersion == "" {
							match = true
							break outArtifactName
						}

						v, err := semver.NewVersion(artifactVersion)
						if err != nil {
							match = artifactVersion == ruleArtifactVersion
							logrus.Debugf("%q - %s", artifactVersion, err)
							break outArtifactName
						}

						c, err := semver.NewConstraint(ruleArtifactVersion)
						if err != nil {
							match = artifactVersion == ruleArtifactVersion
							logrus.Debugf("%q %s", err, ruleArtifactVersion)
							break outArtifactName
						}

						match = c.Check(v)
						break outArtifactName
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
