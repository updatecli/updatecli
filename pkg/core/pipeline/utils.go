package pipeline

import (
	"fmt"
	"strings"
	"text/template"
	"text/template/parse"

	"github.com/heimdalr/dag"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// shouldSkipREsource checks if the resource dependsOn conditions are met.
// if not, the resource is skipped.
// a dependsOn value must follow one of the three following format:
// dependsOn:
//   - (resourceType#)resourceID
//   - (resourceType#)resourceID:and
//   - (resourceType#)resourceID:or
//
// where `resourceType` is the type of resource of the dependent, if none are specified, it defaults to its own type
// `resourceID` is the id of another resource in the manifest
// "and" the boolean operator is optional and can be used to specify that all conditions must be met
// "or" the boolean operator is optional and can be used to specify that at least one condition must be met
// if the boolean operator is not provided, it defaults to "and"
func (p *Pipeline) shouldSkipResource(leaf *Node, depsResults map[string]*Node) bool {
	// exit early
	if len(leaf.DependsOn) == 0 {
		return false
	}

	shouldSkip := true
	for _, dependency := range leaf.DependsOn {
		dependencyResult := depsResults[dependency.ID]
		booleanOperator := dependency.Operator

		switch booleanOperator {
		case andBooleanOperator:
			if leaf.DependsOnChange && dependencyResult.Category == targetCategory {
				if !dependencyResult.Changed {
					// And operator but dep is not changed
					return true
				}
			} else {
				if dependencyResult.Result == result.FAILURE {
					// And operator but dep is failed
					return true
				}
			}
			shouldSkip = false
		case orBooleanOperator:
			if leaf.DependsOnChange && dependencyResult.Category == targetCategory {
				if dependencyResult.Changed {
					// And operator but dep is not changed
					shouldSkip = false
				}
			} else {
				if dependencyResult.Result == result.SUCCESS {
					// And operator but dep is failed
					shouldSkip = false
				}
			}
		}
	}
	return shouldSkip

}

// ExtractCustomKeys parses a Go template and extracts custom keys from
// specific template actions: {{ source "sourceId" }}, {{ condition "conditionid" }},
// and {{ target "targetid" }}. It returns a map where the keys are the action types
// ("source", "condition", "target") and the values are slices of strings representing
// the IDs extracted from the corresponding actions in the template.
func ExtractDepsFromTemplate(tmplStr string) ([]string, error) {
	tmpl, err := template.New("dummy").
		Funcs(template.FuncMap{
			"pipeline":  func(id string) string { return id },
			"source":    func(id string) string { return id },
			"condition": func(id string) string { return id },
			"target":    func(id string) string { return id },
		}).Parse(tmplStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing template: %v", err)
	}
	results := []string{}

	// Walk through the parsed template' s tree nodes
	for _, node := range tmpl.Tree.Root.Nodes {
		if actionNode, ok := node.(*parse.ActionNode); ok {
			for _, command := range actionNode.Pipe.Cmds {
				if len(command.Args) > 1 {
					if identifierNode, ok := command.Args[0].(*parse.IdentifierNode); ok {
						if stringNode, ok := command.Args[1].(*parse.StringNode); ok {
							if identifierNode.Ident == sourceCategory ||
								identifierNode.Ident == conditionCategory ||
								identifierNode.Ident == targetCategory ||
								identifierNode.Ident == pipelineCategory {
								results = append(results, fmt.Sprintf("%s#%s", identifierNode.Ident, stringNode.Text))
							}
						}
					}
				}
			}
		}
	}
	return results, nil
}

func (p *Pipeline) traverseAndWriteDot(d *dag.DAG, node string, dotOutput *strings.Builder, visited map[string]bool) error {
	if visited[node] {
		return nil
	}
	visited[node] = true

	successors, err := d.GetDescendants(node)
	if err != nil {
		return err
	}

	if node != rootVertex {

		parts := strings.Split(node, "#")
		if len(parts) <= 1 {
			return nil
		}
		nodeType := parts[0]
		name := strings.Join(parts[1:], "#")
		var shape, color, kind string
		switch nodeType {
		case sourceCategory:
			shape = "ellipse"
			color = "lightblue"
			if source, ok := p.Sources[name]; ok {
				if source.Config.Name != "" {
					name = source.Config.Name
				}
				kind = source.Config.Kind
			}
		case conditionCategory:
			shape = "diamond"
			color = "orange"
			if condition, ok := p.Conditions[name]; ok {
				if condition.Config.Name != "" {
					name = condition.Config.Name
				}
				kind = condition.Config.Kind
			}
		case targetCategory:
			shape = "box"
			color = "lightyellow"
			if target, ok := p.Targets[name]; ok {
				if target.Config.Name != "" {
					name = target.Config.Name
				}
				kind = target.Config.Kind
			}
		}

		dotOutput.WriteString(
			fmt.Sprintf(
				"    %q [label=\"%s (%s)\", shape=%s, style=filled, color=%s];\n",
				node,
				strings.ReplaceAll(name, `"`, `\"`),
				kind,
				shape,
				color,
			),
		)
	}
	for successor := range successors {
		if node != rootVertex {
			dotOutput.WriteString(
				fmt.Sprintf(
					"    %q -> %q;\n",
					node,
					successor,
				),
			)
		}
		err = p.traverseAndWriteDot(d, successor, dotOutput, visited)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Pipeline) Graph() error {

	resources, err := p.SortedResources()
	if err != nil {
		return err
	}
	resources.ReduceTransitively()
	// Start writing the DOT graph
	var dotOutput strings.Builder
	dotOutput.WriteString("digraph G {\n")
	visited := make(map[string]bool)
	err = p.traverseAndWriteDot(resources, rootVertex, &dotOutput, visited)
	if err != nil {
		return err
	}
	dotOutput.WriteString("}\n")
	logrus.Infof("%s", dotOutput.String())
	return nil
}
