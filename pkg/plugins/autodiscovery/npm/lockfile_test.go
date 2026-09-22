package npm

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type lockedVersionTest struct {
	name       string
	constraint string
	expected   string
}

func assertLockedVersions(t *testing.T, versions lockedVersions, tests []lockedVersionTest) {
	t.Helper()

	for _, tt := range tests {
		assert.Equal(t, tt.expected, versions.version(tt.name, tt.constraint), "%s@%s", tt.name, tt.constraint)
	}
}

// writeRootPackageJson writes a package.json to a temporary directory and returns the path of
// a lock file next to it, so parsers can read the workspaces it declares.
func writeRootPackageJson(t *testing.T, packageJson string) string {
	t.Helper()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "package.json"), []byte(packageJson), 0o600))

	return filepath.Join(dir, "package-lock.json")
}

func TestParsePackageLock(t *testing.T) {
	t.Run("lockfileVersion 3 from testdata", func(t *testing.T) {
		data, err := os.ReadFile("testdata/npmlockfile/package-lock.json")
		require.NoError(t, err)

		versions, err := parsePackageLock("", data, ".")
		require.NoError(t, err)

		assertLockedVersions(t, versions, []lockedVersionTest{
			{name: "axios", constraint: "^1.0.0", expected: "1.2.6"},
			{name: "unknown", constraint: "^1.0.0", expected: ""},
		})
	})

	t.Run("lockfileVersion 3 ignores nested packages", func(t *testing.T) {
		versions, err := parsePackageLock("", []byte(`{
  "lockfileVersion": 3,
  "packages": {
    "": {"name": "project"},
    "node_modules/@mdi/font": {"version": "5.9.55"},
    "node_modules/axios": {"version": "1.2.6"},
    "node_modules/axios/node_modules/follow-redirects": {"version": "1.14.0"},
    "packages/workspace": {"version": "0.1.0"}
  }
}`), ".")
		require.NoError(t, err)

		assertLockedVersions(t, versions, []lockedVersionTest{
			{name: "@mdi/font", constraint: "^5.0.0", expected: "5.9.55"},
			{name: "axios", constraint: "*", expected: "1.2.6"},
			{name: "follow-redirects", constraint: "^1.0.0", expected: ""},
		})
	})

	t.Run("lockfileVersion 1", func(t *testing.T) {
		versions, err := parsePackageLock("", []byte(`{
  "lockfileVersion": 1,
  "dependencies": {
    "axios": {
      "version": "0.21.1",
      "dependencies": {"follow-redirects": {"version": "1.14.0"}}
    }
  }
}`), ".")
		require.NoError(t, err)

		assertLockedVersions(t, versions, []lockedVersionTest{
			{name: "axios", constraint: "^0.21.0", expected: "0.21.1"},
			{name: "follow-redirects", constraint: "^1.0.0", expected: ""},
		})
	})

	t.Run("lockfileVersion 3 with a workspace project", func(t *testing.T) {
		lock := []byte(`{
  "lockfileVersion": 3,
  "packages": {
    "": {"name": "workspace"},
    "node_modules/axios": {"version": "1.2.6"},
    "node_modules/vue": {"version": "3.2.13"},
    "packages/app": {"version": "0.1.0"},
    "packages/app/node_modules/axios": {"version": "1.1.0"},
    "packages/app/node_modules/axios/node_modules/follow-redirects": {"version": "1.14.0"}
  }
}`)

		t.Run("Root project", func(t *testing.T) {
			versions, err := parsePackageLock("", lock, ".")
			require.NoError(t, err)

			assertLockedVersions(t, versions, []lockedVersionTest{
				{name: "axios", constraint: "^1.0.0", expected: "1.2.6"},
			})
		})

		t.Run("Workspace project", func(t *testing.T) {
			lockFile := writeRootPackageJson(t, `{"workspaces": ["packages/*"]}`)

			versions, err := parsePackageLock(lockFile, lock, "packages/app")
			require.NoError(t, err)

			assertLockedVersions(t, versions, []lockedVersionTest{
				// The package installed next to the project takes precedence over the hoisted one
				{name: "axios", constraint: "^1.0.0", expected: "1.1.0"},
				{name: "vue", constraint: "^3.0.0", expected: "3.2.13"},
				{name: "follow-redirects", constraint: "^1.0.0", expected: ""},
			})
		})

		t.Run("Project outside of the workspace", func(t *testing.T) {
			lockFile := writeRootPackageJson(t, `{"workspaces": ["packages/*"]}`)

			_, err := parsePackageLock(lockFile, lock, "tools/foo")
			assert.ErrorIs(t, err, errUnknownImporter)
		})

		t.Run("Local file dependency", func(t *testing.T) {
			// npm records "file:" dependencies by path too, but they aren't installed through the lock file
			lockFile := writeRootPackageJson(t, `{"dependencies": {"app": "file:packages/app"}}`)

			_, err := parsePackageLock(lockFile, lock, "packages/app")
			assert.ErrorIs(t, err, errUnknownImporter)
		})
	})

	t.Run("lockfileVersion 1 without workspace", func(t *testing.T) {
		_, err := parsePackageLock("", []byte(`{
  "lockfileVersion": 1,
  "dependencies": {"axios": {"version": "0.21.1"}}
}`), "tools/foo")
		assert.ErrorIs(t, err, errUnknownImporter)
	})

	t.Run("Invalid file", func(t *testing.T) {
		_, err := parsePackageLock("", []byte(`{`), ".")
		require.Error(t, err)
	})
}

