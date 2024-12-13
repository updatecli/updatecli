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

	centralDefined := false

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

		if rep.ID == "central" {
			centralDefined = true
		}

		repositories = append(repositories, rep)

	}

	if !centralDefined {
		repositories = append(
			repositories,
			repository{ID: "central", URL: "https://repo.maven.apache.org/maven2"},
		)
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

// getRepositoryURL returns the URL of a repository while trying to identify potential maven proxy settings
func getRepositoryURL(pomDirname string, repo repository) string {

	setCredentials := func(settings Settings, url, id string) (string, bool) {

		username, password, foundCredentials := settings.getMatchingServerCredentials(id)
		cred := strings.Join([]string{username, password}, ":")

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

	for _, path := range getSettingsXMLPath(pomDirname) {

		settings := readSettingsXML(path)
		if settings == nil {
			continue
		}

		mirrorURL, mirrorID, foundMirrorOf := settings.getMatchingMirrorOf(repo.ID, repo.URL)

		if mirrorURL != "" && mirrorID == "" {
			logrus.Warningf("Found proxy URL %q for repository %q in config file %q but no mirror ID", mirrorURL, repo.ID, path)
		} else if mirrorURL == "" && mirrorID != "" {
			logrus.Warningf("Found mirror ID %q for repository %q in config file %q but no proxy URL", mirrorID, repo.ID, path)
		}

		if foundMirrorOf {
			logrus.Debugf("Found proxy URL %q for repository %q in config file %q", mirrorURL, mirrorID, path)

			mirrorURL, _ = setCredentials(*settings, mirrorURL, mirrorID)

			return mirrorURL
		}

		if repoURL, foundCredentials := setCredentials(*settings, repo.URL, repo.ID); foundCredentials {
			return repoURL
		}
	}

	if mavenMirrorURLFromEnv != "" {
		return mavenMirrorURLFromEnv
	}

	return repo.URL
}
