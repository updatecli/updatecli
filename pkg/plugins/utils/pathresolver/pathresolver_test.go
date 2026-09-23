package pathresolver

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

// TestResolver_Resolve covers every combination of base directory and boundary.
//
// The boundary cases cover the containment check that protects every file based resource
// against the path traversal and arbitrary file write reported in GHSA-hj4x-hm4v-7wpw.
func TestResolver_Resolve(t *testing.T) {
	absolutePath := absoluteTestPath()
	workingDir := filepath.Join("tmp", "updatecli", "checkout")
	manifestDir := filepath.Join("repo", "updatecli.d")

	testdata := []struct {
		name           string
		pathResolver   Resolver
		path           string
		expectedResult string
		wantErr        bool
	}{
		// No base directory and no boundary, as in a default local run: every path
		// stays as the manifest wrote it.
		{
			name:           "local run keeps a relative path untouched",
			pathResolver:   Resolver{},
			path:           filepath.Join("charts", "values.yaml"),
			expectedResult: filepath.Join("charts", "values.yaml"),
		},
		{
			name:           "local run keeps an absolute path untouched",
			pathResolver:   Resolver{},
			path:           absolutePath,
			expectedResult: absolutePath,
		},
		{
			name:           "local run keeps a dot dot path untouched",
			pathResolver:   Resolver{},
			path:           filepath.Join("..", "..", "etc", "passwd"),
			expectedResult: filepath.Join("..", "..", "etc", "passwd"),
		},
		// A base directory without a boundary: manifest relative resolution.
		{
			name:           "manifest relative path resolves against the base directory",
			pathResolver:   Resolver{BaseDir: manifestDir},
			path:           "package.json",
			expectedResult: filepath.Join(manifestDir, "package.json"),
		},
		{
			name:           "manifest relative resolution leaves an absolute path alone",
			pathResolver:   Resolver{BaseDir: manifestDir},
			path:           absolutePath,
			expectedResult: absolutePath,
		},
		{
			name:           "manifest relative resolution allows escaping its own directory",
			pathResolver:   Resolver{BaseDir: manifestDir},
			path:           filepath.Join("..", "package.json"),
			expectedResult: filepath.Join("repo", "package.json"),
		},
		// A boundary: the scm checkout, where containment is enforced.
		{
			name:           "scm relative path resolves inside the checkout",
			pathResolver:   Resolver{BaseDir: workingDir, Boundary: workingDir},
			path:           filepath.Join("charts", "values.yaml"),
			expectedResult: filepath.Join(workingDir, "charts", "values.yaml"),
		},
		{
			name:         "scm rejects an absolute path",
			pathResolver: Resolver{BaseDir: workingDir, Boundary: workingDir},
			path:         absolutePath,
			wantErr:      true,
		},
		{
			name:         "scm rejects a dot dot escape",
			pathResolver: Resolver{BaseDir: workingDir, Boundary: workingDir},
			path:         filepath.Join("..", "..", "etc", "passwd"),
			wantErr:      true,
		},
		{
			name:         "scm rejects an escape hidden behind a subdirectory",
			pathResolver: Resolver{BaseDir: workingDir, Boundary: workingDir},
			path:         filepath.Join("charts", "..", "..", "escaped.yaml"),
			wantErr:      true,
		},
		{
			name:           "scm allows a dot dot that stays inside the checkout",
			pathResolver:   Resolver{BaseDir: workingDir, Boundary: workingDir},
			path:           filepath.Join("charts", "..", "values.yaml"),
			expectedResult: filepath.Join(workingDir, "values.yaml"),
		},
		// Remote locations are fetched over the network, never joined.
		{
			name:           "https url is left untouched inside a checkout",
			pathResolver:   Resolver{BaseDir: workingDir, Boundary: workingDir},
			path:           "https://nodejs.org/dist/index.json",
			expectedResult: "https://nodejs.org/dist/index.json",
		},
		{
			name:           "http url is left untouched with a base directory",
			pathResolver:   Resolver{BaseDir: manifestDir},
			path:           "http://example.com/index.json",
			expectedResult: "http://example.com/index.json",
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, gotErr := tt.pathResolver.Resolve(tt.path)

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
	pathResolver := Resolver{BaseDir: workingDir, Boundary: workingDir}

	gotResults, gotErr := pathResolver.ResolveAll([]string{"a.yaml", filepath.Join("sub", "b.yaml")})
	require.NoError(t, gotErr)
	assert.Equal(t, []string{
		filepath.Join(workingDir, "a.yaml"),
		filepath.Join(workingDir, "sub", "b.yaml"),
	}, gotResults)

	_, gotErr = pathResolver.ResolveAll([]string{"a.yaml", filepath.Join("..", "..", "escaped.yaml")})
	require.Error(t, gotErr)
}

// TestResolver_Join covers the locations Updatecli only needs in order to find
// something, which are resolved but never held inside the boundary.
func TestResolver_Join(t *testing.T) {
	absolutePath := absoluteTestPath()
	workingDir := filepath.Join("tmp", "updatecli", "checkout")

	// An empty path resolves to the base directory itself, like filepath.Join does.
	// Returning "" instead would silently drop the SCM checkout directory for every
	// caller whose path is optional.
	assert.Equal(t, workingDir, Resolver{BaseDir: workingDir}.Join(""))
	assert.Equal(t, "", Resolver{}.Join(""))
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

// TestResolver_Dir checks that Dir falls back to the directory updatecli was
// started from.
func TestResolver_Dir(t *testing.T) {
	workingDirectory, err := os.Getwd()
	require.NoError(t, err)

	assert.Equal(t, workingDirectory, Resolver{}.Dir())
	assert.Equal(t, "somewhere", Resolver{BaseDir: "somewhere"}.Dir())
}

type fakeScm struct{ directory string }

func (f fakeScm) GetDirectory() string { return f.directory }

// TestResolver_JoinManifest checks that settings overriding the scm, such as the git
// resources "path", resolve from the manifest directory and never from the checkout.
func TestResolver_JoinManifest(t *testing.T) {
	absolutePath := absoluteTestPath()
	checkout := filepath.Join("tmp", "updatecli", "checkout")
	parentRepo := filepath.Join("..", "other-repo")

	// Default mode: no manifest directory, so the path is used as written.
	assert.Equal(t, parentRepo, New(fakeScm{directory: checkout}, "").JoinManifest(parentRepo))
	assert.Equal(t, parentRepo, New(nil, "").JoinManifest(parentRepo))

	// Manifest mode: the path resolves from the manifest directory, with or without an scm.
	assert.Equal(t, "other-repo", New(fakeScm{directory: checkout}, "updatecli.d").JoinManifest(parentRepo))
	assert.Equal(t, "other-repo", New(nil, "updatecli.d").JoinManifest(parentRepo))

	assert.Equal(t, absolutePath, New(fakeScm{directory: checkout}, "updatecli.d").JoinManifest(absolutePath))
}

// TestResolver_JoinRooted checks that an absolute path stays under the SCM checkout, and
// that JoinRooted behaves like Join everywhere else.
func TestResolver_JoinRooted(t *testing.T) {
	absolutePath := absoluteTestPath()
	checkout := filepath.Join("tmp", "updatecli", "checkout")
	withScm := New(fakeScm{directory: checkout}, "")

	assert.Equal(t, filepath.Join(checkout, absolutePath), withScm.JoinRooted(absolutePath))
	assert.Equal(t, filepath.Join(checkout, "Dockerfile"), withScm.JoinRooted("Dockerfile"))
	assert.Equal(t, filepath.Join("tmp", "updatecli", "sibling"), withScm.JoinRooted(filepath.Join("..", "sibling")))
	assert.Equal(t, "https://example.com/x", withScm.JoinRooted("https://example.com/x"))

	// Without an scm there is no checkout to root the path at.
	assert.Equal(t, absolutePath, New(nil, "updatecli.d").JoinRooted(absolutePath))
	assert.Equal(t, filepath.Join("updatecli.d", "Dockerfile"), New(nil, "updatecli.d").JoinRooted("Dockerfile"))
}

// TestResolver_RepositoryDir checks that the implicit git repository is the SCM checkout or
// the process working directory, and never the manifest directory.
func TestResolver_RepositoryDir(t *testing.T) {
	workingDirectory, err := os.Getwd()
	require.NoError(t, err)

	checkout := filepath.Join("tmp", "updatecli", "checkout")

	assert.Equal(t, checkout, New(fakeScm{directory: checkout}, "updatecli.d").RepositoryDir())
	assert.Equal(t, workingDirectory, New(nil, "updatecli.d").RepositoryDir())
	assert.Equal(t, workingDirectory, New(nil, "").RepositoryDir())
}
