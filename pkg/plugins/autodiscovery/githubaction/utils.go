package githubaction

import (
	"io/fs"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// searchWorkflowFiles will look, recursively, for every files containing a GitHub action workflow from a root directory.
func (g *GitHubAction) searchWorkflowFiles(rootDir string, files []string) error {

	err := filepath.Walk(rootDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			logrus.Debugf("something went wrong while walking in %q: %v\n", path, err)
			return err
		}

		for _, foundFile := range files {
			if !info.IsDir() {
				match, err := filepath.Match(foundFile, info.Name())
				if err != nil {
					continue
				}

				// if file doesn't match the pattern, skip it
				if !match {
					continue
				}

				// Ensure our file is in a .github/workflows directory
				workflow := filepath.Dir(path)
				if filepath.Base(workflow) != "workflows" {
					continue
				}

				github := filepath.Dir(workflow)
				if filepath.Base(github) != ".github" {
					continue
				}

				g.workflowFiles = append(g.workflowFiles, path)
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	logrus.Debugf("%d GitHub workflow(s) found", len(g.workflowFiles))

	return nil
}
