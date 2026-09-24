package terraform

import (
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
)

// MatchingRule defines a rule to select Terraform providers.
// A rule matches when every field it sets matches.
type MatchingRule struct {
	// "path" defines a ".terraform.lock.hcl" path pattern.
	//
	// remark:
	//   * the pattern must match the whole path, not just a substring.
	//   * the pattern follows the Go filepath.Match syntax, such as "*" or "?".
	//
	Path string
	// "providers" defines the Terraform providers to match, keyed by provider address as written in ".terraform.lock.hcl".
	//
	// remark:
	//   * an empty value matches any version.
	//   * otherwise the value is a semantic version constraint, such as ">=1.0.0".
	//   * when the version or the constraint cannot be parsed, the value must equal the version.
	//
	// example:
	//   ```
	//   providers:
	//     # ignore every version of this provider
	//     registry.terraform.io/hashicorp/aws:
	//     # ignore the versions matching this constraint
	//     registry.terraform.io/hashicorp/kubernetes: "1.x"
	//   ```
	//
	Providers map[string]string
}

// MatchingRules defines a list of rules.
// The list matches when at least one of its rules matches.
type MatchingRules []MatchingRule

// isMatchingRules checks for each matchingRule if parameters are matching rules and then return true or false.
func (m MatchingRules) isMatchingRules(rootDir, filePath, providerName, providerVersion string) bool {
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
				Checks if provider is matching the provider constraint.
				If both provider constraint is empty and no providerName have been provided then we
				assume the rule is matching

				Otherwise we checks both that version and provider name are matching.
				Version matching uses semantic versioning constraints if possible otherwise
				just compare the version rule and the module version.
			*/

			if len(rule.Providers) > 0 {
				if providerName != "" {
					match := false
					for ruleProviderName, ruleProviderVersion := range rule.Providers {
						if providerName == ruleProviderName {
							if ruleProviderVersion == "" {
								match = true
								break
							}

							v, err := semver.NewVersion(providerVersion)
							if err != nil {
								match = providerVersion == ruleProviderVersion
								logrus.Debugf("%q - %s", providerVersion, err)
								break
							}

							c, err := semver.NewConstraint(ruleProviderVersion)
							if err != nil {
								match = providerVersion == ruleProviderVersion
								logrus.Debugf("%q %s", err, ruleProviderVersion)
								break
							}

							match = c.Check(v)
							break
						}
					}
					ruleResults = append(ruleResults, match)
				}
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
