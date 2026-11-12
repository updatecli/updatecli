package githubaction

import (
	"bytes"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"text/template"

	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
)

type githubActionManifestSpec struct {
	URL               string
	Owner             string
	Repository        string
	Directory         string
	Reference         string
	RelativeFoundFile string
	CommentDigest     string
	JobID             string
	Composite         bool
	StepID            int
}

func (g GitHubAction) getGitHubActionManifest(spec *githubActionManifestSpec) ([]byte, error) {

	var err error

	pinReference := parseActionDigestComment(spec.CommentDigest)
	if g.digest {
		if pinReference == "" {
			// First time we pin a ref
			pinReference = spec.Reference
		}
	}
	switch spec.Reference {
	case "":
		return nil, nil
	case "main", "master":
		if !g.digest {
			return nil, nil
		}
	}

	actionName := ""
	if spec.URL == "" {
		actionName = spec.Owner + "/" + spec.Repository
		spec.URL = defaultGitProviderURL
	} else {
		actionName, err = url.JoinPath(spec.URL, spec.Owner, spec.Repository)
		if err != nil {
			log.Errorf("building URL: %s", err)
		}
	}

	kind, token, err := g.getGitProviderKind(spec.URL)
	if err != nil {
		return nil, fmt.Errorf("getting credentials: %s", err)
	}

	if spec.Directory != "" {
		actionName = strings.Join([]string{actionName, spec.Directory}, "/")
	}

	if len(g.spec.Ignore) > 0 {
		if g.spec.Ignore.isMatchingRules(g.rootDir, spec.RelativeFoundFile, actionName, spec.Reference) {
			log.Debugf("Ignoring GitHub Action %q as matching ignore rule(s)\n", actionName)
			return nil, nil
		}
	}

	if len(g.spec.Only) > 0 {
		if !g.spec.Only.isMatchingRules(g.rootDir, spec.RelativeFoundFile, actionName, spec.Reference) {
			log.Debugf("Ignoring GitHub Action %q as not matching only rule(s)\n", actionName)
			return nil, nil
		}
	}

	if !g.digest && slices.Contains([]string{latestVersionIdentifier, "master", "main"}, spec.Reference) {
		log.Debugf("Ignoring GitHub Action %q as it uses the reference %q \n",
			actionName,
			spec.Reference,
		)
		return nil, nil
	}

	versionFilterRef := spec.Reference
	if g.digest {
		versionFilterRef = pinReference
	}

	versionFilterKind, versionFilterPattern := detectVersionFilter(versionFilterRef)
	versionFilterRegex := "*"
	if !g.spec.VersionFilter.IsZero() {
		versionFilterKind = g.versionFilter.Kind
		versionFilterPattern, err = g.versionFilter.GreaterThanPattern(versionFilterRef)
		versionFilterRegex = g.versionFilter.Regex
		if err != nil {
			log.Debugf("building version filter pattern: %s", err)
			versionFilterPattern = versionFilterRef
		}
	}

	// A non semver reference such as "main" or "master" names one branch, so the branch
	// source must track that branch. The "latest" filter would return the last branch by
	// name and silently move the workflow to another branch.
	branchVersionFilterKind := versionFilterKind
	branchVersionFilterPattern := versionFilterPattern
	if g.spec.VersionFilter.IsZero() &&
		versionFilterKind == latestVersionIdentifier &&
		versionFilterRef != latestVersionIdentifier {
		branchVersionFilterKind = "regex"
		branchVersionFilterPattern = `^\Q` + versionFilterRef + `\E$`
	}

	var providerTemplate, gitUsername string
	switch kind {
	case kindGitHub:
		providerTemplate = workflowManifestGitHubTemplate
		// GitHub accepts this username with both personal access tokens and app tokens
		gitUsername = "x-access-token"
	case kindGitea, kindForgejo:
		providerTemplate = workflowManifestGiteaTemplate
		gitUsername = "oauth2"
	default:
		return nil, fmt.Errorf("unsupported git provider kind %q, skipping", kind)
	}

	// gittag and gitbranch skip authentication without a password, so drop the username too
	if token == "" {
		gitUsername = ""
	}

	tmpl, err := template.New("manifest").Parse(age.ManifestTemplate + workflowManifestActionPartials + providerTemplate)
	if err != nil {
		return nil, fmt.Errorf("parsing template: %s", err)
	}

	var targetKey string
	if spec.Composite {
		targetKey = fmt.Sprintf(`$.runs.steps[%d].uses`, spec.StepID)
	} else {
		targetKey = fmt.Sprintf(`$.jobs.%s.steps[%d].uses`, spec.JobID, spec.StepID)
	}

	params := actionManifestParams{
		Kind:                       kind,
		ActionID:                   g.actionID,
		ActionName:                 actionName,
		Reference:                  spec.Reference,
		CheckReference:             versionFilterRef,
		File:                       spec.RelativeFoundFile,
		URL:                        spec.URL,
		Owner:                      spec.Owner,
		Repository:                 spec.Repository,
		VersionFilterKind:          versionFilterKind,
		VersionFilterPattern:       versionFilterPattern,
		VersionFilterRegex:         versionFilterRegex,
		BranchVersionFilterKind:    branchVersionFilterKind,
		BranchVersionFilterPattern: branchVersionFilterPattern,
		ScmID:                      g.scmID,
		TargetKey:                  targetKey,
		GitUsername:                gitUsername,
		Token:                      token,
		Digest:                     g.digest,
		Age:                        g.spec.Age,
	}

	manifest := bytes.Buffer{}
	if err := tmpl.Execute(&manifest, params); err != nil {
		return nil, err
	}

	return manifest.Bytes(), nil
}

