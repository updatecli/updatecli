package terragrunt

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/redact"
)

const (
	tfWildcard = "updatecliswildcard"
)

// MatchingRule defines a rule to select Terraform modules.
// A rule matches when every field it sets matches.
type MatchingRule struct {
	// "path" defines a Terragrunt file path pattern.
	//
	// remark:
	//   * the pattern must match the whole path, not just a substring.
	//   * the pattern follows the Go filepath.Match syntax, such as "*" or "?".
	//
	Path string
	// "modules" defines the Terraform modules to match, keyed by module source as written in the Terragrunt files.
	//
	// remark:
	//   * a Git source matches every module whose source starts with it.
	//   * a "tfr://" registry source can omit its trailing parts to match a whole registry or namespace.
	//   * an empty value matches any version.
	//   * otherwise the value is a semantic version constraint, such as ">=1.0.0".
	//   * when the version or the constraint cannot be parsed, the value must equal the version.
	//
	// example:
	//   ```
	//   modules:
	//     # ignore every module from a registry
	//     tfr://registry.opentofu.org:
	//     # ignore a specific module
	//     tfr:///terraform-aws-modules/rdss/aws:
	//     # ignore the versions matching this constraint
	//     git@github.com:hashicorp/exampleLongNameForSorting.git: "1.x"
	//   ```
	//
	Modules map[string]string
}

// MatchingRules defines a list of rules.
// The list matches when at least one of its rules matches.
type MatchingRules []MatchingRule

// isMatchingRules checks for each matchingRule if parameters are matching rules and then return true or false.
func (m MatchingRules) isMatchingRules(rootDir string, filePath string, module *terragruntModule) (bool, error) {
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
				Checks if module is matching the module constraint.
				If both module constraint is empty then we
				assume the rule is matching

				Otherwise we checks both that version and module  are matching.
				Version matching uses semantic versioning constraints if possible otherwise
				just compare the version rule and the module version.
			*/

			if len(rule.Modules) > 0 {
				match := false
				for ruleModuleUrl, ruleModuleVersion := range rule.Modules {
					// 1. Parse the ruleModuleUrl
					// tf expect version to have 3 to 4 slash, but let's allow for kind of wildcarding here
					if strings.HasPrefix(ruleModuleUrl, "tfr://") {
						baseUrl := strings.TrimPrefix(ruleModuleUrl, "tfr://")
						// Handle tfr:/// (triple slash) format by removing leading slash
						baseUrl = strings.TrimPrefix(baseUrl, "/")
						suffix := ""
						add := 3
						if baseUrl == "" {
							// ignore the whole default registry
							add = 3
						} else {
							tokens := strings.Split(baseUrl, "/")
							tokenCount := len(tokens)
							if baseUrl == "" {
								tokenCount = 0
							}
							if strings.Contains(tokens[0], ".") {
								// first token is an host, we need to have 4 tokens
								add += 1
							}
							add -= tokenCount
						}
						suffix += strings.Repeat(fmt.Sprintf("/%s", tfWildcard), add)
						if baseUrl == "" {
							suffix = strings.TrimPrefix(suffix, "/")
						}
						ruleModuleUrl += suffix
					}
					ruleModule, err := getModuleFromUrl(ruleModuleUrl, "", true)
					if err != nil {
						return false, fmt.Errorf("invalid module source url %q: %v", redact.URL(ruleModuleUrl), err)
					}
					nameMatch := false
					if ruleModule.source.sourceType == SourceTypeGit || ruleModule.source.sourceType == SourceTypeGithub {
						nameMatch = strings.HasPrefix(module.source.evaluatedSource, ruleModule.source.evaluatedSource)
					} else if ruleModule.source.sourceType == SourceTypeRegistry && module.registryModule != nil {
						if ruleModule.registryModule == nil {
							return false, fmt.Errorf("invalid tfr module source url %q", redact.URL(ruleModuleUrl))
						}
						registryModule := *ruleModule.registryModule
						nameMatch = registryModule.Package.Host == tfWildcard || registryModule.Package.Host == module.registryModule.Package.Host
						if nameMatch {
							nameMatch = registryModule.Package.Namespace == tfWildcard || registryModule.Package.Namespace == module.registryModule.Package.Namespace
							if nameMatch {
								nameMatch = registryModule.Package.Name == tfWildcard || registryModule.Package.Name == module.registryModule.Package.Name
								if nameMatch {
									nameMatch = registryModule.Package.TargetSystem == tfWildcard || registryModule.Package.TargetSystem == module.registryModule.Package.TargetSystem
								}
							}
						}
					}
					if nameMatch {
						if ruleModuleVersion == "" {
							match = true
							break
						}
						moduleVersion := module.source.version
						v, err := semver.NewVersion(moduleVersion)
						if err != nil {
							match = moduleVersion == ruleModuleVersion
							logrus.Debugf("%q - %s", moduleVersion, err)
							break
						}

						c, err := semver.NewConstraint(ruleModuleVersion)
						if err != nil {
							match = moduleVersion == ruleModuleVersion
							logrus.Debugf("%q %s", err, ruleModuleVersion)
							break
						}

						match = c.Check(v)
						break
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
				return true, nil
			}
			ruleResults = []bool{}
		}
	}

	return false, nil
}
