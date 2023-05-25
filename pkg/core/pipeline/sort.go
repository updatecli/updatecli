package pipeline

import (
	"errors"
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

var (
	// ErrNotValidDependsOn is triggered when we define a nonexistent depends on value.
	ErrNotValidDependsOn = errors.New("no valid depends_on value")
	// ErrDependsOnLoopDetected is triggered when we define a dependency loop.
	ErrDependsOnLoopDetected = errors.New("dependency loop detected")
)

// isValidateDependsOn test if we are referencing an exist resource key
func isValidDependsOn(dependsOn string, index map[string]string) bool {

	for val := range index {
		if strings.Compare(dependsOn, val) == 0 {
			return true
		}
	}
	return false
}

// SortedSourcesKeys return a a list of resources by building a DAG
func SortedSourcesKeys(sources *map[string]source.Source) (result []string, err error) {

	d := dag.NewDAG()

	index := map[string]string{}

	index["root"], err = d.AddVertex("root")

	if err != nil {
		return result, err
	}

	// Init Vertices
	for key := range *sources {
		index[key], err = d.AddVertex(key)

		if err != nil {
			return result, err
		}

		err = d.AddEdge(index["root"], index[key])
		if err != nil {
			return result, err
		}
	}

	// Update vertices dependencies based on depends_on
	for key, s := range *sources {
		if len(s.Config.DependsOn) > 0 {
			for _, dep := range s.Config.DependsOn {
				if !isValidDependsOn(dep, index) {
					logrus.Errorf("%s:%q", ErrNotValidDependsOn, dep)
					return result, ErrNotValidDependsOn
				}
				err = d.AddEdge(index[key], index[dep])
				if err != nil {
					if strings.Contains(err.Error(), "would create a loop") {
						logrus.Errorf("Dependency loop detected between Sources[%q] and Sources[%q]",
							key,
							dep)
						return result, ErrDependsOnLoopDetected
					}
					return result, err
				}

			}
		}
	}

	d.ReduceTransitively()

	tmpResult, err := d.GetOrderedDescendants(index["root"])

	if err != nil {
		return result, err
	}

	result = make([]string, len(tmpResult))

	j := 0
	for i := (len(tmpResult) - 1); i >= 0; i-- {
		val, err := d.GetVertex(tmpResult[i])
		if err != nil {
			return result, err
		}
		result[j] = val.(string)
		j++

	}

	if err != nil {
		return result, err
	}

	return result, err
}

// SortedConditionsKeys return a a list of resources by building a DAG
func SortedConditionsKeys(conditions *map[string]condition.Condition) (result []string, err error) {

	d := dag.NewDAG()

	index := map[string]string{}

	index["root"], err = d.AddVertex("root")

	if err != nil {
		return result, err
	}

	// Init Vertices
	for key := range *conditions {
		index[key], err = d.AddVertex(key)

		if err != nil {
			return result, err
		}

		err = d.AddEdge(index["root"], index[key])
		if err != nil {
			return result, err
		}
	}

	// Update vertices dependencies based on depends_on
	for key, s := range *conditions {
		if len(s.Config.DependsOn) > 0 {
			for _, dep := range s.Config.DependsOn {
				if !isValidDependsOn(dep, index) {
					logrus.Errorf("%s:%q", ErrNotValidDependsOn, dep)
					return result, ErrNotValidDependsOn
				}
				err = d.AddEdge(index[key], index[dep])
				if err != nil {
					if strings.Contains(err.Error(), "would create a loop") {
						logrus.Errorf("Dependency loop detected between Conditions[%q] and Conditions[%q]",
							key,
							dep)
						return result, ErrDependsOnLoopDetected
					}
					return result, err
				}
			}
		}
	}

	d.ReduceTransitively()

	tmpResult, err := d.GetOrderedDescendants(index["root"])

	if err != nil {
		return result, err
	}

	result = make([]string, len(tmpResult))

	j := 0
	for i := (len(tmpResult) - 1); i >= 0; i-- {
		val, err := d.GetVertex(tmpResult[i])
		if err != nil {
			return result, err
		}
		result[j] = val.(string)
		j++

	}

	if err != nil {
		return result, err
	}

	return result, err
}

// SortedTargetsKeys return a a list of resources by building a DAG
func SortedTargetsKeys(targets *map[string]target.Target) (result []string, err error) {

	d := dag.NewDAG()

	index := map[string]string{}

	index["root"], err = d.AddVertex("root")

	if err != nil {
		return result, err
	}

	// Init Vertices
	for key := range *targets {
		index[key], err = d.AddVertex(key)

		if err != nil {
			return result, err
		}

		err = d.AddEdge(index["root"], index[key])
		if err != nil {
			return result, err
		}
	}

	// Update vertices dependencies based on depends_on
	for key, s := range *targets {
		if len(s.Config.DependsOn) > 0 {
			for _, dep := range s.Config.DependsOn {
				if !isValidDependsOn(dep, index) {
					logrus.Errorf("%s: %q", ErrNotValidDependsOn, dep)
					return result, ErrNotValidDependsOn
				}
				err = d.AddEdge(index[key], index[dep])
				if err != nil {
					if strings.Contains(err.Error(), "would create a loop") {
						logrus.Errorf("Dependency loop detected between Targets[%q] and Targets[%q]",
							key,
							dep)
						return result, ErrDependsOnLoopDetected
					}
					return result, err
				}
			}
		}
	}

	d.ReduceTransitively()

	tmpResult, err := d.GetOrderedDescendants(index["root"])

	if err != nil {
		return result, err
	}

	result = make([]string, len(tmpResult))

	j := 0
	for i := (len(tmpResult) - 1); i >= 0; i-- {
		val, err := d.GetVertex(tmpResult[i])
		if err != nil {
			return result, err
		}
		result[j] = val.(string)
		j++

	}

	if err != nil {
		return result, err
	}

	return result, err
}
