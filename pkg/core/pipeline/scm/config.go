package scm

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/goware/urlx"
	jschema "github.com/invopop/jsonschema"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/jsonschema"
	"github.com/updatecli/updatecli/pkg/plugins/scms/bitbucket"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git"
	"github.com/updatecli/updatecli/pkg/plugins/scms/gitea"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github"
	"github.com/updatecli/updatecli/pkg/plugins/scms/githubsearch"
	"github.com/updatecli/updatecli/pkg/plugins/scms/gitlab"
	"github.com/updatecli/updatecli/pkg/plugins/scms/gitlabsearch"
	"github.com/updatecli/updatecli/pkg/plugins/scms/stash"
	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
)

type ConfigSpec interface {
	Merge(interface{}) error
}

type Config struct {
	// Kind specifies the scm kind
	Kind string `yaml:",omitempty"`
	// Spec specifies the scm specification
	Spec interface{} `jsonschema:"type=object" yaml:",omitempty"`
	// Disabled is a setting used to disable the local git repository auto configuration
	Disabled bool
}

// Validate returns nil if the Config object is valid
// or a wrapped error of all the failed validations otherwise
func (c Config) Validate() error {
	validationErrs := []string{}

	if c.Disabled {
		if c.Kind != "" {
			validationErrs = append(validationErrs, "specified value for 'kind' found while SCM is disabled")
		}
		if c.Spec != nil {
			validationErrs = append(validationErrs, "specified value for 'spec' found while SCM is disabled")
		}
	} else {
		if c.Kind == "" {
			validationErrs = append(validationErrs, "missing value for parameter 'kind'")
		}
		// Ensure kind is lowercase
		if c.Kind != strings.ToLower(c.Kind) {
			logrus.Warningf("The specified value for the parameter 'kind' (%q) should be lowercase", c.Kind)
		}
		if c.Spec == nil {
			validationErrs = append(validationErrs, "missing value for parameter 'value'")
		}
	}

	if len(validationErrs) > 0 {
		return fmt.Errorf("%w: %s", ErrWrongConfig, strings.Join(validationErrs, ","))
	}
	return nil
}

// AutoGuess tries to fill the Config object receiver with the configuration auto-guessed from the provided directory
// It only returns an error if it fails "hard" (can't guess SCM, etc.).
func (c *Config) AutoGuess(configName, workingDir string, gitHandler gitgeneric.GitHandler) error {
	remotes, err := gitHandler.RemoteURLs(workingDir)
	if err != nil {
		return err
	}
	logrus.Debugf("Found the following remotes in %s: %s", workingDir, remotes)

	originRemoteURL, ok := remotes["origin"]
	if !ok {
		return fmt.Errorf("no remote named 'origin' could be found in the repository %s", workingDir)
	}

	var remoteHostname string
	httpRemoteUrl, err := url.Parse(originRemoteURL)
	if err != nil {
		// Case of non HTTPS scheme: either git or SSH (github.com for instance).
		// Hostname extraction on the "left" of the colon character (':')
		leftUrl := strings.Split(originRemoteURL, ":")[0]
		ux, _ := urlx.Parse(leftUrl)
		remoteHostname, _, _ = urlx.SplitHostPort(ux)

	} else {
		remoteHostname = httpRemoteUrl.Hostname()
	}

	switch remoteHostname {
	case "github.com":
		slashEverywhereUrl := strings.Split(strings.ReplaceAll(originRemoteURL, ":", "/"), "/")
		if len(slashEverywhereUrl) < 3 {
			return fmt.Errorf("unable to parse the following GitHub remote URL: %s", slashEverywhereUrl)
		}

		autoguessSpec := github.Spec{
			Repository: strings.TrimSuffix(slashEverywhereUrl[len(slashEverywhereUrl)-1], ".git"),
			Owner:      slashEverywhereUrl[len(slashEverywhereUrl)-2],
			Directory:  workingDir,
			Branch:     "main",
		}

		// Override configuration from environment variables
		autoguessSpec.MergeFromEnv(strings.ToUpper(fmt.Sprintf("UPDATECLI_SCM_%s", configName)))

		// If user specified settings, then try to merge with autoguessed specification
		currentGhSpec, ok := c.Spec.(github.Spec)
		if c.Spec != nil {
			if !ok {
				return fmt.Errorf("the SCM discovered in the directory %q has a different type ('github') than the specified SCM configuration %q", workingDir, configName)
			}
			if err := autoguessSpec.Merge(currentGhSpec); err != nil {
				return err
			}
		}

		c.Kind = "github"
		c.Spec = autoguessSpec
		return nil
	default:
		autoguessSpec := git.Spec{
			URL:       originRemoteURL,
			Directory: workingDir,
			Branch:    "main",
		}

		// Override configuration from environment variables
		autoguessSpec.MergeFromEnv(strings.ToUpper(fmt.Sprintf("UPDATECLI_SCM_%s", configName)))

		// If user specified settings, then try to merge with autoguessed specification
		currentGitSpec, ok := c.Spec.(git.Spec)

		if c.Spec != nil {
			if !ok {
				return fmt.Errorf("the SCM discovered in the directory %q has a different type ('git') than the specified SCM configuration %q", workingDir, configName)
			}
			if err := autoguessSpec.Merge(currentGitSpec); err != nil {
				return err
			}
		}

		c.Kind = "git"
		c.Spec = autoguessSpec
		return nil
	}
}

// JSONSchema implements the json schema interface to generate the "scm" jsonschema
func (Config) JSONSchema() *jschema.Schema {

	type configAlias Config

	anyOfSpec := map[string]interface{}{
		"bitbucket":    &bitbucket.Spec{},
		"git":          &git.Spec{},
		"gitea":        &gitea.Spec{},
		"github":       &github.Spec{},
		"gitlab":       &gitlab.Spec{},
		"stash":        &stash.Spec{},
		"githubsearch": &githubsearch.Spec{},
		"gitlabsearch": &gitlabsearch.Spec{},
	}

	return jsonschema.AppendOneOfToJsonSchema(configAlias{}, anyOfSpec)
}
