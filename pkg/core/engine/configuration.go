package engine

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"os"
	"strings"

	"github.com/updatecli/updatecli/pkg/core/cmdoptions"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/core/pipeline"
	"github.com/updatecli/updatecli/pkg/core/reports"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/scms/azuredevops"
	"github.com/updatecli/updatecli/pkg/plugins/scms/azuredevopssearch"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github"
	"github.com/updatecli/updatecli/pkg/plugins/scms/githubsearch"
	"github.com/updatecli/updatecli/pkg/plugins/scms/gitlab"
	"github.com/updatecli/updatecli/pkg/plugins/scms/gitlabsearch"
)

// ReadConfigurations read every strategies configuration.
//
// detectManifests fills in the default manifest locations for a manifest source that
// does not name any, and reports whether at least one manifest is available.
//
// Updatecli tries to load the file updatecli.yaml if no manifest was specified.
// If updatecli.yaml doesn't exists then Updatecli parses the directory updatecli.d for
// any manifests. If there is no manifests in the directory updatecli.d then Updatecli
// returns no manifest files.
func (e *Engine) detectManifests(i int) bool {
	if len(e.Options.Manifests[i].Manifests) > 0 {
		return true
	}

	// defaultManifestFilenames defines the default updatecli configuration filenames
	defaultManifestFilenames := []string{"updatecli.yaml"}
	if cmdoptions.Experimental {
		defaultManifestFilenames = append(defaultManifestFilenames, "updatecli.cue")
	}
	// defaultManifestDirname defines the default updatecli manifest directory
	defaultManifestDirname := "updatecli.d"

	for _, filename := range defaultManifestFilenames {
		if _, err := os.Stat(filename); err == nil {
			log.Debugf("Default Updatecli manifest detected %q", filename)
			e.Options.Manifests[i].Manifests = append(e.Options.Manifests[i].Manifests, filename)
		}
	}

	if fs, err := os.Stat(defaultManifestDirname); err == nil {
		if fs.IsDir() {
			log.Debugf("Default Updatecli manifest directory detected %q", defaultManifestDirname)
			e.Options.Manifests[i].Manifests = append(e.Options.Manifests[i].Manifests, defaultManifestDirname)
		}
	}

	return len(e.Options.Manifests[i].Manifests) > 0
}

