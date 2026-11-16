package githubaction

import (
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// unsetEnv removes the environment variable name for the duration of the test
func unsetEnv(t *testing.T, name string) {
	t.Helper()
	t.Setenv(name, "")
	require.NoError(t, os.Unsetenv(name))
}

// setLockEnv pins the environment variables lockEnvironments inherits to PATH and HOME
func setLockEnv(t *testing.T) {
	t.Helper()
	for _, name := range lockEnvironmentNames {
		unsetEnv(t, name)
	}
	t.Setenv("PATH", "/usr/bin")
	t.Setenv("HOME", "/home/updatecli")
}

// stubGHActionsLock replaces the "gh actions-lock" detection for the duration of the test
func stubGHActionsLock(t *testing.T, available bool) {
	t.Helper()
	previous := isGHActionsLockAvailable
	isGHActionsLockAvailable = func() bool { return available }
	t.Cleanup(func() { isGHActionsLockAvailable = previous })
}

// stubGHAuthenticated replaces the gh login detection for the duration of the test
func stubGHAuthenticated(t *testing.T, authenticated bool) {
	t.Helper()
	previous := isGHAuthenticated
	isGHAuthenticated = func() bool { return authenticated }
	t.Cleanup(func() { isGHAuthenticated = previous })
}

func TestIsLockableActionHost(t *testing.T) {
	assert.True(t, isLockableActionHost("github.com"))
	assert.True(t, isLockableActionHost("tenant.ghe.com"))
	assert.False(t, isLockableActionHost("ghes.example.com"))
	assert.False(t, isLockableActionHost("gitea.com"))
}

func TestLockSetup(t *testing.T) {
	unsetGitHubTokenEnv(t)
	setLockEnv(t)

	g, err := New(Spec{}, "testdata/lockfile", "", "")
	require.NoError(t, err)

	stubGHActionsLock(t, false)
	available, _ := g.lockSetup()
	assert.False(t, available, "without the extension")

	stubGHActionsLock(t, true)
	stubGHAuthenticated(t, false)
	available, _ = g.lockSetup()
	assert.False(t, available, "without Updatecli token nor gh login")

	stubGHAuthenticated(t, true)
	available, environments := g.lockSetup()
	assert.True(t, available, "with the gh login")
	assert.Equal(t, []lockEnvironment{
		{Name: "PATH", Inherit: true},
		{Name: "HOME", Inherit: true},
	}, environments)

	stubGHAuthenticated(t, false)
	t.Setenv("UPDATECLI_GITHUB_TOKEN", "updatecli-token")
	available, _ = g.lockSetup()
	assert.True(t, available, "the Updatecli token doesn't need a gh login")
}

func TestLockedWorkflows(t *testing.T) {
	got, err := lockedWorkflows("testdata/lockfile")
	require.NoError(t, err)
	assert.Equal(t, map[string]bool{
		".github/workflows/ci.yaml":     true,
		".github/workflows/pinned.yaml": true,
	}, got)

	got, err = lockedWorkflows("testdata/updatecli")
	require.NoError(t, err)
	assert.Nil(t, got)

	_, err = lockedWorkflows("testdata/lockfile_invalid")
	assert.Error(t, err)
}

func TestLockEnvironments(t *testing.T) {
	setLockEnv(t)

	assert.Equal(t, []lockEnvironment{
		{Name: "PATH", Inherit: true},
		{Name: "HOME", Inherit: true},
		{Name: "GH_TOKEN", Value: "updatecli"},
		{Name: "GH_ACTIONS_LOCK_GITHUB_COM_TOKEN", Value: "updatecli"},
	}, lockEnvironments("updatecli"), "the Updatecli token is passed to gh")

	t.Setenv("GH_TOKEN", "gh")
	t.Setenv("HTTPS_PROXY", "http://proxy:3128")
	t.Setenv("SSL_CERT_FILE", "/etc/ssl/ca.pem")
	assert.Equal(t, []lockEnvironment{
		{Name: "PATH", Inherit: true},
		{Name: "HOME", Inherit: true},
		{Name: "HTTPS_PROXY", Inherit: true},
		{Name: "SSL_CERT_FILE", Inherit: true},
		{Name: "GH_TOKEN", Value: "updatecli"},
		{Name: "GH_ACTIONS_LOCK_GITHUB_COM_TOKEN", Value: "updatecli"},
	}, lockEnvironments("updatecli"), "the Updatecli token overrides GH_TOKEN, proxy and CA settings are inherited")

	assert.Equal(t, []lockEnvironment{
		{Name: "PATH", Inherit: true},
		{Name: "HOME", Inherit: true},
		{Name: "HTTPS_PROXY", Inherit: true},
		{Name: "SSL_CERT_FILE", Inherit: true},
		{Name: "GH_TOKEN", Inherit: true},
	}, lockEnvironments(""), "without Updatecli token, only the variables set are inherited")
}

func TestLockCommand(t *testing.T) {
	sh, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("sh not available")
	}

	dir := t.TempDir()
	// A fake gh records its arguments, one per line
	require.NoError(t, os.WriteFile(dir+"/gh", []byte("#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$GH_ARGS\"\n"), 0o700)) //nolint:gosec // test helper script

	run := func(t *testing.T, dryRun string) (string, bool) {
		args := dir + "/args"
		_ = os.Remove(args)
		cmd := exec.Command(sh, "-c", lockCommand(".github/workflows/it's; touch pwned.yaml")) //nolint:gosec // the command under test
		cmd.Dir = dir
		cmd.Env = []string{"PATH=" + dir + ":/usr/bin:/bin", "DRY_RUN=" + dryRun, "GH_ARGS=" + args}
		require.NoError(t, cmd.Run())
		data, err := os.ReadFile(args)
		return string(data), err == nil
	}

	got, ran := run(t, "false")
	require.True(t, ran)
	assert.Equal(t, "actions-lock\n.github/workflows/it's; touch pwned.yaml\n", got, "the workflow path is a single argument")
	assert.NoFileExists(t, dir+"/pwned.yaml")

	_, ran = run(t, "true")
	assert.False(t, ran, "gh actions-lock doesn't run during a dry run")
}

