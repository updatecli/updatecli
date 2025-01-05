package engine

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline"
	"github.com/updatecli/updatecli/pkg/core/reports"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// ReadConfigurations read every strategies configuration.
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

			// defaultManifestFilename defines the default updatecli configuration filename
			defaultManifestFilename := "updatecli.yaml"
			// defaultManifestDirname defines the default updatecli manifest directory
			defaultManifestDirname := "updatecli.d"

			// If no manifest file is specified, we try to detect one
			if _, err := os.Stat(defaultManifestFilename); err == nil {
				logrus.Debugf("Default Updatecli manifest detected %q", defaultManifestFilename)
				e.Options.Manifests[i].Manifests = append(e.Options.Manifests[i].Manifests, defaultManifestFilename)
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

		for _, manifestFile := range sanitizeUpdatecliManifestFilePath(e.Options.Manifests[i].Manifests) {

			loadedConfigurations, err := config.New(
				config.Option{
					ManifestFile:      manifestFile,
					SecretsFiles:      e.Options.Manifests[i].Secrets,
					ValuesFiles:       e.Options.Manifests[i].Values,
					DisableTemplating: e.Options.Config.DisableTemplating,
				})

			switch err {
			case config.ErrConfigFileTypeNotSupported:
				// Updatecli accepts either a single configuration file or a directory containing multiple configurations.
				// When browsing files from a directory, Updatecli ignores unsupported files.
				continue
			case nil:
				// nothing to do
			default:
				err = fmt.Errorf("%q - %s", manifestFile, err)
				errs = append(errs, err)
				e.Reports = append(e.Reports,
					reports.Report{
						Result: result.FAILURE,
						Err:    err.Error(),
					},
				)
				continue
			}

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
					// don't initially fail as init. of the pipeline still fails even with a successful validation
					err := fmt.Errorf("%q - %s", manifestFile, err)
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
			e = fmt.Errorf("%s\n\t* %s", e.Error(), strings.ReplaceAll(err.Error(), "\n", "\n\t\t* "))
			if errors.Is(err, ErrNoManifestDetected) {
				return err
			}
		}
		return e
	}

	return nil
}
