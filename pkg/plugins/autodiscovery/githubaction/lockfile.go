package githubaction

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	yaml "github.com/goccy/go-yaml"
	"github.com/updatecli/updatecli/pkg/core/log"
)

const (
	// gitHubHostname is the hostname of GitHub
	gitHubHostname = "github.com"
	// actionsLockFile is the path of the GitHub Actions workflow lockfile, relative to the repository root
	actionsLockFile = ".github/workflows/actions.lock"
)

var (
	// lockEnvironmentNames lists the environment variables the "gh actions-lock" command may use
	lockEnvironmentNames = []string{
		"PATH", "HOME", "USER", "TMPDIR", "LANG", "LC_ALL",
		"XDG_CONFIG_HOME", "XDG_DATA_HOME", "XDG_STATE_HOME", "XDG_CACHE_HOME",
		"GH_CONFIG_DIR", "GH_DATA_DIR", "GH_HOST",
		"HTTPS_PROXY", "https_proxy", "HTTP_PROXY", "http_proxy", "NO_PROXY", "no_proxy",
		"SSL_CERT_FILE", "SSL_CERT_DIR",
		"GH_TOKEN", "GH_ACTIONS_LOCK_GITHUB_COM_TOKEN",
	}

	// isGHActionsLockAvailable reports whether the "gh actions-lock" extension can run.
	// The lock command is a POSIX shell command, so it is never available on Windows, where
	// the shell resource runs PowerShell.
	// It is a variable so tests can replace it, and the check runs once per process.
	isGHActionsLockAvailable = sync.OnceValue(func() bool {
		if runtime.GOOS == "windows" {
			return false
		}
		return exec.Command("gh", "actions-lock", "--help").Run() == nil
	})

	// isGHAuthenticated reports whether gh has credentials for its host, GH_HOST or github.com.
	// It is a variable so tests can replace it, and the check runs once per process.
	isGHAuthenticated = sync.OnceValue(func() bool {
		host := os.Getenv("GH_HOST")
		if host == "" {
			host = gitHubHostname
		}
		return exec.Command("gh", "auth", "status", "--hostname", host).Run() == nil
	})
)

// actionsLock holds the part of the GitHub Actions workflow lockfile the crawler reads
type actionsLock struct {
	// Workflows lists the onboarded workflows, keyed by their path relative to the repository root
	Workflows map[string][]string `yaml:"workflows"`
}

// lockSetup reports whether "gh actions-lock" can update the lockfile, and returns the
// environment variables it runs with.
func (g *GitHubAction) lockSetup() (bool, []lockEnvironment) {
	if !isGHActionsLockAvailable() {
		return false, nil
	}

	token := g.lockToken()
	if token == "" && !isGHAuthenticated() {
		log.Warningf("\"gh actions-lock\" has no GitHub credentials: Updatecli resolved no GitHub token and gh isn't logged in")
		return false, nil
	}

	return true, lockEnvironments(token)
}

// isLockableActionHost reports whether "gh actions-lock" can lock the actions of host.
// It supports github.com and GitHub Enterprise Cloud with data residency, not GitHub Enterprise Server.
func isLockableActionHost(host string) bool {
	return host == gitHubHostname || strings.HasSuffix(host, ".ghe.com")
}

// lockedWorkflows returns the workflows onboarded in the GitHub Actions lockfile of rootDir,
// keyed by their slash separated path relative to rootDir.
// It returns a nil map when rootDir has no lockfile.
func lockedWorkflows(rootDir string) (map[string]bool, error) {
	data, err := os.ReadFile(filepath.Join(rootDir, actionsLockFile))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", actionsLockFile, err)
	}

	var lock actionsLock
	if err := yaml.Unmarshal(data, &lock); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", actionsLockFile, err)
	}

	workflows := make(map[string]bool, len(lock.Workflows))
	for workflow := range lock.Workflows {
		workflows[filepath.ToSlash(filepath.Clean(workflow))] = true
	}

	return workflows, nil
}

// lockToken returns the github.com token Updatecli resolves, for the "gh actions-lock" command.
// It returns an empty token, so gh uses its own login, when GH_HOST names another host, since
// a github.com token would then override the gh login of that host.
func (g *GitHubAction) lockToken() string {
	if host := os.Getenv("GH_HOST"); host != "" && host != gitHubHostname {
		log.Debugf("GH_HOST is %q, \"gh actions-lock\" uses the gh login", host)
		return ""
	}

	_, token, err := g.getGitProviderKind(defaultGitProviderURL)
	if err != nil {
		log.Warningf("resolving the GitHub token of \"gh actions-lock\", falling back to the gh login: %s", err)
		return ""
	}

	return token
}

// lockCommand returns the shell command updating the lockfile entries of workflow.
// It does nothing during "updatecli diff", where the workflow file isn't updated, so the
// lockfile in the working tree is never rewritten by a dry run.
func lockCommand(workflow string) string {
	return fmt.Sprintf(`if [ "$DRY_RUN" != "true" ]; then gh actions-lock %s; fi`, shellQuote(workflow))
}

// shellQuote quotes s as a single POSIX shell word
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// yamlQuote escapes s for a single quoted YAML scalar
func yamlQuote(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

// lockEnvironment is an environment variable passed to the "gh actions-lock" command
type lockEnvironment struct {
	Name string
	// Value is the value written in the manifest, unless Inherit is set
	Value string
	// Inherit passes the value of the Updatecli process environment
	Inherit bool
}

// lockEnvironments returns the environment variables of the "gh actions-lock" command.
//
// When Updatecli resolved a GitHub token, gh uses it through GH_TOKEN, which takes
// precedence over the gh login, so the lockfile is updated with the same credentials
// as the rest of the manifest. Otherwise gh falls back to the variables and the login
// of the Updatecli process.
//
// Listing environments replaces the default environment of the shell resource, so the
// variables gh needs to reach the API, such as the proxy and CA settings, are listed too.
// Only the variables set in the Updatecli process are inherited, since the shell resource
// would pass a missing one with an empty value, such as an empty GH_CONFIG_DIR.
func lockEnvironments(token string) []lockEnvironment {
	var environments []lockEnvironment
	for _, name := range lockEnvironmentNames {
		isTokenVariable := name == "GH_TOKEN" || name == "GH_ACTIONS_LOCK_GITHUB_COM_TOKEN"
		if isTokenVariable && token != "" {
			environments = append(environments, lockEnvironment{Name: name, Value: token})
			continue
		}
		if _, found := os.LookupEnv(name); found {
			environments = append(environments, lockEnvironment{Name: name, Inherit: true})
		}
	}

	return environments
}
