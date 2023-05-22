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
