package engine

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/cmdoptions"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/reports"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github"
	"github.com/updatecli/updatecli/pkg/plugins/scms/githubsearch"
	"github.com/updatecli/updatecli/pkg/plugins/scms/gitlab"
	"github.com/updatecli/updatecli/pkg/plugins/scms/gitlabsearch"
)

// ReadConfigurations read every strategies configuration.
//
//nolint:funlen
func (e *Engine) LoadConfigurations() error {
	// Read every strategy files
	errs := []error{}

	ErrNoManifestDetectedCounter := 0

	for i := range e.Options.Manifests {
		// If no manifest file is specified, we try to detect one
		if len(e.Options.Manifests[i].Manifests) == 0 {
			// Updatecli tries to load the file updatecli.yaml if no manifest was specified
			// If updatecli.yaml doesn't exists then Updatecli parses the directory updatecli.d for any manifests.
			// if there is no manifests in the directory updatecli.d then Updatecli returns no manifest files.

			// defaultManifestFilenames defines the default updatecli configuration filenames
			defaultManifestFilenames := []string{"updatecli.yaml"}
			if cmdoptions.Experimental {
				defaultManifestFilenames = append(defaultManifestFilenames, "updatecli.cue")
			}
			// defaultManifestDirname defines the default updatecli manifest directory
			defaultManifestDirname := "updatecli.d"

			// If no manifest file is specified, we try to detect one
			for _, filename := range defaultManifestFilenames {
				if _, err := os.Stat(filename); err == nil {
					logrus.Debugf("Default Updatecli manifest detected %q", filename)
					e.Options.Manifests[i].Manifests = append(e.Options.Manifests[i].Manifests, filename)
				}
			}

			if fs, err := os.Stat(defaultManifestDirname); err == nil {
				if fs.IsDir() {
					logrus.Debugf("Default Updatecli manifest directory detected %q", defaultManifestDirname)
					e.Options.Manifests[i].Manifests = append(e.Options.Manifests[i].Manifests, defaultManifestDirname)
				}
			}

			if len(e.Options.Manifests[i].Manifests) == 0 {
				ErrNoManifestDetectedCounter++
				continue
			}
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
					DisableTemplating: e.Options.Config.DisableTemplating,
				},
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
				logrus.Debugf("No valid manifest detected for %q", manifestFile)
				continue
			}

			// Load special scm configuration such as githubsearch that can generate multiple scm configurations
			// the generated scm configured must be ready before Updatecli start doing any operation such as
			// clone git repositories, using the autoddiscovery to detect potienial updates.
			// for id, loadedConfiguration := range loadedConfigurations {
			for i := 0; i < len(loadedConfigurations); i++ {

				loadedConfiguration := loadedConfigurations[i]

				for scmID, scmConfig := range loadedConfigurations[i].Spec.SCMs {
					switch scmConfig.Kind {
					case githubsearch.Kind:

						logrus.Debugf("Processing githubsearch scm %q for potential multiple repository discovery", scmID)

						ctx := context.Background()
						autodiscoveryScms, err := githubsearch.New(scmConfig.Spec)
						if err != nil {
							return fmt.Errorf("unable to instantiate githubsearch scm %q: %w", scmID, err)
						}
						discoveredSCms, err := autodiscoveryScms.ScmsGenerator(ctx)
						if err != nil {
							return fmt.Errorf("unable to generate scm specs for githubsearch scm %q: %w", scmID, err)
						}

						if len(discoveredSCms) == 0 {
							// We need to trigger an error if the github search didn't discovered SCMs
							// Otherwise Updatecli will not know how to handle this kind of scm later.
							return fmt.Errorf("no scm discovered for githubsearch scm %q", scmID)
						}

						scmConfig.Kind = github.Kind
						scmConfig.Spec = discoveredSCms[0]

						loadedConfigurations[i].Spec.SCMs[scmID] = scmConfig

						for _, spec := range discoveredSCms[1:] {
							newPipeline := loadedConfiguration

							newPipeline.Spec.SCMs = make(map[string]scm.Config, len(loadedConfiguration.Spec.SCMs))
							maps.Copy(newPipeline.Spec.SCMs, loadedConfiguration.Spec.SCMs)

							newSCM := newPipeline.Spec.SCMs[scmID]
							newSCM.Kind = github.Kind
							newSCM.Spec = spec

							newPipeline.Spec.SCMs[scmID] = newSCM

							loadedConfigurations = append(loadedConfigurations, newPipeline)
							logrus.Debugf("githubsearch scm %q added new pipeline configuration for repository %s/%s", scmID, spec.Owner, spec.Repository)
						}

					case gitlabsearch.Kind:

						logrus.Debugf("Processing gitlabsearch scm %q for potential multiple repository discovery", scmID)

						ctx := context.Background()
						glSearchScms, err := gitlabsearch.New(scmConfig.Spec)
						if err != nil {
							return fmt.Errorf("unable to instantiate gitlabsearch scm %q: %w", scmID, err)
						}
						discoveredGitLabSCms, err := glSearchScms.ScmsGenerator(ctx)
						if err != nil {
							return fmt.Errorf("unable to generate scm specs for gitlabsearch scm %q: %w", scmID, err)
						}

						if len(discoveredGitLabSCms) == 0 {
							// We need to trigger an error if the gitlab search did not discover any SCMs.
							// Otherwise Updatecli will not know how to handle this kind of scm later.
							return fmt.Errorf("no scm discovered for gitlabsearch scm %q", scmID)
						}

						scmConfig.Kind = gitlab.Kind
						scmConfig.Spec = discoveredGitLabSCms[0]

						loadedConfigurations[i].Spec.SCMs[scmID] = scmConfig

						for _, spec := range discoveredGitLabSCms[1:] {
							newPipeline := loadedConfiguration

							newPipeline.Spec.SCMs = make(map[string]scm.Config, len(loadedConfiguration.Spec.SCMs))
							maps.Copy(newPipeline.Spec.SCMs, loadedConfiguration.Spec.SCMs)

							newSCM := newPipeline.Spec.SCMs[scmID]
							newSCM.Kind = gitlab.Kind
							newSCM.Spec = spec

							newPipeline.Spec.SCMs[scmID] = newSCM

							loadedConfigurations = append(loadedConfigurations, newPipeline)
							logrus.Debugf("gitlabsearch scm %q added new pipeline configuration for repository %s/%s", scmID, spec.Owner, spec.Repository)
						}
					}
				}
			}

			logrus.Debugf("Loaded %d pipeline configuration(s) from %q", len(loadedConfigurations), manifestFile)

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
