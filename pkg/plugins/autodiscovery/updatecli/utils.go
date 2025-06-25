package updatecli

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/compose"
	goyaml "go.yaml.in/yaml/v3"
)

// searchUpdatecliComposeFiles search, recursively, for every Updatecli compose files starting from a root directory.
func searchUpdatecliComposeFiles(rootDir string, files []string) ([]string, error) {

	composeFiles := []string{}

	logrus.Debugf("Looking for Updatecli Compose manifest(s) in %q", rootDir)

	// To do switch to WalkDir which is more efficient, introduced in 1.16
	err := filepath.Walk(rootDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		for _, f := range files {
			match, err := filepath.Match(f, info.Name())
			if err != nil {
				logrus.Errorln(err)
				continue
			}
			if match {
				composeFiles = append(composeFiles, path)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	logrus.Debugf("%d potential Updatecli manifest(s) found", len(composeFiles))

	return composeFiles, nil
}

// getComposeFileMetadata loads file content from an Updatecli compose file.
func getComposeFileMetadata(filename string) (*compose.Spec, error) {

	var composeFile compose.Spec

	if _, err := os.Stat(filename); err != nil {
		return nil, err
	}

	v, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer v.Close()

	content, err := io.ReadAll(v)
	if err != nil {
		return nil, err
	}

	err = goyaml.Unmarshal(content, &composeFile)

	if err != nil {
		return nil, err
	}

	if len(composeFile.Policies) == 0 {
		return nil, nil
	}

	for _, value := range composeFile.Policies {
		logrus.Debugf("Name: %q\n", value.Policy)
	}

	return &composeFile, nil
}

func getPolicyName(policy string) (name, version string, err error) {
	policyNameArray := strings.Split(policy, "@")
	policyNameArray = strings.Split(policyNameArray[0], ":")

	if len(policyNameArray) > 1 {
		version = policyNameArray[1]
	}
	name = policyNameArray[0]

	if name == "" {
		return "", "", fmt.Errorf("policy name is empty")
	}
	return name, version, nil
}
