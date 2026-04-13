package precommit

import (
	"os"
	"strings"

	goyaml "go.yaml.in/yaml/v3"
)

// precommitRepo represent a precommit repo
type precommitRepo struct {
	Repo       string `yaml:"repo"`
	Rev        string `yaml:"rev"`
	RevComment string
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

	var rootNode goyaml.Node
	err = goyaml.Unmarshal(rawFileContent, &rootNode)
	if err != nil {
		return &data, nil
	}

	if len(rootNode.Content) == 0 {
		return &data, nil
	}

	document := rootNode.Content[0]
	for i := 0; i+1 < len(document.Content); i += 2 {
		keyNode := document.Content[i]
		valueNode := document.Content[i+1]
		if keyNode.Value != "repos" || valueNode.Kind != goyaml.SequenceNode {
			continue
		}

		for repoIndex, repoNode := range valueNode.Content {
			if repoNode.Kind != goyaml.MappingNode {
				continue
			}
			if repoIndex >= len(data.Repos) {
				continue
			}

			var revComment string
			for j := 0; j+1 < len(repoNode.Content); j += 2 {
				repoKey := repoNode.Content[j]
				repoValue := repoNode.Content[j+1]
				switch repoKey.Value {
				case "rev":
					revComment = strings.TrimPrefix(strings.TrimSpace(repoValue.LineComment), "#")
					revComment = strings.TrimSpace(revComment)
				}
			}

			data.Repos[repoIndex].RevComment = revComment
		}
	}

	return &data, nil
}
