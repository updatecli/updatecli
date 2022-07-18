package engine

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/mitchellh/hashstructure"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/jsonschema"
	"github.com/updatecli/updatecli/pkg/core/pipeline"
	"github.com/updatecli/updatecli/pkg/core/pipeline/autodiscovery"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/reports"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/tmp"

	"path/filepath"
	"strings"
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
func GetFiles(root string) (files []string) {
	if root == "" {
		// If no manifest have been provided then we try to see if the file
		// updatecli.yaml exist. If it's then we try to see if the directory updatecli.d
		// if it's still not the case then we return no manifest files.
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
			root = config.DefaultConfigDirname
		}
	}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
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

	return files
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

	err = tmp.Create()
	if err != nil {
		return err
	}

	err = e.LoadConfigurations()

	// Don't exit if we identify at least one valid pipeline configuration
	if err != nil {
		logrus.Errorln(err)
		logrus.Infof("\n%d pipeline(s) successfully loaded\n", len(e.Pipelines))
	}

	// If one git clone fails then we exit
	err = e.InitSCM()

	if err != nil {
		return err
	}

	if !e.Options.Pipeline.AutoDiscovery.Disabled {
		err = e.LoadAutoDiscovery()
		if err != nil {
			return err
		}
	}

	if len(e.Pipelines) == 0 {
		logrus.Errorln(err)
		return fmt.Errorf("no valid pipeline found")
	}

	return nil
}

