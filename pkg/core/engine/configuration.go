package engine

import (
	"errors"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline"
)

// ReadConfigurations read every strategies configuration.
func (e *Engine) LoadConfigurations() error {
	// Read every strategy files
	errs := []error{}

	ErrNoManifestDetectedCounter := 0

	for i := range e.Options.Manifests {
		if e.Options.Manifests[i].IsZero() {
			ErrNoManifestDetectedCounter++
			continue
		}

		for _, manifestFile := range e.Options.Manifests[i].Manifests {

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
				continue
			}

			for id := range loadedConfigurations {
				newPipeline := pipeline.Pipeline{}
				loadedConfiguration := loadedConfigurations[id]

				err = newPipeline.Init(
					&loadedConfiguration,
					e.Options.Pipeline)

				if err == nil {
					e.Pipelines = append(e.Pipelines, newPipeline)
					e.configurations = append(e.configurations, loadedConfiguration)
				} else {
					// don't initially fail as init. of the pipeline still fails even with a successful validation
					err := fmt.Errorf("%q - %s", manifestFile, err)
					errs = append(errs, err)
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
			e = fmt.Errorf("%s\n\t* %s", e.Error(), err)
			if errors.Is(err, ErrNoManifestDetected) {
				return err
			}
		}
		return e
	}

	return nil

}