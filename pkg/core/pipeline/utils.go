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

// shouldSkipResource checks if the resource dependsOn conditions are met.
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
				// Condition dependencies must explicitly succeed. Any non-success
				// condition result must not unlock downstream resources.
				if dependencyResult.Category == conditionCategory && dependencyResult.Result != result.SUCCESS {
					return true
				}
				// A skipped source doesn't provide any value, so running its dependents
				// would have them consume an empty source input.
				if dependencyResult.Category == sourceCategory && dependencyResult.Result == result.SKIPPED {
					return true
				}
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

// pipelineQueryDependency translates the query of a {{ pipeline "..." }} action
// into the resource it reads from, such as `sources.foo.output` into `source#foo`.
//
// Unlike {{ source "foo" }}, {{ pipeline "..." }} takes a path into the pipeline
// configuration rather than a resource id, so only queries rooted at a resource
// map describe a dependency. Any other query, such as `name`, depends on nothing.
// The lookup is case insensitive, as getFieldValueByQuery is.
func pipelineQueryDependency(query string) (string, bool) {
	parts := strings.Split(query, ".")
	if len(parts) < 2 || parts[1] == "" {
		return "", false
	}

	category := ""
	switch strings.ToLower(parts[0]) {
	case "sources":
		category = sourceCategory
	case "conditions":
		category = conditionCategory
	case "targets":
		category = targetCategory
	default:
		return "", false
	}

	return fmt.Sprintf("%s#%s", category, parts[1]), true
}

// ExtractDepsFromTemplate parses a Go template and returns the dependencies its
// actions imply: {{ source "sourceId" }}, {{ condition "conditionid" }} and
// {{ target "targetid" }} each depend on the resource they name, and
// {{ pipeline "sources.sourceId.output" }} on the resource its query reads from.
//
// Each dependency records the action it came from, so that one naming a resource
// the pipeline does not define can be reported against what the user wrote.
func ExtractDepsFromTemplate(tmplStr string) ([]Dependency, error) {
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
	results := []Dependency{}

	// Walk through the parsed template' s tree nodes
	for _, node := range tmpl.Root.Nodes {
		if actionNode, ok := node.(*parse.ActionNode); ok {
			for _, command := range actionNode.Pipe.Cmds {
				if len(command.Args) > 1 {
					if identifierNode, ok := command.Args[0].(*parse.IdentifierNode); ok {
						if stringNode, ok := command.Args[1].(*parse.StringNode); ok {
							from := fmt.Sprintf("the template reference {{ %s %q }}", identifierNode.Ident, stringNode.Text)

							switch identifierNode.Ident {
							case sourceCategory, conditionCategory, targetCategory:
								results = append(results, Dependency{
									ID:       fmt.Sprintf("%s#%s", identifierNode.Ident, stringNode.Text),
									Operator: andBooleanOperator,
									From:     from,
								})
							case pipelineCategory:
								if id, ok := pipelineQueryDependency(stringNode.Text); ok {
									results = append(results, Dependency{
										ID:       id,
										Operator: andBooleanOperator,
										From:     from,
									})
								}
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
