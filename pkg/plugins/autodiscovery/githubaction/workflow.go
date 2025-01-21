package githubaction

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
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

	var root yaml.Node
	err = yaml.Unmarshal(data, &root)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling YAML file %s: %w", filename, err)
	}

	var workflow Workflow
	parseWorkflowNode(&root, &workflow)
	return &workflow, nil
}

// parseWorkflowNode extracts the Workflow structure and captures comments.
func parseWorkflowNode(root *yaml.Node, workflow *Workflow) {
	workflow.Jobs = make(map[string]Job)
	for _, jobNode := range root.Content {
		if jobNode.Kind == yaml.MappingNode {
			for i, keyNode := range jobNode.Content {
				if keyNode.Value == "jobs" {
					jobsNode := jobNode.Content[i+1]
					for j := 0; j < len(jobsNode.Content); j += 2 {
						jobKey := jobsNode.Content[j]
						jobValue := jobsNode.Content[j+1]
						jobName := jobKey.Value
						if _, exists := workflow.Jobs[jobName]; !exists {
							workflow.Jobs[jobName] = Job{}
						}

						for k, stepKey := range jobValue.Content {
							if stepKey.Value == "steps" {
								stepsNode := jobValue.Content[k+1]
								for l := 0; l < len(stepsNode.Content); l++ {
									stepNode := stepsNode.Content[l]
									if stepNode.Kind == yaml.MappingNode {
										var step Step

										for m := 0; m < len(stepNode.Content); m += 2 {
											stepField := stepNode.Content[m]
											stepValue := stepNode.Content[m+1]

											switch stepField.Value {
											case "name":
												step.Name = stepValue.Value
											case "uses":
												step.Uses = stepValue.Value
												step.CommentDigest = strings.TrimSpace(strings.TrimPrefix(stepValue.LineComment, "#"))
											}
										}

										job := workflow.Jobs[jobName]
										job.Steps = append(job.Steps, step)
										workflow.Jobs[jobName] = job
									}
								}
							}
						}
					}
				}
			}
		}
	}
}
