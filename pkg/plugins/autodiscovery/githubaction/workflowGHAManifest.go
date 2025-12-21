package githubaction

import (
	"bytes"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"
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
			logrus.Errorf("building URL: %s", err)
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
			logrus.Debugf("Ignoring GitHub Action %q as matching ignore rule(s)\n", actionName)
			return nil, nil
		}
	}

	if len(g.spec.Only) > 0 {
		if !g.spec.Only.isMatchingRules(g.rootDir, spec.RelativeFoundFile, actionName, spec.Reference) {
			logrus.Debugf("Ignoring GitHub Action %q as not matching only rule(s)\n", actionName)
			return nil, nil
		}
	}

	if !g.digest && slices.Contains([]string{"latest", "master", "main"}, spec.Reference) {
		logrus.Debugf("Ignoring GitHub Action %q as it uses the reference %q \n",
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
			logrus.Debugf("building version filter pattern: %s", err)
			versionFilterPattern = versionFilterRef
		}
	}

	var tmpl *template.Template
	switch kind {
	case kindGitHub:
		tmpl, err = template.New("manifest").Parse(workflowManifestGitHubTemplate)
		if err != nil {
			return nil, fmt.Errorf("parsing template: %s", err)
		}
	case kindGitea:
		tmpl, err = template.New("manifest").Parse(workflowManifestGiteaTemplate)
		if err != nil {
			return nil, fmt.Errorf("parsing template: %s", err)
		}
	default:
		return nil, fmt.Errorf("unsupported git provider kind %q, skipping", kind)
	}

	var targetKey string
	if spec.Composite {
		targetKey = fmt.Sprintf(`$.runs.steps[%d].uses`, spec.StepID)
	} else {
		targetKey = fmt.Sprintf(`$.jobs.%s.steps[%d].uses`, spec.JobID, spec.StepID)
	}

	params := struct {
		ActionID             string
		ActionName           string
		Reference            string
		PinReference         string
		File                 string
		ImageName            string
		URL                  string
		Owner                string
		Repository           string
		VersionFilterKind    string
		VersionFilterPattern string
		VersionFilterRegex   string
		TargetKey            string
		ScmID                string
		Token                string
		Digest               bool
	}{
		ActionID:             g.actionID,
		ActionName:           actionName,
		Reference:            spec.Reference,
		PinReference:         pinReference,
		File:                 spec.RelativeFoundFile,
		URL:                  spec.URL,
		Owner:                spec.Owner,
		Repository:           spec.Repository,
		VersionFilterKind:    versionFilterKind,
		VersionFilterPattern: versionFilterPattern,
		VersionFilterRegex:   versionFilterRegex,
		ScmID:                g.scmID,
		TargetKey:            targetKey,
		Token:                token,
		Digest:               g.digest,
	}

	manifest := bytes.Buffer{}
	if err := tmpl.Execute(&manifest, params); err != nil {
		return nil, err
	}

	return manifest.Bytes(), nil
}
