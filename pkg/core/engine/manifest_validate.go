package engine

import (
	"errors"
	"fmt"
	"sort"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// ValidateManifests checks every manifest against the Updatecli schema and reports what
// does not match, such as a misspelled key that would otherwise be silently ignored.
//
// It returns an error when at least one problem is worth failing on, so that the command
// can gate a pipeline.
func (e *Engine) ValidateManifests(strict bool) error {

	PrintTitle("Manifest Validation")

	problems := []config.SchemaProblem{}
	manifestCount := 0

	collect := func(problem config.SchemaProblem) {
		problems = append(problems, problem)
	}

	for i := range e.Options.Manifests {
		if !e.detectManifests(i) {
			continue
		}

		manifestFiles, manifestPartials := sanitizeUpdatecliManifestFilePath(e.Options.Manifests[i].Manifests)

		for _, manifestFile := range manifestFiles {
			// Reading the manifest goes through the very same reading, templating and
			// parsing steps as a regular run, so that what is validated is what
			// Updatecli actually sees.
			_, err := config.New(
				config.Option{
					PartialFiles:      manifestPartials,
					ManifestFile:      manifestFile,
					SecretsFiles:      e.Options.Manifests[i].Secrets,
					ValuesFiles:       e.Options.Manifests[i].Values,
					ValuesInline:      e.Options.Manifests[i].ValuesInline,
					DisableTemplating: e.Options.Config.DisableTemplating,
					ValidateSchema:    true,
					OnSchemaProblem:   collect,
				},
				e.Options.ManifestOptions,
				e.Options.PipelineIDs,
				e.Options.Labels,
			)

			if errors.Is(err, config.ErrConfigFileTypeNotSupported) {
				// Updatecli accepts either a single manifest or a directory holding
				// several, in which case unsupported files are ignored.
				continue
			}

			manifestCount++

			// A manifest Updatecli cannot load at all is reported as a problem of its
			// own, as the schema could not be checked against it.
			if err != nil {
				problems = append(problems, config.SchemaProblem{
					File:     manifestFile,
					Severity: config.SeverityError,
					Message:  err.Error(),
				})
			}
		}
	}

	if manifestCount == 0 {
		return ErrNoManifestDetected
	}

	sort.SliceStable(problems, func(i, j int) bool {
		if problems[i].File != problems[j].File {
			return problems[i].File < problems[j].File
		}
		return problems[i].Path < problems[j].Path
	})

	errorCount := 0

	for _, problem := range problems {
		if problem.Severity == config.SeverityError || strict {
			errorCount++
			logrus.Errorf("%s", problem.String())
			continue
		}

		logrus.Warningf("%s", problem.String())
	}

	switch {
	case errorCount > 0:
		return fmt.Errorf("%d manifest problem(s) detected in %d manifest(s)", errorCount, manifestCount)
	case len(problems) > 0:
		logrus.Infof("\n%s %d warning(s) detected in %d manifest(s)",
			result.ATTENTION, len(problems), manifestCount)
	default:
		logrus.Infof("\n%s %d manifest(s) successfully validated", result.SUCCESS, manifestCount)
	}

	return nil
}