// LoadConfigurations loads every configuration files and initialize the pipelines
func (e *Engine) LoadConfigurations() error {
	// Checked once here so that a bad command line value is reported once, instead of once
	// for every manifest it is merged into.
	if err := e.Options.ManifestOptions.Validate(); err != nil {
		return fmt.Errorf("invalid manifest options: %w", err)
	}

	// Read every strategy files
	errs := []error{}

	ErrNoManifestDetectedCounter := 0

	for i := range e.Options.Manifests {
		if !e.detectManifests(i) {
			ErrNoManifestDetectedCounter++
			continue
		}

		manifestFiles, manifestPartials := sanitizeUpdatecliManifestFilePath(e.Options.Manifests[i].Manifests)
		for _, manifestFile := range manifestFiles {
			var err error

			formatErr := func() {
				switch len(manifestPartials) {
				case 0:
					err = fmt.Errorf("%s:\n%s", manifestFile, err)
				default:
					err = fmt.Errorf("%s:\n* Partial files:\n\t* %s\n* Error:\n\t%s",
						manifestFile,
						strings.Join(manifestPartials, "\n\t* "),
						strings.ReplaceAll(err.Error(), "\n", "\n\t"),
					)
				}
			}

			loadedConfigurations, err := config.New(
				config.Option{
					PartialFiles:      manifestPartials,
					ManifestFile:      manifestFile,
					SecretsFiles:      e.Options.Manifests[i].Secrets,
					ValuesFiles:       e.Options.Manifests[i].Values,
					ValuesInline:      e.Options.Manifests[i].ValuesInline,
					DisableTemplating: e.Options.Config.DisableTemplating,
					ValidateSchema:    e.Options.Config.ValidateSchema,
				},
				e.Options.ManifestOptions,
				e.Options.PipelineIDs,
				e.Options.Labels,
			)

			switch err {
			case config.ErrConfigFileTypeNotSupported:
				// Updatecli accepts either a single configuration file or a directory containing multiple configurations.
				// When browsing files from a directory, Updatecli ignores unsupported files.
				continue
			case nil:
				// nothing to do
			default:

				formatErr()

				errs = append(errs, err)
				e.Reports = append(e.Reports,
					reports.Report{
						Name:   fmt.Sprintf("Loading manifest %q", manifestFile),
						Result: result.FAILURE,
						Err:    err.Error(),
					},
				)
				continue
			}

			if loadedConfigurations == nil {
				log.Debugf("No valid manifest detected for %q", manifestFile)
				continue
			}

			// Expand search SCMs before initializing pipelines.
			// Appended configurations may contain other searches, so visit them too.
			for j := 0; j < len(loadedConfigurations); j++ {
				for scmID, scmConfig := range loadedConfigurations[j].Spec.SCMs {
					switch scmConfig.Kind {
					case githubsearch.Kind:

						log.Debugf("Processing githubsearch scm %q for potential multiple repository discovery", scmID)

						ctx := context.Background()
						autodiscoveryScms, err := githubsearch.New(scmConfig.Spec)
						if err != nil {
							return fmt.Errorf("unable to instantiate githubsearch scm %q: %w", scmID, err)
						}
						discoveredSCms, err := autodiscoveryScms.ScmsGenerator(ctx)
						if err != nil {
							return fmt.Errorf("unable to generate scm specs for githubsearch scm %q: %w", scmID, err)
						}

						loadedConfigurations, err = expandSCMConfigurations(loadedConfigurations, j, scmID, github.Kind, discoveredSCms)
						if err != nil {
							return err
						}
						for _, spec := range discoveredSCms[1:] {
							log.Debugf("githubsearch scm %q added new pipeline configuration for repository %s/%s", scmID, spec.Owner, spec.Repository)
						}

					case azuredevopssearch.Kind:

						log.Debugf("Processing azuredevopssearch scm %q for potential multiple repository discovery", scmID)

						ctx := context.Background()
						azdoSearchScms, err := azuredevopssearch.New(scmConfig.Spec)
						if err != nil {
							return fmt.Errorf("unable to instantiate azuredevopssearch scm %q: %w", scmID, err)
						}
						discoveredAzureDevOpsSCms, err := azdoSearchScms.ScmsGenerator(ctx)
						if err != nil {
							return fmt.Errorf("unable to generate scm specs for azuredevopssearch scm %q: %w", scmID, err)
						}

						loadedConfigurations, err = expandSCMConfigurations(loadedConfigurations, j, scmID, azuredevops.Kind, discoveredAzureDevOpsSCms)
						if err != nil {
							return err
						}
						for _, spec := range discoveredAzureDevOpsSCms[1:] {
							log.Debugf("azuredevopssearch scm %q added new pipeline configuration for repository %s/%s", scmID, spec.Project, spec.Repository)
						}

					case gitlabsearch.Kind:

						log.Debugf("Processing gitlabsearch scm %q for potential multiple repository discovery", scmID)

						ctx := context.Background()
						glSearchScms, err := gitlabsearch.New(scmConfig.Spec)
						if err != nil {
							return fmt.Errorf("unable to instantiate gitlabsearch scm %q: %w", scmID, err)
						}
						discoveredGitLabSCms, err := glSearchScms.ScmsGenerator(ctx)
						if err != nil {
							return fmt.Errorf("unable to generate scm specs for gitlabsearch scm %q: %w", scmID, err)
						}

						loadedConfigurations, err = expandSCMConfigurations(loadedConfigurations, j, scmID, gitlab.Kind, discoveredGitLabSCms)
						if err != nil {
							return err
						}
						for _, spec := range discoveredGitLabSCms[1:] {
							log.Debugf("gitlabsearch scm %q added new pipeline configuration for repository %s/%s", scmID, spec.Owner, spec.Repository)
						}
					}
				}
			}

			log.Debugf("Loaded %d pipeline configuration(s) from %q", len(loadedConfigurations), manifestFile)

			for id := range loadedConfigurations {
				newPipeline := pipeline.Pipeline{}
				loadedConfiguration := loadedConfigurations[id]

				err = newPipeline.Init(
					&loadedConfiguration,
					e.Options.Pipeline)

				if err == nil {
					e.Pipelines = append(e.Pipelines, &newPipeline)
					e.configurations = append(e.configurations, &loadedConfiguration)
				} else {

					formatErr()

					errs = append(errs, err)
					e.Reports = append(e.Reports,
						reports.Report{
							Result: result.FAILURE,
							Err:    err.Error(),
						},
					)
				}
			}
		}
	}

	if ErrNoManifestDetectedCounter == len(e.Options.Manifests) {
		errs = append(errs, ErrNoManifestDetected)
	}

	if len(errs) > 0 {

		e := errors.New("failed loading pipeline(s)")

		for _, err := range errs {
			e = fmt.Errorf("%s\n\t* %s", e.Error(), strings.ReplaceAll(err.Error(), "\n", "\n\t\t"))
			if errors.Is(err, ErrNoManifestDetected) {
				return err
			}
		}
		return e
	}

	return nil
}

// expandSCMConfigurations replaces a search SCM and copies its pipeline for each additional result.
func expandSCMConfigurations[T any](configurations []config.Config, index int, scmID, kind string, specs []T) ([]config.Config, error) {
	configuration := configurations[index]
	if len(specs) == 0 {
		return configurations, fmt.Errorf("no scm discovered for %s scm %q", configuration.Spec.SCMs[scmID].Kind, scmID)
	}

	for i, spec := range specs {
		newPipeline := configuration
		// Config.Update renders into these maps, so each pipeline needs its own.
		newPipeline.Spec.SCMs = maps.Clone(configuration.Spec.SCMs)
		newPipeline.Spec.Sources = maps.Clone(configuration.Spec.Sources)
		newPipeline.Spec.Conditions = maps.Clone(configuration.Spec.Conditions)
		newPipeline.Spec.Targets = maps.Clone(configuration.Spec.Targets)
		newPipeline.Spec.Actions = maps.Clone(configuration.Spec.Actions)
		newPipeline.Spec.PullRequests = maps.Clone(configuration.Spec.PullRequests)
		newPipeline.Spec.Labels = maps.Clone(configuration.Spec.Labels)
		newPipeline.Spec.AutoDiscovery.Crawlers = maps.Clone(configuration.Spec.AutoDiscovery.Crawlers)
		newSCM := newPipeline.Spec.SCMs[scmID]
		newSCM.Kind = kind
		newSCM.Spec = spec
		newPipeline.Spec.SCMs[scmID] = newSCM

		if i == 0 {
			configurations[index] = newPipeline
		} else {
			configurations = append(configurations, newPipeline)
		}
	}
	return configurations, nil
}
