package engine

import (
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/engine/manifest"
	"github.com/updatecli/updatecli/pkg/core/pipeline"
)

// Options defines application specific behaviors
type Options struct {
	// Config holds the application configuration options: how manifests are read.
	Config config.Option
	// Pipeline holds pipeline execution options: how manifests are run.
	Pipeline pipeline.Options
	// ManifestOptions holds the defaults for the manifest "options" section, applied to
	// the manifests that do not set them themselves.
	//
	// It is named ManifestOptions rather than Manifest to stay distinguishable from the
	// Manifests field below, which lists the manifest files to process.
	ManifestOptions config.ManifestOptions
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
	// ExportToYAML defines whether to export the pipeline reports to YAML files
	ExportToYAML bool
	// DisableUdashReport defines whether to skip publishing pipeline reports to Udash
	DisableUdashReport bool
}
