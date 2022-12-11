package dockerfile

import (
	"sort"

	"github.com/sirupsen/logrus"
)

// Target updates a targeted Dockerfile
func (d *Dockerfile) Target(source string, dryRun bool) (bool, error) {
	dockerfileContent, err := d.contentRetriever.ReadAll(d.spec.File)
	if err != nil {
		return false, err
	}

	logrus.Infof("\n🐋 On (Docker)file %q:\n\n", d.spec.File)

	newDockerfileContent, changedLines, err := d.parser.ReplaceInstructions([]byte(dockerfileContent), source)
	if err != nil {
		return false, err
	}

	if len(changedLines) == 0 {
		return false, err
	}

	lines := []int{}
	for idx := range changedLines {
		lines = append(lines, idx)
	}
	sort.Ints(lines)

	if !dryRun {
		// Write the new Dockerfile content from buffer to file
		err := d.contentRetriever.WriteToFile(string(newDockerfileContent), d.spec.File)
		if err != nil {
			return false, err
		}
	}

	return true, err
}
