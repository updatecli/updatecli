package engine

import (
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/mitchellh/hashstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// InitSCM search and clone only once SCM configurations found.
func (e *Engine) InitSCM() (err error) {
	hashes := []uint64{}

	wg := sync.WaitGroup{}
	channel := make(chan int, 20)
	defer wg.Wait()

	for i := range e.Pipelines {
		pipeline := e.Pipelines[i]

		for j := range pipeline.SCMs {
			s := pipeline.SCMs[j]

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

// finalizeSCMUpdates push all commits for all pipelines SCMs.
func (e *Engine) finalizeSCMUpdates() error {

	errs := []string{}

	changedSCM := map[string][]string{}

	allScm := map[string]map[string]*scm.ScmHandler{}
	logrus.Infof("\n\n%s\n", strings.ToTitle("SCMs"))
	logrus.Infof("%s\n\n", strings.Repeat("=", len("SCMs")+1))

	countScms := 0
	countPushedScms := 0

	for id := range e.Pipelines {
		pipeline := e.Pipelines[id]
		if len(pipeline.Targets) > 0 {
			for _, target := range pipeline.Targets {

				// Sanity check, skip if no SCM is configured
				if target.Scm == nil {
					continue
				}

				s := *target.Scm
				url := s.GetURL()
				_, branch, _ := s.GetBranches()

				if _, ok := allScm[url]; !ok {
					allScm[url] = map[string]*scm.ScmHandler{}
					if _, okok := allScm[url][branch]; !okok {
						allScm[url][branch] = &s
						countScms++
					}
				}

				if target.Push {
					if _, ok := changedSCM[url]; ok {
						if slices.Contains(changedSCM[url], branch) {
							logrus.Debugf("Changes for target %q already pushed to %q on branch %q, skipping...\n", target.Config.Name, url, branch)
							continue
						}
					}

					err := target.PushCommits()
					if err != nil {
						errs = append(errs, fmt.Sprintf("pushing commits for target %q: %s", target.Config.Name, err.Error()))
						target.Result.Result = result.FAILURE
						logrus.Errorf("pushing commits for target %q:\t%q", target.Config.Name, err.Error())
						continue
					}
					logrus.Debugf("Pushed to URL: %q on branch: %q", url, branch)

					if _, ok := changedSCM[url]; !ok {
						changedSCM[url] = []string{}
						if !slices.Contains(changedSCM[url], branch) {
							changedSCM[url] = append(changedSCM[url], branch)
							countPushedScms++
						}
					}
				}
			}
		}
	}

	if countPushedScms == 0 {
		logrus.Info("No SCM repositories required pushing")
	} else {
		logrus.Infof("Pushed changes to %d of %d SCM repositories", countPushedScms, countScms)
	}

	for url := range allScm {
		for branch := range allScm[url] {

			scmHandlerPtr := allScm[url][branch]
			scmHandler := *scmHandlerPtr

			isRemoteBranchUpToDate, err := scmHandler.IsRemoteBranchUpToDate()
			if err != nil {
				errs = append(errs, fmt.Sprintf("checking remote branch status for %q on branch %q: %s", url, branch, err.Error()))
				continue
			}

			if isRemoteBranchUpToDate {
				logrus.Debugf("No changes to push to %q on branch %q\n", url, branch)
				continue
			}

			logrus.Infof("\n\u26A0 According to the git history, some commits must be pushed to %q\n", scmHandler.Summary())

			isPushed, err := scmHandler.Push()
			if err != nil {
				errs = append(errs, fmt.Sprintf("pushing commits to %q on branch %q: %s", url, branch, err.Error()))
			}
			logrus.Debugf("Pushed changes to %q on branch %q: %t\n", url, branch, isPushed)

			logrus.Debugf("SCM URL: %q on branch: %q\n", url, branch)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf(
			"errors occurred while running actions:\n\t* %s",
			strings.Join(errs, "\n\t* "))
	}

	return nil
}
