package pyproject

import (
	"fmt"
	"path/filepath"

	gopep440 "github.com/aquasecurity/go-pep440-version"
	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pep503"
)

// MatchingRule defines a rule to select Python dependencies.
// A rule matches when every field it sets matches.
// Each rule must set at least one field.
type MatchingRule struct {
	// "path" defines a "pyproject.toml" path pattern.
	//
	// remark:
	//   * the pattern must match the whole path, not just a substring.
	//   * the pattern follows the Go filepath.Match syntax, such as "*" or "?".
	//
	Path string `yaml:",omitempty"`
	// "packages" defines the Python packages to match, keyed by package name.
	//
	// remark:
	//   * an empty value matches any version.
	//   * otherwise the value is a PEP 440 version specifier, such as ">=2.0,<3.0".
	//   * when the version or the specifier cannot be parsed, the value must equal the version.
	//
	Packages map[string]string `yaml:",omitempty"`
	// "dependencygroups" defines the PEP 735 "[dependency-groups]" names to match, such as "dev".
	//
	// remark:
	//   * each value is a pattern following the Go filepath.Match syntax, so "*" matches any group.
	//   * names are compared once normalized as PEP 735 requires, ignoring case and treating
	//     runs of "-", "_" and "." alike, so "dev-tools" matches a group declared as "Dev_Tools".
	//   * patterns are normalized the same way, including inside brackets, so "[_.]" means "[-]".
	//   * a group including another one through {include-group = "..."} contains its dependencies
	//     too, as PEP 735 defines, so "dev" matches the dependencies "dev" includes from "test".
	//   * dependencies from "[project.dependencies]" and "[project.optional-dependencies]"
	//     belong to no dependency group, so they never match this field.
	//
	// example:
	//   ```
	//   only:
	//     - dependencygroups:
	//         - dev
	//   ```
	//
	DependencyGroups []string `yaml:",omitempty"`
}

// MatchingRules defines a list of rules.
// The list matches when at least one of its rules matches.
type MatchingRules []MatchingRule

// Validate checks that every rule has at least one non-empty field.
func (m MatchingRules) Validate() error {
	for i, rule := range m {
		if rule.Path == "" && len(rule.Packages) == 0 && len(rule.DependencyGroups) == 0 {
			return fmt.Errorf("rule %d has no valid fields (path, packages or dependencygroups must be specified)", i+1)
		}
		for _, group := range rule.DependencyGroups {
			// Patterns are matched once normalized, so the normalized form is the one to check.
			if _, err := filepath.Match(pep503.NormalizeName(group), ""); err != nil {
				return fmt.Errorf("rule %d has an invalid dependencygroups pattern %q: %w", i+1, group, err)
			}
		}
	}
	return nil
}

// normalized returns a copy of the rules with their dependencygroups patterns normalized as
// PEP 735 requires, so matching compares them with normalized group names as they are.
func (m MatchingRules) normalized() MatchingRules {
	rules := make(MatchingRules, 0, len(m))
	for _, rule := range m {
		if len(rule.DependencyGroups) > 0 {
			patterns := make([]string, 0, len(rule.DependencyGroups))
			for _, pattern := range rule.DependencyGroups {
				patterns = append(patterns, pep503.NormalizeName(pattern))
			}
			rule.DependencyGroups = patterns
		}
		rules = append(rules, rule)
	}
	return rules
}

// isMatchingRules reports whether the given file/package pair matches any rule in the list.
// dependencyGroups holds the normalized names of the PEP 735 groups the package belongs to,
// empty when it belongs to none. The rules must have been normalized first.
// Multiple conditions within one rule are AND-ed; multiple rules are OR-ed.
func (m MatchingRules) isMatchingRules(rootDir, filePath, packageName, packageVersion string, dependencyGroups []string) bool {
	if len(m) == 0 {
		return false
	}

	for _, rule := range m {
		var ruleResults []bool

		if rule.Path != "" {
			fp := filePath
			if filepath.IsAbs(rule.Path) {
				fp = filepath.Join(rootDir, filePath)
			}

			match, err := filepath.Match(rule.Path, fp)
			if err != nil {
				log.Errorf("%s - %q", err, rule.Path)
				continue
			}
			ruleResults = append(ruleResults, match)
			if match {
				log.Debugf("file path %q matching rule %q", fp, rule.Path)
			}
		}

		if len(rule.Packages) > 0 {
			match := false

		outPackage:
			for rulePkgName, rulePkgVersion := range rule.Packages {
				if packageName == rulePkgName {
					if rulePkgVersion == "" {
						match = true
						break outPackage
					}

					v, err := gopep440.Parse(packageVersion)
					if err != nil {
						match = packageVersion == rulePkgVersion
						log.Debugf("%q - %s", packageVersion, err)
						break outPackage
					}

					specifiers, err := gopep440.NewSpecifiers(rulePkgVersion)
					if err != nil {
						match = packageVersion == rulePkgVersion
						log.Debugf("%q %s", err, rulePkgVersion)
						break outPackage
					}

					match = specifiers.Check(v)
					break outPackage
				}
			}
			ruleResults = append(ruleResults, match)
		}

		if len(rule.DependencyGroups) > 0 {
			match := false
		outGroup:
			for _, dependencyGroup := range dependencyGroups {
				for _, pattern := range rule.DependencyGroups {
					// Patterns were checked by Validate, so the error can be ignored.
					if ok, _ := filepath.Match(pattern, dependencyGroup); ok {
						match = true
						break outGroup
					}
				}
			}
			ruleResults = append(ruleResults, match)
		}

		// All conditions in this rule must pass (AND semantics).
		isAllMatching := true
		for _, r := range ruleResults {
			if !r {
				isAllMatching = false
				break
			}
		}
		if isAllMatching && len(ruleResults) > 0 {
			return true
		}
	}

	return false
}
