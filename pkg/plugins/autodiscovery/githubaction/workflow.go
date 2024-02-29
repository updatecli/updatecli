package githubaction

import (
	"fmt"
	"os"

	"sigs.k8s.io/yaml"
)

// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstepsuses

type Workflow struct {
	// Name of the workflow
	Name string         `yaml:"name, omitempty"`
	Jobs map[string]Job `yaml:"jobs, omitempty"`
}

type Job struct {
	Steps []Step `yaml:"steps, omitempty"`
}

type Step struct {
	Name string `yaml:"name, omitempty"`
	Uses string `yaml:"uses, omitempty"`
}

func loadGitHubActionWorkflow(filename string) (*Workflow, error) {

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("opening file %s:%s", filename, err)
	}

	w := Workflow{}
	err = yaml.Unmarshal(data, &w)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling GitHub action workflow file %s: %s", filename, err)
	}

	return &w, nil
}
