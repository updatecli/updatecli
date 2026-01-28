package engine

import (
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/engine/manifest"
	"github.com/updatecli/updatecli/pkg/core/pipeline"
)

// Options defines application specific behaviors
type Options struct {
	// Config holds the application configuration options
	Config config.Option
	// Pipeline holds pipeline execution options
	Pipeline pipeline.Options
	// Manifests holds a list of manifests to process
	Manifests []manifest.Manifest
	// DisplayFlavor defines the flavor of the display output
	DisplayFlavor string
	// GraphFlavor defines the flavor of the dependency graph
	GraphFlavor string
	// PipelineIDs holds a list of pipeline IDs to filter on
	PipelineIDs []string
	// Labels holds a map of labels to filter on
	Labels map[string]string
}
