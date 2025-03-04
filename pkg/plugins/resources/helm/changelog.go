package helm

import (
	"bytes"
	"fmt"
	"html/template"
	"maps"
	"slices"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	githubChangelog "github.com/updatecli/updatecli/pkg/plugins/changelog/github/v3"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/repo"
)

const (
	artifactHubChangesAnnotation = "artifacthub.io/changes"
	artifactHubLinksAnnotation   = "artifacthub.io/links"

	// CHANGELOGTEMPLATE contains helm chart changelog fallback template
	CHANGELOGTEMPLATE string = `
Remark: We couldn't identify a way to automatically retrieve changelog information.
Please use following information to take informed decision

{{ if .Name }}Helm Chart: {{ .Name }}{{ end }}
{{ if .Description }}{{ .Description }}{{ end }}
{{ if .Home }}Project Home: {{ .Home }}{{ end }}
{{ if .KubeVersion }}Require Kubernetes Version: {{ .KubeVersion }}{{end}}
{{ if .Created }}Version created on the {{ .Created }}{{ end}}
{{ if .Sources }}
Sources:
{{ range $index, $source := .Sources }}
* {{ $source }}
{{ end }}
{{ end }}
{{ if .URLs }}
URL:
{{ range $index, $url := .URLs }}
* {{ $url }}
{{ end }}
{{ end }}
`
)

// Changelog returns a rendered template with this chart version information
func (c Chart) Changelog(from, to string) *result.Changelogs {
	index, err := c.GetRepoIndexFromURL()

	if err != nil {
		logrus.Debugf("failed to get helm repository index: %s", err)
		return nil
	}

	e, err := index.Get(c.spec.Name, c.foundVersion.OriginalVersion)
	if err != nil {
		logrus.Debugf("failed to get helm chart information: %s", err)
		return nil
	}

	index.SortEntries()

	changelogs := c.getChangelogsFromArtifactHubAnnotations(index, from, to)

	if len(changelogs) > 0 {
		return &changelogs
	} else {
		logrus.Debugf("no changelog found in %s annotation for versions between %s and %s", artifactHubChangesAnnotation, from, to)

		changelogs = c.getChangelogsFromGithubReleases(index, from, to)
		if len(changelogs) > 0 {
			return &changelogs
		}

		t := template.Must(template.New("changelog").Parse(CHANGELOGTEMPLATE))
		buffer := new(bytes.Buffer)

		type params struct {
			Name        string
			Description string
			Home        string
			KubeVersion string
			Created     string
			URLs        []string `json:"url"`
			Sources     []string
		}

		err = t.Execute(buffer, params{
			Name:        e.Name,
			Description: e.Description,
			Home:        e.Home,
			KubeVersion: e.KubeVersion,
			Created:     e.Created.String(),
			URLs:        e.URLs,
			Sources:     e.Sources})

		if err != nil {
			logrus.Debugf("failed to render helm chart information: %s", err)
			return nil
		}

		changelog := buffer.String()
		return &result.Changelogs{
			{
				Title:       from,
				Body:        changelog,
				PublishedAt: e.Created.String(),
			},
		}
	}
}

func (c Chart) getChangelogsFromGithubReleases(index repo.IndexFile, from string, to string) result.Changelogs {

	chartVersion, err := index.Get(c.spec.Name, from)
	if err != nil {
		logrus.Debugf("failed to get helm chart information: %s", err)
		return nil
	}

	linksAnnotation, ok := chartVersion.Annotations[artifactHubLinksAnnotation]
	if !ok {
		logrus.Debugf("no %s annotation found in %s", artifactHubLinksAnnotation, from)
		return nil
	}
	type link struct {
		URL  string `yaml:"url"`
		Name string `yaml:"name"`
	}
	var links []link
	err = yaml.Unmarshal([]byte(linksAnnotation), &links)
	if err != nil {
		logrus.Debugf("failed to unmarshal %s annotation: %s", artifactHubLinksAnnotation, err)
		return nil
	}

	linkIndex := slices.IndexFunc(links, func(l link) bool {
		return l.Name == "Chart Source"
	})
	if linkIndex == -1 {
		logrus.Debugf("no chart source link found in %s annotation", artifactHubLinksAnnotation)
		return nil
	}

	chartSourceLink := links[linkIndex].URL
	logrus.Debugf("chart source link: %s", chartSourceLink)
	if !strings.HasPrefix(chartSourceLink, "https://github.com/") {
		logrus.Debugf("chart source link is not a github link: %s", chartSourceLink)
		return nil
	}

	changelog := c.getChangelogFromGitHub(chartSourceLink, from, to)
	return changelog
}