func TestParsePnpmLock(t *testing.T) {
	t.Run("lockfileVersion 9 from testdata", func(t *testing.T) {
		data, err := os.ReadFile("testdata/pnpmlockfile/pnpm-lock.yaml")
		require.NoError(t, err)

		versions, err := parsePnpmLock("", data, ".")
		require.NoError(t, err)

		assertLockedVersions(t, versions, []lockedVersionTest{
			{name: "@mdi/font", constraint: "5.9.55", expected: "5.9.55"},
		})
	})

	t.Run("lockfileVersion 9 ignores other workspace projects", func(t *testing.T) {
		versions, err := parsePnpmLock("", []byte(`lockfileVersion: '9.0'
importers:
  .:
    devDependencies:
      typescript:
        specifier: ^5.0.0
        version: 5.1.6
  packages/app:
    dependencies:
      react:
        specifier: ^17.0.0
        version: 17.0.2
`), ".")
		require.NoError(t, err)

		assertLockedVersions(t, versions, []lockedVersionTest{
			{name: "typescript", constraint: "^5.0.0", expected: "5.1.6"},
			{name: "react", constraint: "^17.0.0", expected: ""},
		})
	})

	t.Run("lockfileVersion 9 from a workspace project", func(t *testing.T) {
		versions, err := parsePnpmLock("", []byte(`lockfileVersion: '9.0'
importers:
  .:
    devDependencies:
      typescript:
        specifier: ^5.0.0
        version: 5.1.6
  packages/app:
    dependencies:
      react:
        specifier: ^17.0.0
        version: 17.0.2
`), "packages/app")
		require.NoError(t, err)

		assertLockedVersions(t, versions, []lockedVersionTest{
			{name: "react", constraint: "^17.0.0", expected: "17.0.2"},
			{name: "typescript", constraint: "^5.0.0", expected: ""},
		})
	})

	t.Run("lockfileVersion 9 from a project outside of the workspace", func(t *testing.T) {
		_, err := parsePnpmLock("", []byte(`lockfileVersion: '9.0'
importers:
  .:
    devDependencies:
      typescript:
        specifier: ^5.0.0
        version: 5.1.6
`), "tools/foo")
		assert.ErrorIs(t, err, errUnknownImporter)
	})

	t.Run("lockfileVersion 6 with peer dependencies", func(t *testing.T) {
		versions, err := parsePnpmLock("", []byte(`lockfileVersion: '6.0'
dependencies:
  react-dom:
    specifier: ^18.0.0
    version: 18.2.0(react@18.2.0)
devDependencies:
  typescript:
    specifier: ^5.0.0
    version: 5.1.6
`), ".")
		require.NoError(t, err)

		assertLockedVersions(t, versions, []lockedVersionTest{
			{name: "react-dom", constraint: "^18.0.0", expected: "18.2.0"},
			{name: "typescript", constraint: "^5.0.0", expected: "5.1.6"},
		})
	})

	t.Run("lockfileVersion 5", func(t *testing.T) {
		versions, err := parsePnpmLock("", []byte(`lockfileVersion: 5.4
specifiers:
  axios: ^1.0.0
dependencies:
  axios: 1.2.6_debug@4.3.4
`), ".")
		require.NoError(t, err)

		assertLockedVersions(t, versions, []lockedVersionTest{
			{name: "axios", constraint: "^1.0.0", expected: "1.2.6"},
		})
	})
}

