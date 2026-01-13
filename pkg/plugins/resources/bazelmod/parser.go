package bazelmod

import (
	"fmt"
	"regexp"
	"strings"
)

// BazelDep represents a bazel_dep() function call
type BazelDep struct {
	Name     string
	Version  string
	RepoName string // Optional repo_name parameter
	LineNum  int    // Line number in the original file (1-indexed)
	FullLine string // The complete line(s) for this bazel_dep call
}

// ModuleFile represents a parsed MODULE.bazel file
type ModuleFile struct {
	Content  string
	Deps     []BazelDep
	Lines    []string
}

var (
	// Regex pattern to match bazel_dep() calls
	// Matches: bazel_dep(name = "module_name", version = "1.2.3")
	// Also handles multi-line calls and optional repo_name
	bazelDepPattern = regexp.MustCompile(`(?m)^\s*bazel_dep\s*\(\s*(?:name\s*=\s*"([^"]+)"\s*,?\s*)?(?:version\s*=\s*"([^"]+)"\s*,?\s*)?(?:repo_name\s*=\s*"([^"]+)"\s*,?\s*)?\)`)
	
	// Pattern to find the start of a bazel_dep call (may span multiple lines)
	bazelDepStartPattern = regexp.MustCompile(`(?m)^\s*bazel_dep\s*\(`)
)

// ParseModuleFile parses a MODULE.bazel file and extracts all bazel_dep() calls
func ParseModuleFile(content string) (*ModuleFile, error) {
	lines := strings.Split(content, "\n")
	moduleFile := &ModuleFile{
		Content: content,
		Lines:   lines,
		Deps:    []BazelDep{},
	}

	// Find all bazel_dep calls
	// We need to handle multi-line calls, so we'll search for opening parentheses
	// and then parse until the closing parenthesis
	
	i := 0
	for i < len(lines) {
		line := lines[i]
		
		// Check if this line starts a bazel_dep call
		if bazelDepStartPattern.MatchString(line) {
			dep, endLine, err := parseBazelDep(lines, i)
			if err != nil {
				return nil, fmt.Errorf("parsing bazel_dep at line %d: %w", i+1, err)
			}
			if dep != nil && dep.Name != "" {
				moduleFile.Deps = append(moduleFile.Deps, *dep)
			}
			i = endLine
		} else {
			i++
		}
	}

	return moduleFile, nil
}

// parseBazelDep parses a bazel_dep() call that may span multiple lines
func parseBazelDep(lines []string, startLine int) (*BazelDep, int, error) {
	dep := &BazelDep{
		LineNum: startLine + 1,
	}
	
	// Collect all lines until we find the closing parenthesis
	var depLines []string
	parenCount := 0
	i := startLine
	
	for i < len(lines) {
		line := lines[i]
		depLines = append(depLines, line)
		
		// Count parentheses to find the end of the function call
		for _, char := range line {
			if char == '(' {
				parenCount++
			} else if char == ')' {
				parenCount--
				if parenCount == 0 {
					// Found the closing parenthesis
					dep.FullLine = strings.Join(depLines, "\n")
					
					// Extract name, version, and repo_name using regex
					content := strings.Join(depLines, " ")
					matches := bazelDepPattern.FindStringSubmatch(content)
					
					if len(matches) >= 3 {
						dep.Name = matches[1]
						dep.Version = matches[2]
						if len(matches) >= 4 && matches[3] != "" {
							dep.RepoName = matches[3]
						}
					} else {
						// Try a more flexible parsing approach
						dep.Name = extractParam(content, "name")
						dep.Version = extractParam(content, "version")
						dep.RepoName = extractParam(content, "repo_name")
					}
					
					return dep, i + 1, nil
				}
			}
		}
		
		i++
	}
	
	// If we get here, we didn't find a closing parenthesis
	return nil, startLine + 1, fmt.Errorf("unclosed bazel_dep() call starting at line %d", startLine+1)
}

// extractParam extracts a parameter value from a bazel_dep call string
func extractParam(content, paramName string) string {
	// Pattern: paramName = "value"
	pattern := regexp.MustCompile(fmt.Sprintf(`%s\s*=\s*"([^"]+)"`, regexp.QuoteMeta(paramName)))
	matches := pattern.FindStringSubmatch(content)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

// FindDepByName finds a bazel_dep by module name
func (mf *ModuleFile) FindDepByName(name string) *BazelDep {
	for i := range mf.Deps {
		if mf.Deps[i].Name == name {
			return &mf.Deps[i]
		}
	}
	return nil
}

// UpdateDepVersion updates the version of a specific module in the file content
func (mf *ModuleFile) UpdateDepVersion(moduleName, newVersion string) (string, error) {
	dep := mf.FindDepByName(moduleName)
	if dep == nil {
		return "", fmt.Errorf("module %q not found in MODULE.bazel", moduleName)
	}

	// Replace the version in the full bazel_dep line(s)
	// Pattern: version = "old_version" (handles whitespace variations)
	versionPattern := regexp.MustCompile(`(version\s*=\s*")[^"]+(")`)
	
	// Replace only once (the version in this specific bazel_dep)
	updatedDepLines := versionPattern.ReplaceAllStringFunc(dep.FullLine, func(match string) string {
		// Extract the quotes and replace the version
		parts := versionPattern.FindStringSubmatch(match)
		if len(parts) >= 3 {
			return fmt.Sprintf(`%s%s%s`, parts[1], newVersion, parts[2])
		}
		return match
	})
	
	// Replace in the original content
	lines := mf.Lines
	startLine := dep.LineNum - 1
	
	// Find the range of lines for this dep
	depLines := strings.Split(dep.FullLine, "\n")
	endLine := startLine + len(depLines) - 1
	
	// Reconstruct the file with the updated lines
	var newLines []string
	newLines = append(newLines, lines[:startLine]...)
	newLines = append(newLines, strings.Split(updatedDepLines, "\n")...)
	if endLine+1 < len(lines) {
		newLines = append(newLines, lines[endLine+1:]...)
	}
	
	return strings.Join(newLines, "\n"), nil
}

