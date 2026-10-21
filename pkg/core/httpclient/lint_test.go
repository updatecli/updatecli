package httpclient_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// TestNoDirectHTTPClientUsage fails if production code builds its own
// net/http Client or Transport instead of going through the unified httpclient
// package.
//
// The forbidigo rule in .golangci.yml bans the net/http helpers. It cannot ban
// building a Client or Transport without also banning the Client type, which
// the httpclient constructors return, and the Client.Transport field, so this
// test covers literals and new() calls of both types only.
func TestNoDirectHTTPClientUsage(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed: cannot determine test file path")
	}

	// This file lives at pkg/core/httpclient/lint_test.go, so repo root is three levels up.
	repoRoot, err := filepath.Abs(filepath.Join(filepath.Dir(currentFile), "..", "..", ".."))
	if err != nil {
		t.Fatalf("resolving repo root: %v", err)
	}

	err = filepath.WalkDir(repoRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if shouldSkipDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		// Test files talk to httptest servers with a bare client.
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		rel, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}

		// The httpclient package must build the unified client itself.
		if strings.HasPrefix(rel, filepath.Join("pkg", "core", "httpclient")+string(filepath.Separator)) {
			return nil
		}

		for _, v := range scanFile(t, path, rel) {
			t.Error(v)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking repository: %v", err)
	}
}

// scanFile returns a violation for each http.Client or http.Transport literal
// or new() call in path, whatever name net/http is imported under.
func scanFile(t *testing.T, path, rel string) []string {
	t.Helper()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
	if err != nil {
		t.Logf("warning: could not parse %s: %v", rel, err)
		return nil
	}

	pkgName := netHTTPName(file)
	if pkgName == "" {
		return nil
	}

	isForbidden := func(expr ast.Expr) bool {
		sel, ok := expr.(*ast.SelectorExpr)
		if !ok || (sel.Sel.Name != "Client" && sel.Sel.Name != "Transport") {
			return false
		}
		id, ok := sel.X.(*ast.Ident)
		return ok && id.Name == pkgName
	}

	var violations []string
	ast.Inspect(file, func(n ast.Node) bool {
		var found bool
		switch n := n.(type) {
		case *ast.CompositeLit:
			found = isForbidden(n.Type)
		case *ast.CallExpr:
			id, ok := n.Fun.(*ast.Ident)
			found = ok && id.Name == "new" && len(n.Args) == 1 && isForbidden(n.Args[0])
		}
		if found {
			violations = append(violations,
				fmt.Sprintf("%s:%d: building an http.Client or http.Transport directly, use pkg/core/httpclient instead", rel, fset.Position(n.Pos()).Line))
		}
		return true
	})

	return violations
}

// netHTTPName returns the name net/http is imported under in file, or "" if
// file does not import it.
func netHTTPName(file *ast.File) string {
	for _, imp := range file.Imports {
		if p, err := strconv.Unquote(imp.Path.Value); err != nil || p != "net/http" {
			continue
		}
		if imp.Name != nil {
			return imp.Name.Name
		}
		return "http"
	}
	return ""
}

// shouldSkipDir returns true for directories that must never be scanned.
func shouldSkipDir(name string) bool {
	switch name {
	case "vendor", ".git", "testdata":
		return true
	}
	return false
}
