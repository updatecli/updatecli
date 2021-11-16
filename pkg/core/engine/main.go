package engine

import (
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

			if s.Scm != nil {
				err = Clone(&s.Scm, channel, &hashes, &wg)
				if err != nil {
					return err
				}
			}
		}
	}
	logrus.Infof("Repository retrieved: %d\n", len(hashes))

	return err
}

// Clone parses a scm configuration then clone the git repository if needed.
func Clone(
	s *scm.ScmHandler,
	channel chan int,
	hashes *[]uint64,
	wg *sync.WaitGroup) error {

	hash, err := hashstructure.Hash(s, nil)
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
		}(*s)
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
	if err != nil {
		return err
	}

	err = e.InitSCM()
	if err != nil {
		return err
	}

	return err
}

// ReadConfigurations read every strategies configuration.
func (e *Engine) LoadConfigurations() error {
	// Read every strategy files
	for _, cfgFile := range GetFiles(e.Options.File) {

		loadedConfiguration, err := config.New(cfgFile, e.Options.ValuesFiles, e.Options.SecretsFiles)

		if err != nil && err != config.ErrConfigFileTypeNotSupported {
			logrus.Errorf("%s\n\n", err)
			continue
		} else if err == config.ErrConfigFileTypeNotSupported {
			continue
		}

		newPipeline := pipeline.Pipeline{}
		newPipeline.Init(
			&loadedConfiguration,
			e.Options.Pipeline)

		e.Pipelines = append(e.Pipelines, newPipeline)
		e.configurations = append(e.configurations, loadedConfiguration)
	}

	return nil

}

// Run run the full process one yaml file.
func (e *Engine) Run() (err error) {
	logrus.Infof("\n\n%s\n", strings.Repeat("+", len("Pipeline")+4))
	logrus.Infof("+ %s +\n", strings.ToTitle("Pipeline"))
	logrus.Infof("%s\n\n", strings.Repeat("+", len("Pipeline")+4))

	for _, pipeline := range e.Pipelines {

		err := pipeline.Run()

		e.Reports = append(e.Reports, pipeline.Report)

		if err != nil {
			logrus.Printf("Pipeline %q failed\n", pipeline.Title)
			logrus.Printf("Skipping due to:\n\t%q\n", err)
			logrus.Println(err)
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

	for _, pipeline := range e.Pipelines {

		logrus.Infof("\n\n%s\n", strings.Repeat("#", len(pipeline.Config.Name)+4))
		logrus.Infof("# %s #\n", strings.ToTitle(pipeline.Config.Name))
		logrus.Infof("%s\n\n", strings.Repeat("#", len(pipeline.Config.Name)+4))

		err = pipeline.Config.Display()
		if err != nil {
			return err
		}

	}
	return nil
}