func TestParseYarnLock(t *testing.T) {
	expected := []lockedVersionTest{
		{name: "@mdi/font", constraint: "5.9.55", expected: "5.9.55"},
		{name: "axios", constraint: "^1.0.0", expected: "1.2.6"},
		{name: "axios", constraint: "^1.1.0", expected: "1.2.6"},
		{name: "axios", constraint: "^2.0.0", expected: ""},
		{name: "follow-redirects", constraint: "^1.15.0", expected: "1.15.2"},
	}

	t.Run("Yarn classic", func(t *testing.T) {
		versions, err := parseYarnLock("", []byte(`# THIS IS AN AUTOGENERATED FILE. DO NOT EDIT THIS FILE DIRECTLY.
# yarn lockfile v1


"@mdi/font@5.9.55":
  version "5.9.55"
  resolved "https://registry.yarnpkg.com/@mdi/font/-/font-5.9.55.tgz"

axios@^1.0.0, axios@^1.1.0:
  version "1.2.6"
  resolved "https://registry.yarnpkg.com/axios/-/axios-1.2.6.tgz"
  dependencies:
    follow-redirects "^1.15.0"
    version "9.9.9"

follow-redirects@^1.15.0:
  version "1.15.2"
`), ".")
		require.NoError(t, err)

		assertLockedVersions(t, versions, expected)
	})

	t.Run("Yarn berry", func(t *testing.T) {
		versions, err := parseYarnLock("", []byte("__metadata:\r\n"+
			"  version: 6\r\n"+
			"  cacheKey: 8\r\n"+
			"\r\n"+
			"\"@mdi/font@npm:5.9.55\":\r\n"+
			"  version: 5.9.55\r\n"+
			"  resolution: \"@mdi/font@npm:5.9.55\"\r\n"+
			"\r\n"+
			"\"axios@npm:^1.0.0, axios@npm:^1.1.0\":\r\n"+
			"  version: 1.2.6\r\n"+
			"  dependencies:\r\n"+
			"    follow-redirects: ^1.15.0\r\n"+
			"\r\n"+
			"\"follow-redirects@npm:^1.15.0\":\r\n"+
			"  version: 1.15.2\r\n"), ".")
		require.NoError(t, err)

		assertLockedVersions(t, versions, expected)
	})

	lock := []byte(`axios@^1.0.0:
  version "1.2.6"
`)

	for name, workspaces := range map[string]string{
		"Workspaces list":   `["packages/*"]`,
		"Workspaces object": `{"packages": ["packages/*"], "nohoist": ["**/react"]}`,
	} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"workspaces": `+workspaces+`}`), 0o600))
			lockFile := filepath.Join(dir, "yarn.lock")

			versions, err := parseYarnLock(lockFile, lock, "packages/app")
			require.NoError(t, err)
			assertLockedVersions(t, versions, []lockedVersionTest{
				{name: "axios", constraint: "^1.0.0", expected: "1.2.6"},
			})

			_, err = parseYarnLock(lockFile, lock, "tools/foo")
			assert.ErrorIs(t, err, errUnknownImporter)
		})
	}

	t.Run("Project without workspaces", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name": "project"}`), 0o600))

		_, err := parseYarnLock(filepath.Join(dir, "yarn.lock"), lock, "tools/foo")
		assert.ErrorIs(t, err, errUnknownImporter)
	})
}

func TestNormalizeYarnDescriptor(t *testing.T) {
	for descriptor, expected := range map[string]string{
		"":                     "",
		"__metadata":           "__metadata",
		"axios@^1.0.0":         "axios@^1.0.0",
		"axios@npm:^1.0.0":     "axios@^1.0.0",
		"@mdi/font@5.9.55":     "@mdi/font@5.9.55",
		"@mdi/font@npm:5.9.55": "@mdi/font@5.9.55",
	} {
		assert.Equal(t, expected, normalizeYarnDescriptor(descriptor), descriptor)
	}
}

