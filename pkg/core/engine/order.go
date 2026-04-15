package engine

import (
	"fmt"
	"slices"
	"strings"

	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline"
)

type manifestNode struct {
	id        string
	groupID   string
	name      string
	dependsOn []string
	index     int
}

// OrderPipelines resolves manifest-level dependencies and reorders pipelines accordingly.
func (e *Engine) OrderPipelines() error {
	if len(e.Pipelines) == 0 {
		return nil
	}

	nodes := make([]manifestNode, len(e.Pipelines))
	indicesByGroupID := make(map[string][]int, len(e.Pipelines))

	for i, p := range e.Pipelines {
		if p == nil || p.Config == nil {
			return fmt.Errorf("pipeline at index %d is missing configuration", i)
		}

		manifestID := p.Config.ManifestID()
		if manifestID == "" {
			return fmt.Errorf("pipeline %q has no manifest identifier", p.Name)
		}

		nodes[i] = manifestNode{
			id:        manifestID,
			groupID:   p.Config.DependencyID(),
			name:      p.Name,
			dependsOn: uniqueManifestDependencies(p.Config.Spec.DependsOn),
			index:     i,
		}

		if nodes[i].groupID != "" {
			indicesByGroupID[nodes[i].groupID] = append(indicesByGroupID[nodes[i].groupID], i)
		}
	}

	indegree := make([]int, len(nodes))
	successors := make([][]int, len(nodes))

	for i, node := range nodes {
		for _, dependencyID := range node.dependsOn {
			dependencyIndexes, ok := indicesByGroupID[dependencyID]
			if !ok || len(dependencyIndexes) == 0 {
				return fmt.Errorf("manifest %s depends on unknown manifest id %q",
					describeManifestNode(node), dependencyID)
			}

			for _, dependencyIndex := range dependencyIndexes {
				if dependencyIndex == i {
					return fmt.Errorf("manifest %s cannot depend on its own id %q",
						describeManifestNode(node), dependencyID)
				}

				successors[dependencyIndex] = append(successors[dependencyIndex], i)
				indegree[i]++
			}
		}
	}

	ready := make([]int, 0, len(nodes))
	for i := range nodes {
		if indegree[i] == 0 {
			ready = append(ready, i)
		}
	}

	orderedIndices := make([]int, 0, len(nodes))
	for len(ready) > 0 {
		currentBatch := append([]int{}, ready...)
		ready = ready[:0]

		for _, current := range currentBatch {
			orderedIndices = append(orderedIndices, current)

			for _, successor := range successors[current] {
				indegree[successor]--
				if indegree[successor] == 0 {
					ready = append(ready, successor)
				}
			}
		}

		slices.Sort(ready)
	}

	if len(orderedIndices) != len(nodes) {
		cycle := []string{}
		for i, degree := range indegree {
			if degree > 0 {
				cycle = append(cycle, describeManifestNode(nodes[i]))
			}
		}

		return fmt.Errorf("manifest dependency cycle detected involving %s", strings.Join(cycle, ", "))
	}

	orderedPipelines := make([]*pipeline.Pipeline, 0, len(orderedIndices))
	orderedConfigurations := make([]*config.Config, 0, len(orderedIndices))

	for _, index := range orderedIndices {
		orderedPipelines = append(orderedPipelines, e.Pipelines[index])
		orderedConfigurations = append(orderedConfigurations, e.Pipelines[index].Config)
	}

	e.Pipelines = orderedPipelines
	e.configurations = orderedConfigurations

	return nil
}

func describeManifestNode(node manifestNode) string {
	manifestID := node.groupID
	if manifestID == "" {
		manifestID = node.id
	}

	if node.name == "" {
		return fmt.Sprintf("manifest id %q", manifestID)
	}

	return fmt.Sprintf("manifest %q (id %q)", node.name, manifestID)
}

func describePipeline(p *pipeline.Pipeline, index int) string {
	if p == nil || p.Config == nil {
		return fmt.Sprintf("pipeline at index %d", index)
	}

	return describeManifestNode(manifestNode{
		id:   p.Config.ManifestID(),
		name: p.Name,
	})
}

func uniqueManifestDependencies(dependencies []string) []string {
	result := []string{}
	seen := map[string]struct{}{}

	for _, dependency := range dependencies {
		if _, ok := seen[dependency]; ok {
			continue
		}

		seen[dependency] = struct{}{}
		result = append(result, dependency)
	}

	return result
}

func mergeManifestDependencies(baseDependencies, inheritedDependencies []string) []string {
	result := append([]string{}, inheritedDependencies...)
	result = append(result, baseDependencies...)

	return uniqueManifestDependencies(result)
}
