package maven

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/updatecli/updatecli/pkg/core/log"

	"github.com/beevik/etree"
)

// propertyRegex matches a version relying on a property, such as "${log4j.version}".
var propertyRegex = regexp.MustCompile(`\$\{.*\}`)

// isPropertyVersion returns true for a version relying on a property, such as "${log4j.version}".
func isPropertyVersion(version string) bool {
	return propertyRegex.MatchString(version)
}

// propertyNameRegex captures the property name of a version made of a single property, such as "${log4j.version}".
var propertyNameRegex = regexp.MustCompile(`^\$\{([^${}]+)\}$`)

// propertyName returns the name of the property a version is made of, such as "log4j.version" for "${log4j.version}".
func propertyName(version string) (string, bool) {
	match := propertyNameRegex.FindStringSubmatch(version)
	if match == nil {
		return "", false
	}
	return match[1], true
}

// elemText returns the text of an XML element without the whitespace around it,
// which POM files often have when an element spans several lines.
func elemText(elem *etree.Element) string {
	return strings.TrimSpace(elem.Text())
}

// isVersionRange returns true for a Maven version range, such as "[1.0,2.0)" or "[1.5,)".
func isVersionRange(version string) bool {
	return strings.ContainsAny(version, "[](),")
}

// searchPomFiles will look, recursively, for every files named Chart.yaml from a root directory.
func searchPomFiles(rootDir string, files []string) ([]string, error) {

	pomFiles := []string{}

	log.Debugf("Looking for Maven pom.xml files in %q", rootDir)

	// To do switch to WalkDir which is more efficient, introduced in 1.16
	err := filepath.Walk(rootDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			log.Warningf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		for _, f := range files {
			if info.Name() == f {
				pomFiles = append(pomFiles, path)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return pomFiles, nil
}

// getRepositories retrieves all repositories information from a pom.xm
func getParentFromPom(doc *etree.Document) parentPom {

	p := parentPom{}

	parent := doc.FindElement("//project/parent")

	if parent == nil {
		return p
	}

	if elem := parent.FindElement("./groupId"); elem != nil {
		p.GroupID = elemText(elem)
	}

	if elem := parent.FindElement("./artifactId"); elem != nil {
		p.ArtifactID = elemText(elem)
	}

	if elem := parent.FindElement("./version"); elem != nil {
		p.Version = elemText(elem)
	}

	if elem := parent.FindElement("./packaging"); elem != nil {
		p.Packaging = elemText(elem)
	}

	if elem := parent.FindElement("./relativePath"); elem != nil {
		p.RelativePath = elemText(elem)
	}

	return p
}

// getRepositories retrieves all repositories information from a pom.xm
func getRepositoriesFromPom(doc *etree.Document) []repository {
	repositories := []repository{}

	repositoriesPath := "//project/repositories/*"

	for _, repositoryElem := range doc.FindElements(repositoriesPath) {

		if repositoryElem == nil {
			repositories = append(repositories, repository{})
			continue
		}

		rep := repository{}

		if elem := repositoryElem.FindElement("./url"); elem != nil {
			rep.URL = elemText(elem)
		}

		if elem := repositoryElem.FindElement("./id"); elem != nil {
			rep.ID = elemText(elem)
		}

		repositories = append(repositories, rep)

	}

	return repositories
}

// getDependenciesFromPom parse a pom.xml and return all dependencies
func getDependenciesFromPom(doc *etree.Document) []dependency {
	dependencies := []dependency{}

	dependenciesPath := "//project/dependencies/*"
	for _, dependencyElem := range doc.FindElements(dependenciesPath) {
		dep := dependency{}

		if dependencyElem == nil {
			dependencies = append(dependencies, dep)
			continue
		}

		if elem := dependencyElem.FindElement("./groupId"); elem != nil {
			dep.GroupID = elemText(elem)
		}

		if elem := dependencyElem.FindElement("./artifactId"); elem != nil {
			dep.ArtifactID = elemText(elem)
		}

		if elem := dependencyElem.FindElement("./version"); elem != nil {
			dep.Version = elemText(elem)
		}

		dependencies = append(dependencies, dep)

	}
	return dependencies
}

// getDependencyManagementsFromPom parse a pom.xml and return all dependencies
func getDependencyManagementsFromPom(doc *etree.Document) []dependency {
	dependencies := []dependency{}

	dependenciesPath := "//project/dependencyManagement/dependencies/*"
	for _, dependencyElem := range doc.FindElements(dependenciesPath) {

		dep := dependency{}

		if dependencyElem == nil {
			dependencies = append(dependencies, dep)
			continue
		}

		if elem := dependencyElem.FindElement("./groupId"); elem != nil {
			dep.GroupID = elemText(elem)
		}

		if elem := dependencyElem.FindElement("./artifactId"); elem != nil {
			dep.ArtifactID = elemText(elem)
		}

		if elem := dependencyElem.FindElement("./version"); elem != nil {
			dep.Version = elemText(elem)
		}

		if elem := dependencyElem.FindElement("./type"); elem != nil {
			dep.Type = elemText(elem)
		}

		if elem := dependencyElem.FindElement("./scope"); elem != nil {
			dep.Scope = elemText(elem)
		}

		dependencies = append(dependencies, dep)

	}
	return dependencies
}

// getPropertiesFromPom returns the properties defined in a pom.xml, by name.
func getPropertiesFromPom(doc *etree.Document) map[string]string {
	properties := map[string]string{}

	for _, elem := range doc.FindElements("//project/properties/*") {
		properties[elem.Tag] = elemText(elem)
	}

	return properties
}

// getMavenRepositoriesURL returns the URL for all repository while trying to identify potential credentials and mirrors
// It will look for repositories in the pom.xml and in the settings.xml
func getMavenRepositoriesURL(pomFile string, doc *etree.Document) (mavenRepositories []string) {

	repositories := getRepositoriesFromPom(doc)

	settingsXMLMap := map[string]Settings{}

	for _, path := range getSettingsXMLPath(filepath.Dir(pomFile)) {
		settings := readSettingsXML(path)
		if settings == nil {
			log.Debugf("No settings.xml content found in %q", path)
			continue
		}

		settingsXMLRepositories := settings.getRepositoriesFromSettingsXML()

		if len(settingsXMLRepositories) > 0 {
			log.Debugf("Found %d repositories in %q", len(settingsXMLRepositories), path)
			for _, repo := range settingsXMLRepositories {
				log.Debugf("\t* repository %q %q", repo.ID, repo.URL)
			}

			repositories = append(repositories, settingsXMLRepositories...)
		}

		settingsXMLMap[path] = *settings
	}

	setCredentials := func(settings Settings, url, id string) (string, bool) {

		username, password := settings.getMatchingServerCredentials(id)

		cred := strings.Join([]string{username, password}, ":")

		foundCredentials := username != ""
		if foundCredentials {
			if strings.HasPrefix(url, "http://") {
				return strings.Replace(url, "http://", fmt.Sprintf("http://%s@", cred), 1), foundCredentials
			} else if strings.HasPrefix(url, "https://") {
				return strings.Replace(url, "https://", fmt.Sprintf("https://%s@", cred), 1), foundCredentials
			} else if strings.HasPrefix(url, "file:/") {
				log.Debugf("Skipping credentials for file based repository %q, feel free to open an issue on github.com/updatecli/updatecli", url)
			} else {
				return fmt.Sprintf("http://%s@%s", cred, url), foundCredentials
			}
		}

		return url, foundCredentials
	}

	for _, repo := range repositories {
		// If settings.xml is empty, we add the repository URL to the list
		// as nothing can modify repository URL from the pom.xml
		if len(settingsXMLMap) == 0 {
			mavenRepositories = append(mavenRepositories, repo.URL)
			continue
		}

		for path, settings := range settingsXMLMap {

			mirrorURL, mirrorID, foundMirrorOf := settings.getMatchingMirrorOf(repo.ID, repo.URL)

			if mirrorURL != "" && mirrorID == "" {
				log.Warningf("Found proxy URL %q for repository %q in config file %q but no mirror ID", mirrorURL, repo.ID, path)
			} else if mirrorURL == "" && mirrorID != "" {
				log.Warningf("Found mirror ID %q for repository %q in config file %q but no proxy URL", mirrorID, repo.ID, path)
			}

			if foundMirrorOf {
				log.Debugf("Found proxy URL %q for repository %q in config file %q", mirrorURL, mirrorID, path)
				mirrorURL, _ = setCredentials(settings, mirrorURL, mirrorID)
				mavenRepositories = append(mavenRepositories, mirrorURL)
				continue
			}

			if repoURL, foundCredentials := setCredentials(settings, repo.URL, repo.ID); foundCredentials {
				mavenRepositories = append(mavenRepositories, repoURL)
				continue
			}

			if mavenMirrorURLFromEnv != "" {
				mavenRepositories = append(mavenRepositories, mavenMirrorURLFromEnv)
				continue
			}

			mavenRepositories = append(mavenRepositories, repo.URL)
		}
	}

	if len(repositories) == 0 {
		mavenCentralURL := "https://repo.maven.apache.org/maven2"
		for _, settings := range settingsXMLMap {

			mirrorURL, mirrorID := settings.getCentralMatchingMirrorOf()

			if mirrorURL != "" {
				if repoURL, foundCredentials := setCredentials(settings, mirrorURL, mirrorID); foundCredentials {
					mavenRepositories = append(mavenRepositories, repoURL)
					continue
				}
			}

			if repoURL, foundCredentials := setCredentials(settings, mavenCentralURL, "central"); foundCredentials {
				mavenRepositories = append(mavenRepositories, repoURL)
				continue
			}
		}
	}

	return mavenRepositories
}