func TestMatchWorkspacePattern(t *testing.T) {
	for _, tt := range []struct {
		pattern  string
		importer string
		expected bool
	}{
		{pattern: "packages/*", importer: "packages/app", expected: true},
		{pattern: "packages/*", importer: "packages/group/app", expected: false},
		{pattern: "packages/**", importer: "packages/app", expected: true},
		{pattern: "packages/**", importer: "packages/group/app", expected: true},
		{pattern: "packages/**/app", importer: "packages/app", expected: true},
		{pattern: "packages/**/app", importer: "packages/group/app", expected: true},
		{pattern: "packages/**/app", importer: "packages/group/lib", expected: false},
		{pattern: "./apps/web", importer: "apps/web", expected: true},
		{pattern: "apps/web", importer: "tools/foo", expected: false},
	} {
		assert.Equal(t, tt.expected, isWorkspace(filepath.Dir(writeRootPackageJson(t, `{"workspaces": ["`+tt.pattern+`"]}`)), tt.importer),
			"%s matching %s", tt.pattern, tt.importer)
	}
}

func TestLoadLockedVersions(t *testing.T) {
	t.Run("lock file next to the package.json", func(t *testing.T) {
		versions, err := loadLockedVersions("testdata/npmlockfile", "testdata/npmlockfile")
		require.NoError(t, err)

		assert.Equal(t, filepath.Join("testdata", "npmlockfile", "package-lock.json"), versions.lockFile)
		assertLockedVersions(t, versions, []lockedVersionTest{
			{name: "axios", constraint: "^1.0.0", expected: "1.2.6"},
		})
	})

	t.Run("workspace lock file at the root", func(t *testing.T) {
		versions, err := loadLockedVersions("testdata/npmworkspace/packages/app", "testdata/npmworkspace")
		require.NoError(t, err)

		assert.Equal(t, filepath.Join("testdata", "npmworkspace", "package-lock.json"), versions.lockFile)
		assertLockedVersions(t, versions, []lockedVersionTest{
			{name: "axios", constraint: "^1.0.0", expected: "1.2.6"},
		})
	})

	t.Run("lock file above the searched directory", func(t *testing.T) {
		versions, err := loadLockedVersions("testdata/npmworkspace/packages/app", "testdata/npmworkspace/packages")
		require.NoError(t, err)

		assert.Empty(t, versions.lockFile)
		assertLockedVersions(t, versions, []lockedVersionTest{
			{name: "axios", constraint: "^1.0.0", expected: ""},
		})
	})

	t.Run("lock file of an unrelated parent project", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "package-lock.json"), []byte(`{
  "lockfileVersion": 3,
  "packages": {
    "": {"name": "project"},
    "node_modules/axios": {"version": "1.2.6"}
  }
}`), 0o600))
		project := filepath.Join(dir, "tools", "foo")
		require.NoError(t, os.MkdirAll(project, 0o755))

		versions, err := loadLockedVersions(project, dir)
		require.NoError(t, err)

		assert.Empty(t, versions.lockFile)
		assertLockedVersions(t, versions, []lockedVersionTest{
			{name: "axios", constraint: "^1.0.0", expected: ""},
		})
	})

	t.Run("missing lock file", func(t *testing.T) {
		dir := t.TempDir()

		versions, err := loadLockedVersions(dir, dir)
		require.NoError(t, err)

		assertLockedVersions(t, versions, []lockedVersionTest{
			{name: "axios", constraint: "^1.0.0", expected: ""},
		})
	})

	t.Run("malformed lock file", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "package-lock.json"), []byte("{not json"), 0o600))

		_, err := loadLockedVersions(dir, dir)
		assert.ErrorContains(t, err, "parsing lock file")
	})

	t.Run("unreadable lock file", func(t *testing.T) {
		dir := t.TempDir()
		// A directory in place of the lock file fails to be read on every platform
		require.NoError(t, os.Mkdir(filepath.Join(dir, "pnpm-lock.yaml"), 0o755))

		_, err := loadLockedVersions(dir, dir)
		assert.ErrorContains(t, err, "reading lock file")
	})
}
