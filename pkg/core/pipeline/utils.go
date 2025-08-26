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

		if leaf.DependsOnChange && dependencyResult.Category != targetCategory {
			continue
		}
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
					// Or operator but dep is not changed
					shouldSkip = false
				}
			} else {
				if dependencyResult.Result == result.SUCCESS {
					// Or operator but dep is failed
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
	for _, node := range tmpl.Root.Nodes {
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

// Define constants for valid GraphFlavor values
const (
	GraphFlavorDot     = "dot"
	GraphFlavorMermaid = "mermaid"
)

// ValidateGraphFlavor checks if the GraphFlavor value is valid
func ValidateGraphFlavor(flavor string) error {
	switch flavor {
	case GraphFlavorDot, GraphFlavorMermaid:
		return nil
	default:
		return fmt.Errorf("invalid graph flavor %q: must be 'dot' or 'mermaid'", flavor)
	}
}

func (p *Pipeline) traverseAndWriteGraph(d *dag.DAG, node string, graphFlavor string, graphOutput *strings.Builder, visited map[string]bool) error {
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
		var shape, color, kind, openingBracket, closingBracket string
		switch nodeType {
		case sourceCategory:
			shape = "ellipse"
			color = "lightblue"
			openingBracket = "(["
			closingBracket = "])"
			if source, ok := p.Sources[name]; ok {
				if source.Config.Name != "" {
					name = source.Config.Name
				}
				kind = source.Config.Kind
			}
		case conditionCategory:
			shape = "diamond"
			color = "orange"
			openingBracket = "{"
			closingBracket = "}"
			if condition, ok := p.Conditions[name]; ok {
				if condition.Config.Name != "" {
					name = condition.Config.Name
				}
				kind = condition.Config.Kind
			}
		case targetCategory:
			shape = "box"
			color = "lightyellow"
			openingBracket = "("
			closingBracket = ")"
			if target, ok := p.Targets[name]; ok {
				if target.Config.Name != "" {
					name = target.Config.Name
				}
				kind = target.Config.Kind
			}
		}

		switch graphFlavor {
		case GraphFlavorDot:

			fmt.Fprintf(
				graphOutput,
				"    %q [label=\"%s (%s)\", shape=%s, style=filled, color=%s];\n",
				node,
				strings.ReplaceAll(name, `"`, `\"`),
				kind,
				shape,
				color,
			)

		case GraphFlavorMermaid:
			fmt.Fprintf(
				graphOutput,
				"    %s%s\"%s (%s)\"%s\n",
				node,
				openingBracket,
				strings.ReplaceAll(name, `"`, `:#quot;`),
				kind,
				closingBracket,
			)
		default:
			logrus.Warningf("Unsupported graph flavor: %s", graphFlavor)
		}
	}
	for successor := range successors {
		if node != rootVertex {
			switch graphFlavor {
			case GraphFlavorDot:
				fmt.Fprintf(
					graphOutput,
					"    %q -> %q;\n",
					node,
					successor,
				)
			case GraphFlavorMermaid:
				fmt.Fprintf(
					graphOutput,
					"    %s --> %s\n",
					node,
					successor,
				)
			default:
				logrus.Warningf("Unsupported graph flavor: %s", graphFlavor)
			}
		}
		err = p.traverseAndWriteGraph(d, successor, graphFlavor, graphOutput, visited)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Pipeline) Graph(flavor string) (string, error) {

	var graphOutput strings.Builder

	switch flavor {
	case GraphFlavorDot:
		graphOutput.WriteString("digraph G {\n")
	case GraphFlavorMermaid:
		graphOutput.WriteString("graph TD\n")
	default:
		return "", fmt.Errorf("unsupported graph flavor: %s", flavor)
	}

	resources, err := p.SortedResources()
	if err != nil {
		return "", err
	}
	resources.ReduceTransitively()
	visited := make(map[string]bool)

	err = p.traverseAndWriteGraph(resources, rootVertex, flavor, &graphOutput, visited)
	if err != nil {
		return "", err
	}
	if flavor == GraphFlavorDot {
		graphOutput.WriteString("}\n")
	}
	return graphOutput.String(), nil
}
