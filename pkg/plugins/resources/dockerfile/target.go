package dockerfile

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/sirupsen/logrus"
)

// Target updates a targeted Dockerfile
func (d *Dockerfile) Target(source, workingDir string, dryRun bool) (changed bool, files []string, message string, err error) {
	filePath := d.spec.File
	if !filepath.IsAbs(d.spec.File) {
		filePath = filepath.Join(workingDir, d.spec.File)
		logrus.Debugf("Relative path detected: changing to absolute path from SCM: %q", filePath)
	}

	dockerfileContent, err := d.contentRetriever.ReadAll(filePath)
	if err != nil {
		return false, files, message, err
	}

	logrus.Infof("\n🐋 On (Docker)file %q:\n\n", d.spec.File)

	newDockerfileContent, changedLines, err := d.parser.ReplaceInstructions([]byte(dockerfileContent), source)
	if err != nil {
		return false, files, message, err
	}

	if len(changedLines) == 0 {
		return false, files, message, err
	}

	lines := []int{}
	for idx := range changedLines {
		lines = append(lines, idx)
	}
	sort.Ints(lines)

	message = fmt.Sprintf("changed lines %v of file %q", lines, d.spec.File)
	files = append(files, d.spec.File)

	if !dryRun {
		// Write the new Dockerfile content from buffer to file
		err := d.contentRetriever.WriteToFile(string(newDockerfileContent), filePath)
		if err != nil {
			return false, files, message, err
		}
	}

	return true, files, message, err
}