// actionManifestParams holds the values of the GitHub and Gitea action manifest templates
type actionManifestParams struct {
	// Kind is the git provider kind, such as "github" or "gitea"
	Kind       string
	ActionID   string
	ActionName string
	Reference  string
	// CheckReference is the reference the conditions check and the version filters start from
	CheckReference       string
	File                 string
	URL                  string
	Owner                string
	Repository           string
	VersionFilterKind    string
	VersionFilterPattern string
	VersionFilterRegex   string
	// BranchVersionFilterKind and BranchVersionFilterPattern only apply to the branch source
	BranchVersionFilterKind    string
	BranchVersionFilterPattern string
	TargetKey                  string
	ScmID                      string
	// GitUsername is the basic auth username of the gittag and gitbranch resources
	GitUsername string
	Token       string
	Digest      bool
	Age         age.Spec
}

// actionSourceParams narrows actionManifestParams to one source, "release", "tag", or "branch",
// and to the condition and target sharing its ID
type actionSourceParams struct {
	actionManifestParams
	SourceID      string
	FilterKind    string
	FilterPattern string
}

// Source returns the template values of the source sourceID
func (p actionManifestParams) Source(sourceID string) actionSourceParams {
	s := actionSourceParams{
		actionManifestParams: p,
		SourceID:             sourceID,
		FilterKind:           p.VersionFilterKind,
		FilterPattern:        p.VersionFilterPattern,
	}

	if sourceID == "branch" {
		s.FilterKind = p.BranchVersionFilterKind
		s.FilterPattern = p.BranchVersionFilterPattern
	}

	return s
}

// GitKind returns the git resource kind listing the source references
func (p actionSourceParams) GitKind() string {
	if p.SourceID == "branch" {
		return "gitbranch"
	}
	return "gittag"
}

// TargetSourceID returns the ID of the source whose value the target writes
func (p actionSourceParams) TargetSourceID() string {
	if p.Digest {
		return p.SourceID + "_digest"
	}
	return p.SourceID
}

// TargetTitle returns the start of the target name
func (p actionSourceParams) TargetTitle() string {
	if p.Kind == kindGitHub {
		return "deps(github): bump Action " + p.SourceID + " for " + p.ActionName
	}
	return "deps(gitea): bump " + p.ActionName
}
