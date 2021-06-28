package engine

import (
	"github.com/heimdalr/dag"
	"github.com/olblak/updateCli/pkg/core/engine/condition"
	"github.com/olblak/updateCli/pkg/core/engine/source"
	"github.com/olblak/updateCli/pkg/core/engine/target"
)

/*
	I can't find a good way to avoid code duplication
	with the three following functions
*/

// SortedSourcesKeys return a a list of resources by building a DAG
func SortedSourcesKeys(sources *map[string]source.Source) (result []string, err error) {

	d := dag.NewDAG()

	index := map[string]string{}

	index["root"], err = d.AddVertex("root")

	if err != nil {
		return result, err
	}

	// Init Vertice
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
		if len(s.DependsOn) > 0 {
			for _, dep := range s.DependsOn {
				err = d.AddEdge(index[key], index[dep])
				if err != nil {
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

	// Init Vertice
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
		if len(s.DependsOn) > 0 {
			for _, dep := range s.DependsOn {
				err = d.AddEdge(index[key], index[dep])
				if err != nil {
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

	// Init Vertice
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
		if len(s.DependsOn) > 0 {
			for _, dep := range s.DependsOn {
				err = d.AddEdge(index[key], index[dep])
				if err != nil {
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
