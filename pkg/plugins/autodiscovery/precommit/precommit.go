package precommit

import (
	"os"

	goyaml "go.yaml.in/yaml/v3"
)

// precommitRepo represent a precommit repo
type precommitRepo struct {
	Repo string `yaml:"repo"`
	Rev  string `yaml:"rev"`
}

// precommitData represent the useful struct content of .pre-commit-config.yaml
type precommitData struct {
	Repos []precommitRepo `yaml:"repos,omitempty"`
}

// loadPrecommitData read a file and return its content
func loadPrecommitData(filename string) (*precommitData, error) {

	rawFileContent, _ := os.ReadFile(filename)
	var data precommitData

	err := goyaml.Unmarshal(rawFileContent, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}
