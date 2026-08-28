package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// absoluteTestPath returns an OS appropriate absolute path so the absolute-path
// rules are exercised identically on POSIX and Windows.
func absoluteTestPath() string {
	if runtime.GOOS == "windows" {
		return `C:\Windows\Temp\evil`
	}
	return "/etc/cron.d/evil"
}

// TestResolver_Resolve covers the whole base directory x boundary matrix.
//
// The boundary half is the containment primitive protecting every file based resource
// against the path traversal / arbitrary file write reported in GHSA-hj4x-hm4v-7wpw.
func TestResolver_Resolve(t *testing.T) {
	absolutePath := absoluteTestPath()
	workingDir := filepath.Join("tmp", "updatecli", "checkout")
	manifestDir := filepath.Join("repo", "updatecli.d")

	testdata := []struct {
		name           string
		resolver       Resolver
		path           string
		expectedResult string
		wantErr        bool
	}{
		// No base directory and no boundary: the historical local run, every path
		// is left exactly as the manifest wrote it.
		{
			name:           "local run keeps a relative path untouched",
			resolver:       Resolver{},
			path:           filepath.Join("charts", "values.yaml"),
			expectedResult: filepath.Join("charts", "values.yaml"),
		},
		{
			name:           "local run keeps an absolute path untouched",
			resolver:       Resolver{},
			path:           absolutePath,
			expectedResult: absolutePath,
		},
		{
			name:           "local run keeps a dot dot path untouched",
			resolver:       Resolver{},
			path:           filepath.Join("..", "..", "etc", "passwd"),
			expectedResult: filepath.Join("..", "..", "etc", "passwd"),
		},
		// A base directory without a boundary: manifest relative resolution.
		{
			name:           "manifest relative path resolves against the base directory",
			resolver:       Resolver{BaseDir: manifestDir},
			path:           "package.json",
			expectedResult: filepath.Join(manifestDir, "package.json"),
		},
		{
			name:           "manifest relative resolution leaves an absolute path alone",
			resolver:       Resolver{BaseDir: manifestDir},
			path:           absolutePath,
			expectedResult: absolutePath,
		},
		{
			name:           "manifest relative resolution allows escaping its own directory",
			resolver:       Resolver{BaseDir: manifestDir},
			path:           filepath.Join("..", "package.json"),
			expectedResult: filepath.Join("repo", "package.json"),
		},
		// A boundary: the scm checkout, where containment is enforced.
		{
			name:           "scm relative path resolves inside the checkout",
			resolver:       Resolver{BaseDir: workingDir, Boundary: workingDir},
			path:           filepath.Join("charts", "values.yaml"),
			expectedResult: filepath.Join(workingDir, "charts", "values.yaml"),
		},
		{
			name:     "scm rejects an absolute path",
			resolver: Resolver{BaseDir: workingDir, Boundary: workingDir},
			path:     absolutePath,
			wantErr:  true,
		},
		{
			name:     "scm rejects a dot dot escape",
			resolver: Resolver{BaseDir: workingDir, Boundary: workingDir},
			path:     filepath.Join("..", "..", "etc", "passwd"),
			wantErr:  true,
		},
		{
			name:     "scm rejects an escape hidden behind a subdirectory",
			resolver: Resolver{BaseDir: workingDir, Boundary: workingDir},
			path:     filepath.Join("charts", "..", "..", "escaped.yaml"),
			wantErr:  true,
		},
		{
			name:           "scm allows a dot dot that stays inside the checkout",
			resolver:       Resolver{BaseDir: workingDir, Boundary: workingDir},
			path:           filepath.Join("charts", "..", "values.yaml"),
			expectedResult: filepath.Join(workingDir, "values.yaml"),
		},
		// Remote locations are fetched over the network, never joined.
		{
			name:           "https url is left untouched inside a checkout",
			resolver:       Resolver{BaseDir: workingDir, Boundary: workingDir},
			path:           "https://nodejs.org/dist/index.json",
			expectedResult: "https://nodejs.org/dist/index.json",
		},
		{
			name:           "http url is left untouched with a base directory",
			resolver:       Resolver{BaseDir: manifestDir},
			path:           "http://example.com/index.json",
			expectedResult: "http://example.com/index.json",
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, gotErr := tt.resolver.Resolve(tt.path)

			if tt.wantErr {
				require.Error(t, gotErr)
				assert.Empty(t, gotResult)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.expectedResult, gotResult)
		})
	}
}

// TestResolver_ResolveAll ensures a list fails as a whole on its first bad path.
func TestResolver_ResolveAll(t *testing.T) {
	workingDir := filepath.Join("tmp", "updatecli", "checkout")
	resolver := Resolver{BaseDir: workingDir, Boundary: workingDir}

	gotResults, gotErr := resolver.ResolveAll([]string{"a.yaml", filepath.Join("sub", "b.yaml")})
	require.NoError(t, gotErr)
	assert.Equal(t, []string{
		filepath.Join(workingDir, "a.yaml"),
		filepath.Join(workingDir, "sub", "b.yaml"),
	}, gotResults)

	_, gotErr = resolver.ResolveAll([]string{"a.yaml", filepath.Join("..", "..", "escaped.yaml")})
	require.Error(t, gotErr)
}

// TestResolver_Join covers the locations Updatecli only needs in order to find
// something, which are resolved but never held inside the boundary.
func TestResolver_Join(t *testing.T) {
	absolutePath := absoluteTestPath()
	workingDir := filepath.Join("tmp", "updatecli", "checkout")

	assert.Equal(t, "", Resolver{BaseDir: workingDir}.Join(""))
	assert.Equal(t, "charts", Resolver{}.Join("charts"))
	assert.Equal(t, filepath.Join(workingDir, "charts"), Resolver{BaseDir: workingDir}.Join("charts"))
	assert.Equal(t, absolutePath, Resolver{BaseDir: workingDir}.Join(absolutePath))
	assert.Equal(t, "https://example.com/x", Resolver{BaseDir: workingDir}.Join("https://example.com/x"))
	// Unlike Resolve, escaping the boundary is allowed: a repository or a shell
	// working directory legitimately sits outside of the checkout.
	assert.Equal(t,
		filepath.Join("tmp", "updatecli", "sibling"),
		Resolver{BaseDir: workingDir, Boundary: workingDir}.Join(filepath.Join("..", "sibling")))
}

// TestResolver_Dir falls back to the process working directory, which is what
// "resolve relative to wherever updatecli was started from" means.
func TestResolver_Dir(t *testing.T) {
	workingDirectory, err := os.Getwd()
	require.NoError(t, err)

	assert.Equal(t, workingDirectory, Resolver{}.Dir())
	assert.Equal(t, "somewhere", Resolver{BaseDir: "somewhere"}.Dir())
}
