package pipeline

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/heimdalr/dag"
	"github.com/sirupsen/logrus"
	"go.yaml.in/yaml/v3"
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
	// From explains how this dependency was declared, so that an unresolvable one
	// can be reported in the terms the user wrote it in.
	//
	// Only the origins that can actually name a missing resource set it: a
	// `dependson` entry and a template reference such as {{ source "foo" }}. A
	// `sourceid` field is checked by config validation before the graph is built,
	// and the implicit edge every target has on the conditions is drawn from the
	// conditions themselves, so neither can reach an unresolvable dependency.
	From string
}

// unresolvableDependencyError reports a dependency that names no resource of the
// graph. It carries its own message rather than prefixing ErrNotValidDependsOn,
// whose wording says nothing the detailed message does not already say, while
// still matching that sentinel through errors.Is.
type unresolvableDependencyError struct {
	msg string
}

func (e unresolvableDependencyError) Error() string { return e.msg }
func (e unresolvableDependencyError) Unwrap() error { return ErrNotValidDependsOn }

// knownResources lists the ids of every resource of the given category held by
// the graph, so that a typo in an unresolvable dependency is easy to spot.
func knownResources(d *dag.DAG, category string) string {
	ids := []string{}
	for id := range d.GetVertices() {
		if key, found := strings.CutPrefix(id, category+"#"); found {
			ids = append(ids, key)
		}
	}

	if len(ids) == 0 {
		return fmt.Sprintf("no %s is defined in this pipeline", category)
	}

	slices.Sort(ids)

	return fmt.Sprintf("known %ss: %s", category, strings.Join(ids, ", "))
}

type Node struct {
	ID              string
	Category        string
	DependsOn       []Dependency
	DependsOnChange bool
	Result          string
	Changed         bool
}

func addResourceToDag(dag *dag.DAG, id, Category string, DependsOn []string, DependsOnChange bool, additionalDependencies []Dependency) (err error) {
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
		deps = append(deps, Dependency{
			ID:       fmt.Sprintf("%s#%s", category, key),
			Operator: booleanOperator,
			From:     fmt.Sprintf("the dependson entry %q", dependency),
		})
	}
	deps = append(deps, additionalDependencies...)
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
	// Update vertices dependencies based on depends_on
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
			depCategory, depKey, _ := strings.Cut(dep.ID, "#")

			details := []string{}
			if dep.From != "" {
				details = append(details, "declared by "+dep.From)
			}
			details = append(details, knownResources(dag, depCategory))

			return unresolvableDependencyError{
				msg: fmt.Sprintf("%s %q depends on %s %q which is not defined in this pipeline\n\t%s",
					Category, ID,
					depCategory, depKey,
					strings.Join(details, "\n\t"),
				),
			}
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
		additionalDeps, err := ExtractDepsFromTemplate(string(s))
		if err != nil {
			return result, err
		}
		err = addResourceToDag(d, id, sourceCategory, resource.Config.DependsOn, false, additionalDeps)
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
		additionalDeps, err := ExtractDepsFromTemplate(string(s))
		if err != nil {
			return result, err
		}
		if resource.Config.SourceID != "" {
			additionalDeps = append(additionalDeps, Dependency{
				ID:       fmt.Sprintf("%s#%s", sourceCategory, resource.Config.SourceID),
				Operator: andBooleanOperator,
			})
		}
		err = addResourceToDag(d, id, conditionCategory, resource.Config.DependsOn, false, additionalDeps)
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
		additionalDeps, err := ExtractDepsFromTemplate(string(s))
		if err != nil {
			return result, err
		}
		if resource.Config.SourceID != "" {
			additionalDeps = append(additionalDeps, Dependency{
				ID:       fmt.Sprintf("%s#%s", sourceCategory, resource.Config.SourceID),
				Operator: andBooleanOperator,
			})
		}
		// For targets we need to handle the condition sorting
		// By default, a target depends on all conditions, and they are treated as an and dependency
		// This behavior can be deactivated by setting DisableConditions to false
		if !resource.Config.DisableConditions {
			// if no condition is defined, we evaluate all conditions
			for conditionID := range p.Conditions {
				additionalDeps = append(additionalDeps, Dependency{
					ID:       fmt.Sprintf("%s#%s", conditionCategory, conditionID),
					Operator: andBooleanOperator,
				})
			}
		}
		err = addResourceToDag(d, id, targetCategory, resource.Config.DependsOn, resource.Config.DependsOnChange, additionalDeps)
		if err != nil {
			return result, err
		}
	}
	// Now that the dag is complete, we can add the `depends_on` vertex
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
