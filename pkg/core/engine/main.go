package engine

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/mitchellh/hashstructure"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/reports"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/tmp"

	"path/filepath"
	"strings"
)

var (
	// ErrNoManifestDetected is the error message returned by Updatecli if it can't find manifest
	ErrNoManifestDetected error = errors.New("no Updatecli manifest detected")
)

// Engine defined parameters for a specific engine run.
type Engine struct {
	configurations []config.Config
	Pipelines      []pipeline.Pipeline
	Options        Options
	Reports        reports.Reports
}

// Clean remove every traces from an updatecli run.
func (e *Engine) Clean() (err error) {
	err = tmp.Clean()
	return
}

// GetFiles return an array with every valid files.
func GetFiles(root []string) (files []string) {
	if len(root) == 0 {
		// Updatecli tries to load the file updatecli.yaml if no manifest provided
		// If updatecli.yaml doesn't exists then Updatecli parses the directory updatecli.d for any manifests.
		// if there is no manifests in the directory updatecli.d then Updatecli returns no manifest files.
		_, err := os.Stat(config.DefaultConfigFilename)
		if !errors.Is(err, os.ErrNotExist) {
			logrus.Debugf("Default Updatecli manifest detected %q", config.DefaultConfigFilename)
			return []string{config.DefaultConfigFilename}
		}

		fs, err := os.Stat(config.DefaultConfigDirname)
		if errors.Is(err, os.ErrNotExist) {
			return []string{}
		}

		if fs.IsDir() {
			logrus.Debugf("Default Updatecli manifest directory detected %q", config.DefaultConfigDirname)
			root = []string{config.DefaultConfigDirname}
		}
	}

	for _, r := range root {
		err := filepath.Walk(r, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				logrus.Errorf("\n%s File %s: %s\n", result.FAILURE, path, err)
				os.Exit(1)
			}
			if info.Mode().IsRegular() {
				files = append(files, path)
			}
			return nil
		})

		if err != nil {
			logrus.Errorf("err - %s", err)
		}
	}

	// Remove duplicates manifest files
	result := []string{}
	exist := map[string]bool{}

	for v := range files {
		if !exist[files[v]] {
			exist[files[v]] = true
			result = append(result, files[v])
		}
	}

	return result
}

// InitSCM search and clone only once SCM configurations found.
func (e *Engine) InitSCM() (err error) {
	hashes := []uint64{}

	wg := sync.WaitGroup{}
	channel := make(chan int, 20)
	defer wg.Wait()

	for _, pipeline := range e.Pipelines {
		for _, s := range pipeline.SCMs {

			if s.Handler != nil {
				err = Clone(&s.Handler, channel, &hashes, &wg)
				if err != nil {
					return err
				}
			}
		}
	}
	logrus.Infof("\nSCM repository retrieved: %d", len(hashes))

	return err
}

// Clone parses a scm configuration then clone the git repository if needed.
func Clone(
	s *scm.ScmHandler,
	channel chan int,
	hashes *[]uint64,
	wg *sync.WaitGroup) error {

	scmhandler := *s

	hash, err := hashstructure.Hash(scmhandler.GetDirectory(), nil)
	if err != nil {
		return err
	}
	found := false

	for _, h := range *hashes {
		if h == hash {
			found = true
		}
	}

	if !found {
		*hashes = append(*hashes, hash)
		wg.Add(1)
		go func(s scm.ScmHandler) {
			channel <- 1
			defer wg.Done()
			_, err := s.Clone()
			if err != nil {
				logrus.Errorf("err - %s", err)
			}
		}(scmhandler)
		<-channel

	}

	return nil
}

// Prepare run every actions needed before going further.
func (e *Engine) Prepare() (err error) {
	logrus.Infof("\n\n%s\n", strings.Repeat("+", len("Prepare")+4))
	logrus.Infof("+ %s +\n", strings.ToTitle("Prepare"))
	logrus.Infof("%s\n\n", strings.Repeat("+", len("Prepare")+4))

	var defaultCrawlersEnabled bool

	err = tmp.Create()
	if err != nil {
		return err
	}

	err = e.LoadConfigurations()
	if !errors.Is(err, ErrNoManifestDetected) && err != nil {
		logrus.Errorln(err)
		logrus.Infof("\n%d pipeline(s) successfully loaded\n", len(e.Pipelines))
	}

	if errors.Is(err, ErrNoManifestDetected) {
		defaultCrawlersEnabled = true
	}

	// If one git clone fails then Updatecli exits
	// scm initialization must be done before autodiscovery as we need to identify
	// in advance git repository directories to analyze them for possible common update scenarii
	err = e.InitSCM()
	if err != nil {
		return err
	}

	err = e.LoadAutoDiscovery(defaultCrawlersEnabled)
	if err != nil {
		return err
	}

	if len(e.Pipelines) == 0 {
		return fmt.Errorf("no valid pipeline found")
	}

	return nil
}

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
