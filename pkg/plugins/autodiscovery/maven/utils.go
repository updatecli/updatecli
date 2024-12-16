package maven

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/beevik/etree"
	"github.com/sirupsen/logrus"
)

// searchPomFiles will look, recursively, for every files named Chart.yaml from a root directory.
func searchPomFiles(rootDir string, files []string) ([]string, error) {

	pomFiles := []string{}

	logrus.Debugf("Looking for Maven pom.xml files in %q", rootDir)

	// To do switch to WalkDir which is more efficient, introduced in 1.16
	err := filepath.Walk(rootDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
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
		p.GroupID = elem.Text()
	}

	if elem := parent.FindElement("./artifactId"); elem != nil {
		p.ArtifactID = elem.Text()
	}

	if elem := parent.FindElement("./version"); elem != nil {
		p.Version = elem.Text()
	}

	if elem := parent.FindElement("./packaging"); elem != nil {
		p.Packaging = elem.Text()
	}

	if elem := parent.FindElement("./relativePath"); elem != nil {
		p.RelativePath = elem.Text()
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
			rep.URL = elem.Text()
		}

		if elem := repositoryElem.FindElement("./id"); elem != nil {
			rep.ID = elem.Text()
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
			dep.GroupID = elem.Text()
		}

		if elem := dependencyElem.FindElement("./artifactId"); elem != nil {
			dep.ArtifactID = elem.Text()
		}

		if elem := dependencyElem.FindElement("./version"); elem != nil {
			dep.Version = elem.Text()
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
			dep.GroupID = elem.Text()
		}

		if elem := dependencyElem.FindElement("./artifactId"); elem != nil {
			dep.ArtifactID = elem.Text()
		}

		if elem := dependencyElem.FindElement("./version"); elem != nil {
			dep.Version = elem.Text()
		}

		dependencies = append(dependencies, dep)

	}
	return dependencies
}

// getMavenRepositoriesURL returns the URL for all repository while trying to identify potential credentials and mirrors
// It will look for repositories in the pom.xml and in the settings.xml
func getMavenRepositoriesURL(pomFile string, doc *etree.Document) (mavenRepositories []string) {

	repositories := getRepositoriesFromPom(doc)

	settingsXMLMap := map[string]Settings{}

	for _, path := range getSettingsXMLPath(filepath.Dir(pomFile)) {
		settings := readSettingsXML(path)
		if settings == nil {
			logrus.Debugf("No settings.xml content found in %q", path)
			continue
		}

		settingsXMLRepositories := settings.getRepositoriesFromSettingsXML()

		if len(settingsXMLRepositories) > 0 {
			logrus.Debugf("Found %d repositories in %q", len(settingsXMLRepositories), path)
			for _, repo := range settingsXMLRepositories {
				logrus.Debugf("\t* repository %q %q", repo.ID, repo.URL)
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
				logrus.Debugf("Skipping credentials for file based repository %q, feel free to open an issue on github.com/updatecli/updatecli", url)
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
				logrus.Warningf("Found proxy URL %q for repository %q in config file %q but no mirror ID", mirrorURL, repo.ID, path)
			} else if mirrorURL == "" && mirrorID != "" {
				logrus.Warningf("Found mirror ID %q for repository %q in config file %q but no proxy URL", mirrorID, repo.ID, path)
			}

			if foundMirrorOf {
				logrus.Debugf("Found proxy URL %q for repository %q in config file %q", mirrorURL, mirrorID, path)
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
