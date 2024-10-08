package pipeline

import (
	"errors"
	"fmt"
	"strings"

	"github.com/heimdalr/dag"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/condition"
	"github.com/updatecli/updatecli/pkg/core/pipeline/source"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
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

func addResourceToDag(dag *dag.DAG, id, Category string, DependsOn []string, DependsOnChange bool) (err error) {
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

func handleResourceDependencies(dag *dag.DAG, ID, Category string, additionalDependencies []string) (err error) {
	myId := fmt.Sprintf("%s#%s", Category, ID)
	// Update vertices dependencies based on depends_on
	rawNode, err := dag.GetVertex(myId)
	if err != nil {
		return nil
	}
	node, ok := rawNode.(Node)
	if !ok {
		return fmt.Errorf("Could not reconstruct node")
	}

	var deps []string
	deps = append(deps, additionalDependencies...)
	for _, dep := range node.DependsOn {
		deps = append(deps, dep.ID)
	}
	for _, depId := range deps {
		_, err = dag.GetVertex(depId)
		if err != nil {
			logrus.Errorf("no valid depends_on value: %q", depId)
			return ErrNotValidDependsOn
		}
		err = dag.AddEdge(depId, myId)
		if err != nil {
			if strings.Contains(err.Error(), "would create a loop") {
				logrus.Errorf("Dependency loop detected between %q and %q",
					depId,
					myId)
				return ErrDependsOnLoopDetected
			}
			return err
		}
	}
	if Category == conditionCategory {

	}
	return err
}

// SortedResources return a list of resources by building a DAG
func SortedResources(sources *map[string]source.Source, conditions *map[string]condition.Condition, targets *map[string]target.Target) (result *dag.DAG, err error) {

	d := dag.NewDAG()
	d.Options(dag.Options{VertexHashFunc: func(v interface{}) interface{} {
		switch n := v.(type) {
		case Node:
			return n.ID
		}
		return v
	}})

	err = d.AddVertexByID(rootVertex, Node{ID: rootVertex, Category: dummyCategory})

	if err != nil {
		return result, err
	}
	// Add sources to dag
	for id, resource := range *sources {
		err = addResourceToDag(d, id, sourceCategory, resource.Config.DependsOn, false)
		if err != nil {
			return result, err
		}
	}
	// Add conditions to dag
	for id, resource := range *conditions {
		err = addResourceToDag(d, id, conditionCategory, resource.Config.DependsOn, false)
		if err != nil {
			return result, err
		}
	}
	// Add target to dag
	for id, resource := range *targets {
		err = addResourceToDag(d, id, targetCategory, resource.Config.DependsOn, resource.Config.DependsOnChange)
		if err != nil {
			return result, err
		}
	}
	// Now that the dag is complete, we can add the `depends_on` vertice
	for id := range *sources {
		err = handleResourceDependencies(d, id, sourceCategory, nil)
	}
	for id := range *conditions {
		additionalDepIds := []string{}
		condition := (*conditions)[id]
		if condition.Config.SourceID != "" {
			additionalDepIds = append(additionalDepIds, fmt.Sprintf("source#%s", condition.Config.SourceID))
		}
		err = handleResourceDependencies(d, id, conditionCategory, additionalDepIds)
	}
	for id := range *targets {
		additionalDepIds := []string{}
		target := (*targets)[id]
		if target.Config.SourceID != "" {
			additionalDepIds = append(additionalDepIds, fmt.Sprintf("source#%s", target.Config.SourceID))
		}
		// For targets we need to handle the condition sorting
		// By default, a target depends on all conditions, and they are treated as an and dependency
		// This behavior can be deactivated by setting DisableConditions to false
		if !target.Config.DisableConditions {
			switch len(target.Config.DeprecatedConditionIDs) > 0 {
			case true:
				for _, conditionID := range target.Config.DeprecatedConditionIDs {
					additionalDepIds = append(additionalDepIds, fmt.Sprintf("condition#%s", conditionID))
				}
			case false:
				// if no condition is defined, we evaluate all conditions
				for conditionID := range *conditions {
					additionalDepIds = append(additionalDepIds, fmt.Sprintf("condition#%s", conditionID))
				}
			}
		}
		err = handleResourceDependencies(d, id, targetCategory, additionalDepIds)
	}
	if err != nil {
		return result, err
	}
	d.ReduceTransitively()
	return d, err
}
