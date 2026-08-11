package httpclient_test

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// forbiddenPatterns lists HTTP usage patterns that bypass the unified client.
// All production code must go through pkg/core/httpclient instead.
var forbiddenPatterns = []string{
	"&http.Client{",
	"http.DefaultClient",
	"http.DefaultTransport",
	"http.Get(",
	"http.Post(",
	"http.Head(",
	"http.PostForm(",
}

// TestNoDirectHTTPClientUsage scans all Go source files in the repository and
// fails if any production code constructs or uses an HTTP client directly instead
// of going through the unified httpclient package.
func TestNoDirectHTTPClientUsage(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed: cannot determine test file path")
	}

	// This file lives at pkg/core/httpclient/lint_test.go, so repo root is three levels up.
	repoRoot := filepath.Join(filepath.Dir(currentFile), "..", "..", "..")
	repoRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		t.Fatalf("resolving repo root: %v", err)
	}

	var violations []string

	err = filepath.Walk(repoRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if shouldSkipDir(info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip test files — httptest servers legitimately use bare http.Client.
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Skip the httpclient package itself; it must define the unified client.
		rel, relErr := filepath.Rel(repoRoot, path)
		if relErr == nil && strings.HasPrefix(rel, filepath.Join("pkg", "core", "httpclient")) {
			return nil
		}

		fileViolations, scanErr := scanFile(path, repoRoot)
		if scanErr != nil {
			// Non-fatal: log the error but continue scanning.
			t.Logf("warning: could not scan %s: %v", rel, scanErr)
			return nil
		}
		violations = append(violations, fileViolations...)
		return nil
	})
	if err != nil {
		t.Fatalf("walking repository: %v", err)
	}

	for _, v := range violations {
		t.Errorf("%s", v)
	}
}

// scanFile reads path line by line and returns a violation string for each line
// that contains a forbidden pattern and is not a comment.
func scanFile(path, repoRoot string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	rel, err := filepath.Rel(repoRoot, path)
	if err != nil {
		rel = path
	}

	var violations []string
	scanner := bufio.NewScanner(f)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "//") {
			continue
		}

		// Strip trailing inline comments to avoid false positives.
		code := line
		if idx := strings.Index(code, " //"); idx >= 0 {
			code = code[:idx]
		}

		for _, pattern := range forbiddenPatterns {
			if strings.Contains(code, pattern) {
				violations = append(violations, fmt.Sprintf(
					"%s:%d: forbidden HTTP pattern %q — use pkg/core/httpclient instead: %s",
					rel, lineNum, pattern, trimmed,
				))
			}
		}
	}

	return violations, scanner.Err()
}

// shouldSkipDir returns true for directories that must never be scanned.
func shouldSkipDir(name string) bool {
	switch name {
	case "vendor", ".git", "testdata":
		return true
	}
	return false
}
