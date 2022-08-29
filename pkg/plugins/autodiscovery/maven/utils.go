package maven

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/beevik/etree"
)

// searchPomFiles will look, recursively, for every files named Chart.yaml from a root directory.
func searchPomFiles(rootDir string, files []string) ([]string, error) {

	pomFiles := []string{}

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
func getRepositoriesFromPom(doc *etree.Document) []repository {
	repositories := []repository{}

	repositoriesPath := "//project/repositories/*"

	for _, repositoryElem := range doc.FindElements(repositoriesPath) {
		URLElem := repositoryElem.FindElement("./url")

		if URLElem == nil {
			continue
		}

		repo := repository{
			URL: URLElem.Text(),
		}

		idElem := repositoryElem.FindElement("./id")
		if idElem != nil {
			repo.ID = idElem.Text()
		}

		repositories = append(repositories, repo)

	}

	return repositories
}

// getDependenciesFromPom parse a pom.xml and return all dependencies
func getDependenciesFromPom(doc *etree.Document) []dependency {
	dependencies := []dependency{}

	dependenciesPath := "//project/dependencies/*"
	for _, dependencyElem := range doc.FindElements(dependenciesPath) {
		GroupIDElem := dependencyElem.FindElement("./groupId")

		if GroupIDElem == nil {
			continue
		}
		dep := dependency{
			GroupID: GroupIDElem.Text(),
		}

		artifactIDElem := dependencyElem.FindElement("./artifactId")
		if artifactIDElem != nil {
			dep.ArtifactID = artifactIDElem.Text()
		}

		versionElem := dependencyElem.FindElement("./version")
		if versionElem != nil {
			dep.Version = versionElem.Text()
		}

		dependencies = append(dependencies, dep)

	}
	return dependencies
}
