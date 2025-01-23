package githubaction

import (
	"fmt"
	"os"
	"strings"

	yaml "github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/parser"
	"github.com/sirupsen/logrus"
)

type Workflow struct {
	Name string         `yaml:"name,omitempty"`
	Jobs map[string]Job `yaml:"jobs,omitempty"`
}

type Job struct {
	Steps     []Step    `yaml:"steps,omitempty"`
	Container Container `yaml:"container,omitempty"`
}

type Container struct {
	Image string `yaml:"image,omitempty"`
}

type Step struct {
	Name          string `yaml:"name,omitempty"`
	Uses          string `yaml:"uses,omitempty"`
	CommentDigest string // Captured comment for the 'uses' field
}

func loadGitHubActionWorkflow(filename string) (*Workflow, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file %s: %w", filename, err)
	}

	var workflow Workflow
	err = yaml.Unmarshal(data, &workflow)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling YAML file %s: %w", filename, err)
	}

	for i := range workflow.Jobs {
		for j := range workflow.Jobs[i].Steps {
			query := fmt.Sprintf("$.jobs.%s.steps[%d].uses", i, j)
			path, err := yaml.PathString(query)
			if err != nil {
				logrus.Debugf("skipping %q, error creating yaml path: %v", query, err)
				continue
			}

			d, err := parser.ParseBytes(data, parser.ParseComments)
			if err != nil {
				logrus.Debugf("skipping %q, error parsing yaml file: %v", query, err)
				continue
			}

			n, err := path.FilterFile(d)
			if err != nil {
				logrus.Debugf("skipping %q, error filtering node: %v", query, err)
				continue
			}

			comment := n.GetComment()
			if comment != nil {
				comment := strings.TrimPrefix(comment.String(), "#")
				comment = strings.TrimPrefix(comment, " ")

				workflow.Jobs[i].Steps[j].CommentDigest = comment
			}
		}
	}

	return &workflow, nil
}