// ReadConfigurations read every strategies configuration.
func (e *Engine) LoadConfigurations() error {
	// Read every strategy files
	errs := []error{}

	for _, manifestFile := range GetFiles(e.Options.Config.ManifestFile) {

		loadedConfiguration, err := config.New(
			config.Option{
				ManifestFile:      manifestFile,
				SecretsFiles:      e.Options.Config.SecretsFiles,
				ValuesFiles:       e.Options.Config.ValuesFiles,
				DisableTemplating: e.Options.Config.DisableTemplating,
			})

		switch err {
		case config.ErrConfigFileTypeNotSupported:
			// Updatecli accepts either a single configuration file or a directory containing multiple configurations.
			// When browsing files from a directory, we don't want to record error due to unsupported files.
			continue
		case nil:
			// nothing to do
		default:
			err = fmt.Errorf("%q - %s", manifestFile, err)
			errs = append(errs, err)
			continue
		}

		newPipeline := pipeline.Pipeline{}
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

	if len(errs) > 0 {

		e := errors.New("failed loading pipeline(s)")

		for _, err := range errs {
			e = fmt.Errorf("%s\n\t* %s", e.Error(), err)
		}
		return e
	}

	return nil

}

// Run runs the full process for one yaml file.
func (e *Engine) Run() (err error) {
	logrus.Infof("\n\n%s\n", strings.Repeat("+", len("Pipeline")+4))
	logrus.Infof("+ %s +\n", strings.ToTitle("Pipeline"))
	logrus.Infof("%s\n\n", strings.Repeat("+", len("Pipeline")+4))

	for _, pipeline := range e.Pipelines {

		err := pipeline.Run()

		e.Reports = append(e.Reports, pipeline.Report)

		if err != nil {
			logrus.Printf("Pipeline %q failed\n", pipeline.Title)
			logrus.Printf("Skipping due to:\n\t%s\n", err)
			continue
		}
	}

	err = e.Reports.Show()
	if err != nil {
		return err
	}
	totalSuccessPipeline, totalChangedAppliedPipeline, totalFailedPipeline, totalSkippedPipeline := e.Reports.Summary()

	totalPipeline := totalSuccessPipeline + totalChangedAppliedPipeline + totalFailedPipeline + totalSkippedPipeline

	logrus.Infof("Run Summary")
	logrus.Infof("===========\n")
	logrus.Infof("Pipeline(s) run:")
	logrus.Infof("  * Changed:\t%d", totalChangedAppliedPipeline)
	logrus.Infof("  * Failed:\t%d", totalFailedPipeline)
	logrus.Infof("  * Skipped:\t%d", totalSkippedPipeline)
	logrus.Infof("  * Succeeded:\t%d", totalSuccessPipeline)
	logrus.Infof("  * Total:\t%d", totalPipeline)

	// Exit on error if at least one pipeline failed
	if totalFailedPipeline > 0 {
		return fmt.Errorf("%d over %d pipeline failed", totalFailedPipeline, totalPipeline)
	}

	return err
}

// Show displays configurations that should be apply.
func (e *Engine) Show() error {

	err := e.LoadConfigurations()

	if err != nil {
		return err
	}

	if !e.Options.Pipeline.AutoDiscovery.Disabled {
		err = e.LoadAutoDiscovery()
		if err != nil {
			return err
		}
	}

	for _, pipeline := range e.Pipelines {

		logrus.Infof("\n\n%s\n", strings.Repeat("#", len(pipeline.Config.Spec.Name)+4))
		logrus.Infof("# %s #\n", strings.ToTitle(pipeline.Config.Spec.Name))
		logrus.Infof("%s\n\n", strings.Repeat("#", len(pipeline.Config.Spec.Name)+4))

		err = pipeline.Config.Display()
		if err != nil {
			return err
		}

	}
	return nil
}

func GenerateSchema(baseSchemaID, schemaDir string) error {

	logrus.Infof("\n\n%s\n", strings.Repeat("+", len("Json Schema")+4))
	logrus.Infof("+ %s +\n", strings.ToTitle("Json Schema"))
	logrus.Infof("%s\n\n", strings.Repeat("+", len("Json Schema")+4))

	err := jsonschema.CloneCommentDirectory()

	if err != nil {
		return err
	}

	defer func() {
		tmperr := jsonschema.CleanCommentDirectory()
		if err != nil {
			err = fmt.Errorf("%s\n%s", err, tmperr)
		}
	}()

	s := jsonschema.New(baseSchemaID, schemaDir)
	err = s.GenerateSchema(&config.Config{})
	if err != nil {
		return err
	}

	logrus.Infof("```\n%s\n```\n", s)

	err = s.Save()
	if err != nil {
		return err
	}

	return s.GenerateSchema(&config.Config{})
}

// LoadAutoDiscovery will try to guess available pipelines based on specific directory
func (e *Engine) LoadAutoDiscovery() error {

	logrus.Infof("\n\n%s\n", strings.Repeat("+", len("Auto Discovery")+4))
	logrus.Infof("+ %s +\n", strings.ToTitle("Auto Discovery"))
	logrus.Infof("%s\n\n", strings.Repeat("+", len("Auto Discovery")+4))

	var autoDiscoveryPipelines []pipeline.Pipeline

	for _, p := range e.Pipelines {
		if p.Config.Spec.AutoDiscovery.Crawlers != nil {
			autoDiscoveryPipelines = append(autoDiscoveryPipelines, p)
		}
	}

	// At least run once
	// Failing
	//autoDiscoveryPipelines = append(autoDiscoveryPipelines, pipeline.Pipeline{})

	for _, p := range autoDiscoveryPipelines {

		var sc scm.Config
		var autodiscoveryScm scm.Scm
		var found bool

		if len(p.Config.Spec.AutoDiscovery.ScmId) > 0 {
			autodiscoveryScm, found = p.SCMs[p.Config.Spec.AutoDiscovery.ScmId]

			if found {
				sc = *autodiscoveryScm.Config
			}
		}

		c, err := autodiscovery.New(
			p.Config.Spec.AutoDiscovery,
			autodiscoveryScm.Handler,
			&sc)

		if err != nil {
			logrus.Errorln(err)
			return err
		}

		errs := []error{}

		manifests, err := c.Run()

		if err != nil {
			logrus.Errorln(err)
			return err
		}

		if len(manifests) == 0 {
			logrus.Infof("nothing detected")
		}

		for i := range manifests {
			logrus.Infof("%v. %s", i, manifests[i].Name)

			newConfig := config.Config{
				Spec: manifests[i],
			}

			newPipeline := pipeline.Pipeline{}
			err = newPipeline.Init(&newConfig, e.Options.Pipeline)

			if err == nil {
				e.Pipelines = append(e.Pipelines, newPipeline)
				e.configurations = append(e.configurations, newConfig)
			} else {
				// don't initially fail as init. of the pipeline still fails even with a successful validation
				err := fmt.Errorf("%q - %s", manifests[i].Name, err)
				errs = append(errs, err)
			}
			if len(errs) > 0 {
				logrus.Errorf("Error(s) happened while generating Updatecli pipeline manifest")
				for i := range errs {
					logrus.Errorf("%v", errs[i])
				}
			}
		}

	}

	return nil

}