func (c Chart) getChangelogFromGitHub(chartSourceLink, from, to string) result.Changelogs {
	parsedRepo := strings.Split(strings.TrimPrefix(chartSourceLink, "https://github.com/"), "/")
	if len(parsedRepo) < 2 {
		logrus.Debugf("invalid chart source link: %s", chartSourceLink)
		return nil
	}

	changelog := githubChangelog.Changelog{
		Owner:      parsedRepo[0],
		Repository: parsedRepo[1],
	}

	fromLongVersion := c.getLongVersion(from)
	toLongVersion := c.getLongVersion(to)
	releases, err := changelog.Search(fromLongVersion, toLongVersion)
	if err != nil {
		logrus.Debugf("failed to search github releases: %s", err)
	}

	// exclude fromVersion from releases
	releases = slices.DeleteFunc(releases, func(c result.Changelog) bool {
		return c.Title == fromLongVersion
	})

	return releases
}

func (c Chart) getLongVersion(version string) string {
	return fmt.Sprintf("%s-%s", c.spec.Name, version)
}

func (c Chart) getChangelogsFromArtifactHubAnnotations(index repo.IndexFile, from string, to string) result.Changelogs {
	var changelogs result.Changelogs
	for _, version := range index.Entries[c.spec.Name] {
		versionObj, err := semver.NewVersion(version.Version)
		if err != nil {
			continue
		}

		fromVersion, err := semver.NewVersion(from)
		if err != nil {
			continue
		}

		var toVersion *semver.Version
		if to != "" {
			toVersion, err = semver.NewVersion(to)
			if err != nil {
				continue
			}
		}

		if versionObj.GreaterThan(fromVersion) && (toVersion == nil || versionObj.LessThanEqual(toVersion)) {
			changesAnnotation, ok := version.Annotations[artifactHubChangesAnnotation]
			if ok {
				changelogs = append(changelogs, result.Changelog{
					Title:       version.Version,
					Body:        parseChangeAnnotation(changesAnnotation),
					PublishedAt: version.Created.String(),
				})
			}
		}
	}

	return changelogs
}

type change struct {
	Kind        string `yaml:"kind,omitempty"`
	Description string `yaml:"description,omitempty"`
}

func parseChangeAnnotation(changesAnnotation string) string {

	var changes []change
	var yamlData = []byte(changesAnnotation)
	// Try to unmarshal as a list of changes first
	err := yaml.Unmarshal(yamlData, &changes)
	if err != nil {
		// If that fails, try to unmarshal as a list of strings
		var simpleList []string
		err = yaml.Unmarshal(yamlData, &simpleList)
		if err != nil {
			logrus.Debugf("failed to unmarshal %s annotation: %s", artifactHubChangesAnnotation, err)
			return changesAnnotation
		}
		return renderSimpleList(simpleList)
	}

	return renderChanges(changes)
}

func renderSimpleList(list []string) string {
	var renderedList string
	for _, item := range list {
		renderedList += fmt.Sprintf("* %s\n", item)
	}
	return renderedList
}

func renderChanges(changes []change) string {
	var renderedChanges string
	changesByKind := make(map[string][]change)
	for _, ch := range changes {

		if _, ok := changesByKind[ch.Kind]; !ok {
			changesByKind[ch.Kind] = make([]change, 0)
		}
		changesByKind[ch.Kind] = append(changesByKind[ch.Kind], ch)
	}

	sortedKinds := slices.Sorted(maps.Keys(changesByKind))

	// Render changes in alphabetical order by kind
	for _, kind := range sortedKinds {
		renderedChanges += fmt.Sprintf("## %s\n\n", capitalize(kind))
		for _, ch := range changesByKind[kind] {
			renderedChanges += fmt.Sprintf("* %s\n", ch.Description)
		}
		renderedChanges += "\n"
	}
	return renderedChanges
}

func capitalize(s string) string {
	return strings.ToUpper(s[:1]) + s[1:]
}