func TestDiscoverManifestsLockfile(t *testing.T) {
	unsetGitHubTokenEnv(t)
	setLockEnv(t)
	t.Setenv("UPDATECLI_GITHUB_TOKEN", "updatecli-token")
	t.Setenv("GITHUB_TOKEN", "github-token")

	disabled := false
	newCrawler := func(t *testing.T, rootDir, scmID string, digest *bool) GitHubAction {
		g, err := New(Spec{Digest: digest}, rootDir, scmID, "")
		require.NoError(t, err)
		return g
	}

	expectedLockTarget := `
  lock:
    dependson:
      - 'target#release:or'
      - 'target#tag:or'
      - 'target#branch:or'
    dependsonchange: true
    disableconditions: true
    disablesourceinput: true
    name: 'deps(github): update .github/workflows/actions.lock for .github/workflows/ci.yaml'
    kind: 'shell'
    spec:
      command: 'if [ "$DRY_RUN" != "true" ]; then gh actions-lock ''.github/workflows/ci.yaml''; fi'
      environments:
        - name: 'PATH'
        - name: 'HOME'
        - name: 'GH_TOKEN'
          value: 'updatecli-token'
        - name: 'GH_ACTIONS_LOCK_GITHUB_COM_TOKEN'
          value: 'updatecli-token'
      changedif:
        kind: 'file/checksum'
        spec:
          files:
            - '.github/workflows/actions.lock'
`

	t.Run("gh actions-lock available", func(t *testing.T) {
		stubGHActionsLock(t, true)

		manifests, err := newCrawler(t, "testdata/lockfile", "", &disabled).DiscoverManifests()
		require.NoError(t, err)
		require.Len(t, manifests, 3)

		assert.Contains(t, string(manifests[0]), expectedLockTarget,
			"the lock target uses UPDATECLI_GITHUB_TOKEN rather than GITHUB_TOKEN")
		assert.NotContains(t, string(manifests[1]), "gh actions-lock", "other.yaml isn't onboarded in the lockfile")
		for _, manifest := range manifests {
			assert.NotContains(t, string(manifest), "gitea.com", "a Gitea action of a locked workflow can't update the lockfile")
			assert.NotContains(t, string(manifest), "ghes.example.com", "gh actions-lock doesn't support GitHub Enterprise Server")
		}
	})

	t.Run("gh actions-lock available with scm", func(t *testing.T) {
		stubGHActionsLock(t, true)

		manifests, err := newCrawler(t, "testdata/lockfile", "default", &disabled).DiscoverManifests()
		require.NoError(t, err)
		require.Len(t, manifests, 3)

		assert.Contains(t, string(manifests[0]), `
    kind: 'shell'
    scmid: 'default'
    spec:
      command: 'if [ "$DRY_RUN" != "true" ]; then gh actions-lock ''.github/workflows/ci.yaml''; fi'`)
	})

	t.Run("default digest", func(t *testing.T) {
		stubGHActionsLock(t, true)

		manifests, err := newCrawler(t, "testdata/lockfile", "", nil).DiscoverManifests()
		require.NoError(t, err)
		require.Len(t, manifests, 3)

		assert.NotContains(t, string(manifests[0]), "release_digest", "a locked workflow isn't pinned by default")
		assert.Contains(t, string(manifests[1]), "release_digest", "a workflow without lock keeps the default digest")
		assert.Contains(t, string(manifests[2]), "release_digest", "a locked step already pinned stays pinned")
	})

	t.Run("digest enabled explicitly", func(t *testing.T) {
		stubGHActionsLock(t, true)

		enabled := true
		manifests, err := newCrawler(t, "testdata/lockfile", "", &enabled).DiscoverManifests()
		require.NoError(t, err)
		require.Len(t, manifests, 3)

		assert.Contains(t, string(manifests[0]), "release_digest")
	})

	t.Run("gh actions-lock unavailable", func(t *testing.T) {
		stubGHActionsLock(t, false)

		manifests, err := newCrawler(t, "testdata/lockfile", "", &disabled).DiscoverManifests()
		require.NoError(t, err)
		require.Len(t, manifests, 1, "the actions of the locked workflows are skipped")
		assert.Contains(t, string(manifests[0]), "actions/setup-go")
	})

	t.Run("invalid lockfile", func(t *testing.T) {
		stubGHActionsLock(t, true)

		manifests, err := newCrawler(t, "testdata/lockfile_invalid", "", &disabled).DiscoverManifests()
		require.NoError(t, err, "an invalid lockfile doesn't stop the crawler")
		assert.Empty(t, manifests, "the actions of the workflows next to an invalid lockfile are skipped")
	})
}

