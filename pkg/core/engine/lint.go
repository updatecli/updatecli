package engine

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

func (e *Engine) Lint(rootDir string) error {
	PrintTitle("Lint Updatecli policies")
	fs := afero.NewOsFs()
	policies, err := processDirectoriesWithPolicyYaml(fs, rootDir)
	if err != nil {
		logrus.Errorf("command failed: %s", err)
		os.Exit(1)
	}

	for _, policy := range policies {
		fmt.Println("Policy directory:", policy)
		_, err := verifyPolicyFiles(fs, policy)
		if err != nil {
			logrus.Errorf("  * %s for policy %s\n", err, policy)
			continue
		}
		// TODO: validate at the end.
	}
	return nil
}

// processDirectoriesWithPolicyYaml returns the folders with the Policy.yaml
func processDirectoriesWithPolicyYaml(fs afero.Fs, rootFolder string) ([]string, error) {
	const policyFileName = "Policy.yaml"

	var dirs []string
	err := afero.Walk(fs, rootFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Mode().IsRegular() && filepath.Base(path) == policyFileName {
			dir := filepath.Dir(path)
			dirs = append(dirs, dir)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return dirs, nil
}

// verifyPolicyFiles checks if the required policy files exist in the given directory.
func verifyPolicyFiles(fs afero.Fs, policyDir string) (bool, error) {
	requiredFiles := []string{"values.yaml", "README.md", "Policy.yaml", "CHANGELOG.md"}

	for _, file := range requiredFiles {
		filePath := filepath.Join(policyDir, file)
		if _, err := fs.Stat(filePath); os.IsNotExist(err) {
			return false, fmt.Errorf("file '%s' missing", filePath)
		} else if err != nil {
			return false, err
		}
	}

	return true, nil
}
