package pyproject

import (
	"fmt"
	"path/filepath"

	gopep440 "github.com/aquasecurity/go-pep440-version"
	"github.com/sirupsen/logrus"
)

// MatchingRule specifies a rule to include or exclude pyproject.toml dependencies.
type MatchingRule struct {
	// Path specifies a pyproject.toml path pattern. The pattern must match the full path,
	// not just a substring. Wildcards accepted by filepath.Match are supported.
	Path string `yaml:",omitempty"`
	// Packages specifies the list of Python packages to match, keyed by package name.
	// The value is a PEP 440 version specifier (e.g. ">=2.0,<3.0") or empty to match any version.
	Packages map[string]string `yaml:",omitempty"`
}

// MatchingRules is a slice of MatchingRule.
type MatchingRules []MatchingRule

// Validate checks that every rule has at least one non-empty field.
func (m MatchingRules) Validate() error {
	for i, rule := range m {
		if rule.Path == "" && len(rule.Packages) == 0 {
			return fmt.Errorf("rule %d has no valid fields (path or packages must be specified)", i+1)
		}
	}
	return nil
}

// isMatchingRules reports whether the given file/package pair matches any rule in the list.
// Multiple conditions within one rule are AND-ed; multiple rules are OR-ed.
func (m MatchingRules) isMatchingRules(rootDir, filePath, packageName, packageVersion string) bool {
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
				logrus.Errorf("%s - %q", err, rule.Path)
				continue
			}
			ruleResults = append(ruleResults, match)
			if match {
				logrus.Debugf("file path %q matching rule %q", fp, rule.Path)
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
						logrus.Debugf("%q - %s", packageVersion, err)
						break outPackage
					}

					specifiers, err := gopep440.NewSpecifiers(rulePkgVersion)
					if err != nil {
						match = packageVersion == rulePkgVersion
						logrus.Debugf("%q %s", err, rulePkgVersion)
						break outPackage
					}

					match = specifiers.Check(v)
					break outPackage
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
