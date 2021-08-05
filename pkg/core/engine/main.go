package engine

import (
	"os"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/mitchellh/hashstructure"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/context"
	"github.com/updatecli/updatecli/pkg/core/reports"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/scm"
	"github.com/updatecli/updatecli/pkg/core/tmp"

	"path/filepath"
	"strings"
)

// Engine defined parameters for a specific engine run.
type Engine struct {
	configurations []config.Config
	Contexts       []context.Context
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
			logrus.Errorf("\n\u26A0 File %s: %s\n", path, err)
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

	for _, conf := range e.configurations {
		for _, source := range conf.Sources {

			if len(source.Scm) > 0 {
				err = Clone(&source.Scm, &hashes, channel, &wg)
				if err != nil {
					return err
				}
			}
		}
		for _, condition := range conf.Conditions {
			if len(condition.Scm) > 0 {

				err = Clone(&condition.Scm, &hashes, channel, &wg)
				if err != nil {
					return err
				}

			}
		}

		for _, target := range conf.Targets {
			if len(target.Scm) > 0 {

				err = Clone(&target.Scm, &hashes, channel, &wg)
				if err != nil {
					return err
				}
			}
		}
	}

	return err
}

// Clone parses a scm configuration then clone the git repository if needed.
func Clone(
	SCM *map[string]interface{},
	hashes *[]uint64,
	channel chan int,
	wg *sync.WaitGroup) error {

	hash, err := hashstructure.Hash(SCM, nil)
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
		s, _, err := scm.Unmarshal(*SCM)
		if err != nil {
			logrus.Errorf("err - %s", err)
		}
		*hashes = append(*hashes, hash)
		wg.Add(1)
		go func(s scm.Scm) {
			channel <- 1
			defer wg.Done()
			_, err := s.Clone()
			if err != nil {
				logrus.Errorf("err - %s", err)
			}
		}(s)
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

	err = e.ReadConfigurations()
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
func (e *Engine) ReadConfigurations() error {
	// Read every strategy files
	for _, cfgFile := range GetFiles(e.Options.File) {

		c, err := config.New(cfgFile, e.Options.ValuesFiles, e.Options.SecretsFiles)

		if err != nil && err != config.ErrConfigFileTypeNotSupported {
			logrus.Errorf("%s\n\n", err)
			continue
		} else if err == config.ErrConfigFileTypeNotSupported {
			continue
		}

		e.configurations = append(e.configurations, c)
	}
	return nil

}

// Run run the full process one yaml file.
func (e *Engine) Run() (err error) {
	logrus.Infof("\n\n%s\n", strings.Repeat("+", len("Run")+4))
	logrus.Infof("+ %s +\n", strings.ToTitle("Run"))
	logrus.Infof("%s\n\n", strings.Repeat("+", len("Run")+4))

	for id, conf := range e.configurations {

		currentContext := context.Context{}
		currentContext.Init(&e.configurations[id])

		currentReport := reports.Report{}
		currentReport.Init(
			e.configurations[id].Name,
			len(e.configurations[id].Sources),
			len(e.configurations[id].Conditions),
			len(e.configurations[id].Targets))

		if len(conf.Title) > 0 {
			logrus.Infof("\n\n%s\n", strings.Repeat("#", len(conf.Title)+4))
			logrus.Infof("# %s #\n", strings.ToTitle(conf.Title))
			logrus.Infof("%s\n\n", strings.Repeat("#", len(conf.Title)+4))

		} else {
			logrus.Infof("\n\n%s\n", strings.Repeat("#", len(conf.Name)+4))
			logrus.Infof("# %s #\n", strings.ToTitle(conf.Name))
			logrus.Infof("%s\n\n", strings.Repeat("#", len(conf.Name)+4))
		}

		err = RunSources(
			&e.configurations[id],
			&currentReport,
			&currentContext)

		if err != nil {
			logrus.Errorf("Error occurred while running sources - %q", err.Error())
			e.Reports = append(e.Reports, currentReport)
			e.Contexts = append(e.Contexts, currentContext)
			continue
		}

		if len(conf.Conditions) > 0 {

			ok, err := RunConditions(
				&e.configurations[id],
				&currentContext,
				&currentReport)

			if err != nil {
				logrus.Infof("\n%s error happened during condition evaluation\n\n", result.FAILURE)
				e.Reports = append(e.Reports, currentReport)
				e.Contexts = append(e.Contexts, currentContext)
				continue
			} else if !ok {
				logrus.Infof("\n%s condition not met, skipping pipeline\n", result.FAILURE)
				e.Reports = append(e.Reports, currentReport)
				e.Contexts = append(e.Contexts, currentContext)
				continue
			}

		}

		if len(conf.Targets) > 0 {
			err := RunTargets(
				&e.configurations[id],
				&e.Options.Target,
				&currentReport,
				&currentContext)

			if err != nil {
				logrus.Errorf("%s %v\n", result.FAILURE, err)
				e.Reports = append(e.Reports, currentReport)
				e.Contexts = append(e.Contexts, currentContext)
				continue
			}
		}

		if err != nil {
			logrus.Errorf("\n%s %s \n\n", result.FAILURE, err)
		}

		e.Reports = append(e.Reports, currentReport)
		e.Contexts = append(e.Contexts, currentContext)
	}

	err = e.Reports.Show()
	if err != nil {
		return err
	}
	successCounter, changedCounter, failedCounter, err := e.Reports.Summary()
	if err != nil {
		return err
	}

	logrus.Infof("Run Summary")
	logrus.Infof("===========")
	logrus.Infof("%d job run", successCounter+changedCounter+failedCounter)
	logrus.Infof("%d job succeed", successCounter)
	logrus.Infof("%d job failed", failedCounter)
	logrus.Infof("%d job applied changes", changedCounter)

	logrus.Infof("")

	return err
}

// Show displays configurations that should be apply.
func (e *Engine) Show() error {

	err := e.ReadConfigurations()

	if err != nil {
		return err
	}

	for _, conf := range e.configurations {

		logrus.Infof("\n\n%s\n", strings.Repeat("#", len(conf.Name)+4))
		logrus.Infof("# %s #\n", strings.ToTitle(conf.Name))
		logrus.Infof("%s\n\n", strings.Repeat("#", len(conf.Name)+4))

		err = conf.Display()
		if err != nil {
			return err
		}

	}
	return nil
}
