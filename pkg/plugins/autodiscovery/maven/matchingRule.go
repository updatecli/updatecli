package maven

import (
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
)

// MatchingRule allows to specifies rules to identify manifest
type MatchingRule struct {
	// Path specifies a Helm chart path pattern, the pattern requires to match all of name, not just a substring.
	Path string `yaml:",omitempty"`
	// GroupIDs specifies the list of Maven GroupIDs to check
	GroupIDs []string `yaml:",omitempty"`
	// ArtifactIDs specifies the list of Maven ArtifactIDs to check
	ArtifactIDs map[string]string `yaml:",omitempty"`
}

type MatchingRules []MatchingRule

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
