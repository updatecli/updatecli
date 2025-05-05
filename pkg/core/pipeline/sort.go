package pipeline

import (
	"errors"
	"fmt"
	"strings"

	"github.com/heimdalr/dag"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

/*
	I can't find a good way to avoid code duplication
	with the three following functions
*/

const (
	rootVertex         string = "root"
	dummyCategory      string = "dummy"
	sourceCategory     string = "source"
	conditionCategory  string = "condition"
	targetCategory     string = "target"
	pipelineCategory   string = "pipeline"
	andBooleanOperator string = "and"
	orBooleanOperator  string = "or"
)

var (
	// ErrNotValidDependsOn is triggered when we define a nonexistent depends on value.
	ErrNotValidDependsOn = errors.New("no valid depends_on value")
	// ErrDependsOnLoopDetected is triggered when we define a dependency loop.
	ErrDependsOnLoopDetected = errors.New("dependency loop detected")
)

type Dependency struct {
	ID       string
	Operator string
}

type Node struct {
	ID              string
	Category        string
	DependsOn       []Dependency
	DependsOnChange bool
	Result          string
	Changed         bool
}

func addResourceToDag(dag *dag.DAG, id, Category string, DependsOn []string, DependsOnChange bool, additionalDependencies []string) (err error) {
	// Add the category to the id
	ID := fmt.Sprintf("%s#%s", Category, id)
	// Craft the dendencies
	var deps []Dependency
	for _, dependency := range DependsOn {
		key, booleanOperator, category := parseDependsOnValue(dependency)
		if category == "" {
			// By default dependencies should be handled inside of one's category
			category = Category
		}
		deps = append(deps, Dependency{ID: fmt.Sprintf("%s#%s", category, key), Operator: booleanOperator})
	}
	for _, dependency := range additionalDependencies {
		deps = append(deps, Dependency{ID: dependency, Operator: andBooleanOperator})
	}
	// Add the node to the graph
	node := Node{ID: ID, Category: Category, DependsOn: deps, DependsOnChange: DependsOnChange}
	err = dag.AddVertexByID(ID, node)
	if err != nil {
		return nil
	}
	// Make the node depends on root
	err = dag.AddEdge(rootVertex, ID)
	return err
}

func handleResourceDependencies(dag *dag.DAG, ID, Category string) (err error) {
	myId := fmt.Sprintf("%s#%s", Category, ID)
	// Update vertex dependencies based on depends_on
	rawNode, err := dag.GetVertex(myId)
	if err != nil {
		return nil
	}
	node, ok := rawNode.(Node)
	if !ok {
		return fmt.Errorf("could not reconstruct node")
	}

	for _, dep := range node.DependsOn {
		_, err = dag.GetVertex(dep.ID)
		if err != nil {
			return ErrNotValidDependsOn
		}
		err = dag.AddEdge(dep.ID, myId)
		if err != nil {
			if strings.Contains(err.Error(), "would create a loop") {
				logrus.Debugf("Dependency loop detected between %q and %q",
					dep.ID,
					myId)
				return ErrDependsOnLoopDetected
			} else if err.Error() == fmt.Sprintf("edge between '%s' and '%s' is already known", dep.ID, myId) {
				// This can happens as we have 4 ways to add dependencies:
				// 1. DependsOn
				// 2. SourceID (For `conditions` and `targets`)
				// 3. ConditionIds (For `targets`)
				// 4. RunTime Deps
				// We can ignore this
				err = nil
			} else {
				return err
			}
		}
	}
	return err
}

// SortedResources return a list of resources by building a DAG
func (p *Pipeline) SortedResources() (result *dag.DAG, err error) {
	d := dag.NewDAG()
	d.Options(dag.Options{VertexHashFunc: func(v interface{}) interface{} {
		switch n := v.(type) {
		case Node:
			return n.ID
		}
		return v
	}})
	// Add a dummy root to ensure we have a starting point for the transversal
	err = d.AddVertexByID(rootVertex, Node{ID: rootVertex, Category: dummyCategory})
	if err != nil {
		return result, err
	}
	// Add sources to dag
	for id, resource := range p.Sources {
		// Marshal to parse runtimeDeps
		s, err := yaml.Marshal(resource.Config)
		if err != nil {
			return result, err
		}
		additionalDepIds, err := ExtractDepsFromTemplate(string(s))
		if err != nil {
			return result, err
		}
		err = addResourceToDag(d, id, sourceCategory, resource.Config.DependsOn, false, additionalDepIds)
		if err != nil {
			return result, err
		}
	}
	// Add conditions to dag
	for id, resource := range p.Conditions {
		// Marshal to parse runtimeDeps
		s, err := yaml.Marshal(resource.Config)
		if err != nil {
			return result, err
		}
		additionalDepIds, err := ExtractDepsFromTemplate(string(s))
		if err != nil {
			return result, err
		}
		if resource.Config.SourceID != "" {
			additionalDepIds = append(additionalDepIds, fmt.Sprintf("source#%s", resource.Config.SourceID))
		}
		err = addResourceToDag(d, id, conditionCategory, resource.Config.DependsOn, false, additionalDepIds)
		if err != nil {
			return result, err
		}
	}
	// Add target to dag
	for id, resource := range p.Targets {
		// Marshal to parse runtimeDeps
		s, err := yaml.Marshal(resource.Config)
		if err != nil {
			return result, err
		}
		additionalDepIds, err := ExtractDepsFromTemplate(string(s))
		if err != nil {
			return result, err
		}
		if resource.Config.SourceID != "" {
			additionalDepIds = append(additionalDepIds, fmt.Sprintf("source#%s", resource.Config.SourceID))
		}
		// For targets we need to handle the condition sorting
		// By default, a target depends on all conditions, and they are treated as an and dependency
		// This behavior can be deactivated by setting DisableConditions to false
		if !resource.Config.DisableConditions {
			// if no condition is defined, we evaluate all conditions
			for conditionID := range p.Conditions {
				additionalDepIds = append(additionalDepIds, fmt.Sprintf("condition#%s", conditionID))
			}
		}
		err = addResourceToDag(d, id, targetCategory, resource.Config.DependsOn, resource.Config.DependsOnChange, additionalDepIds)
		if err != nil {
			return result, err
		}
	}
	// Now that the dag is complete, we can add the `depends_on` vertice
	for id := range p.Sources {
		err = handleResourceDependencies(d, id, sourceCategory)
		if err != nil {
			return result, err
		}
	}
	for id := range p.Conditions {
		err = handleResourceDependencies(d, id, conditionCategory)
		if err != nil {
			return result, err
		}
	}
	for id := range p.Targets {
		err = handleResourceDependencies(d, id, targetCategory)
		if err != nil {
			return result, err
		}
	}
	if err != nil {
		return result, err
	}
	return d, err
}