func TestLockToken(t *testing.T) {
	unsetGitHubTokenEnv(t)
	setLockEnv(t)
	t.Setenv("UPDATECLI_GITHUB_TOKEN", "updatecli-token")

	g, err := New(Spec{}, "testdata/lockfile", "", "")
	require.NoError(t, err)

	assert.Equal(t, "updatecli-token", g.lockToken())

	t.Setenv("GH_HOST", "github.com")
	assert.Equal(t, "updatecli-token", g.lockToken())

	t.Setenv("GH_HOST", "tenant.ghe.com")
	assert.Empty(t, g.lockToken(), "a github.com token would override the gh login of another host")
}

func TestRunGHProbe(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh not available")
	}

	dir := t.TempDir()
	// A fake gh succeeds, fails or hangs depending on its first argument
	script := "#!/bin/sh\ncase \"$1\" in\n  ok) exit 0 ;;\n  hang) sleep 10 ;;\n  *) exit 1 ;;\nesac\n"
	require.NoError(t, os.WriteFile(dir+"/gh", []byte(script), 0o700)) //nolint:gosec // test helper script
	t.Setenv("PATH", dir+":"+os.Getenv("PATH"))

	previous := ghProbeTimeout
	ghProbeTimeout = 200 * time.Millisecond
	t.Cleanup(func() { ghProbeTimeout = previous })

	assert.True(t, runGHProbe("ok"))
	assert.False(t, runGHProbe("fail"))

	start := time.Now()
	assert.False(t, runGHProbe("hang"), "a probe still running at the deadline fails")
	assert.Less(t, time.Since(start), 5*time.Second, "the probe is killed at the deadline")
}
