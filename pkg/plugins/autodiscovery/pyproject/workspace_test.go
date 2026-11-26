package pyproject

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchUvLock(t *testing.T) {
	root := "testdata/uv_workspace"

	testdata := []struct {
		name     string
		dir      string
		rootDir  string
		expected string
	}{
		{name: "the workspace root uses its uv.lock", dir: root, rootDir: root, expected: root},
		{name: "a member uses the workspace root uv.lock", dir: root + "/packages/api", rootDir: root, expected: root},
		// As in uv, the closest pyproject.toml decides, even when an outer workspace matches the project
		{name: "a project nested in another project is its own root", dir: root + "/libs/x/sub", rootDir: root, expected: ""},
		{name: "a project with its own uv.lock uses it", dir: root + "/libs/x", rootDir: root, expected: root + "/libs/x"},
		{name: "a project the workspace doesn't include has no uv.lock", dir: root + "/tools/standalone", rootDir: root, expected: ""},
		{name: "a project the workspace excludes has no uv.lock", dir: root + "/packages/legacy", rootDir: root, expected: ""},
		{name: "the search doesn't go above the root directory", dir: root + "/packages/api", rootDir: root + "/packages", expected: ""},
		{name: "a uv.lock next to the pyproject.toml always applies", dir: "testdata/simple_project", rootDir: "testdata", expected: "testdata/simple_project"},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, searchUvLock(tt.dir, tt.rootDir))
		})
	}
}

func TestMatchGlobSegments(t *testing.T) {
	testdata := []struct {
		pattern  string
		path     string
		expected bool
	}{
		{pattern: "packages/*", path: "packages/api", expected: true},
		{pattern: "packages/*", path: "packages/api/sub", expected: false},
		{pattern: "libs/**/sub", path: "libs/sub", expected: true},
		{pattern: "libs/**/sub", path: "libs/x/y/sub", expected: true},
		{pattern: "libs/**/sub", path: "libs/x", expected: false},
	}

	for _, tt := range testdata {
		t.Run(tt.pattern+" "+tt.path, func(t *testing.T) {
			assert.Equal(t, tt.expected, matchGlobSegments(strings.Split(tt.pattern, "/"), strings.Split(tt.path, "/")))
		})
	}
}

func TestDiscoverManifestsUvWorkspace(t *testing.T) {
	p, err := New(Spec{}, "testdata/uv_workspace", "", "")
	require.NoError(t, err)
	p.uvAvailable = true

	manifests, err := p.DiscoverManifests()
	require.NoError(t, err)
	byName := manifestsByName(t, manifests)
	require.Len(t, byName, 3, "manifests: %v", byName)

	// requests, declared by the root and a member, is re-locked by a single manifest rather than
	// by one per project opening conflicting pull requests. It belongs to the first project the
	// walk finds, packages/api, as "packages" sorts before "pyproject.toml". Its version must
	// satisfy both declarations, ">=2.28" and ">=2.28,<2.33".
	requests, found := byName[`'deps(pypi): bump "requests" for "api" project'`]
	require.True(t, found, "manifests: %v", byName)
	assert.Contains(t, requests, "pattern: '>=2.32.5, >=2.28,>=2.28,<2.33'\n")

	// A workspace member is bounded at the version the root uv.lock resolves, and uv runs at the root
	fastapi := byName[`'deps(pypi): bump "fastapi" for "api" project'`]
	assert.Contains(t, fastapi, "pattern: '>=0.115.0, >=0.100'\n")
	assert.Contains(t, fastapi, "workdir: '.'\n")

	// A project the workspace doesn't include gets no uv target and no bound
	flask := byName[`'deps(pypi): bump "flask" for "standalone" project'`]
	assert.Contains(t, flask, "pattern: '>=3.0'\n")
	assert.NotContains(t, flask, "targets:")
}

func TestDiscoverManifestsUvWorkspaceMemberRootDir(t *testing.T) {
	// "rootdir" pointing at a workspace member still finds the uv.lock at the workspace root
	p, err := New(Spec{RootDir: "packages/api"}, "testdata/uv_workspace", "", "")
	require.NoError(t, err)
	p.uvAvailable = true

	manifests, err := p.DiscoverManifests()
	require.NoError(t, err)
	byName := manifestsByName(t, manifests)
	require.Len(t, byName, 2, "manifests: %v", byName)

	for name, manifest := range byName {
		assert.Contains(t, manifest, "workdir: '.'\n", name)
		assert.Contains(t, manifest, "uv lock --upgrade-package", name)
	}

	// Only the member is discovered, so only its declaration of requests constrains it
	assert.Contains(t, byName[`'deps(pypi): bump "fastapi" for "api" project'`], "pattern: '>=0.115.0, >=0.100'\n")
	assert.Contains(t, byName[`'deps(pypi): bump "requests" for "api" project'`], "pattern: '>=2.32.5, >=2.28,<2.33'\n")
}

func TestDiscoverSecurityManifestsUvWorkspace(t *testing.T) {
	p, err := New(map[string]interface{}{"vulnerability": map[string]interface{}{}}, "testdata/uv_workspace", "", "")
	require.NoError(t, err)
	p.uvAvailable = true

	manifests, err := p.DiscoverManifests()
	require.NoError(t, err)
	byName := manifestsByName(t, manifests)

	// requests, declared by the root and a member, is updated once in the shared uv.lock
	require.Len(t, byName, 2, "manifests: %v", byName)
	for name, manifest := range byName {
		assert.Contains(t, manifest, "workdir: '.'\n", name)
	}

	assert.Contains(t, byName[`'deps(pypi): bump "fastapi" to {{ source "fastapi" }}'`], "version: '0.115.0'\n")
	assert.Contains(t, byName[`'deps(pypi): bump "requests" to {{ source "requests" }}'`], "version: '2.32.5'\n")
}

func TestLockStateCombinedConstraint(t *testing.T) {
	lock := newLockState()
	for _, file := range []string{"testdata/uv_workspace/packages/api/pyproject.toml", "testdata/uv_workspace/pyproject.toml"} {
		data, err := loadPyprojectData(file)
		require.NoError(t, err)
		lock.addConstraints(data)
	}
	lock.constraints["foo"] = []string{">=2", "<2"}

	testdata := []struct {
		name     string
		pkg      string
		locked   string
		expected string
	}{
		// Every distinct declaration counts, in a stable order, so the source picks a version all accept
		{name: "declarations are joined", pkg: "Requests", expected: ">=2.28,>=2.28,<2.33"},
		{name: "declarations the locked version satisfies are joined", pkg: "requests", locked: "2.32.5", expected: ">=2.28,>=2.28,<2.33"},
		{name: "a single declaration", pkg: "fastapi", expected: ">=0.100"},
		{name: "an unknown package has no constraint", pkg: "unknown", expected: ""},
		// A declaration uv doesn't apply, which the locked version doesn't satisfy, is left out
		{name: "a declaration the locked version doesn't satisfy is left out", pkg: "foo", locked: "2.5", expected: ">=2"},
		{name: "without a locked version, every declaration is kept", pkg: "foo", expected: "<2,>=2"},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, lock.combinedConstraint(tt.pkg, tt.locked))
		})
	}
}
