package reports

import (
	"encoding/xml"
	"fmt"
	"os"
	"sort"

	"github.com/sirupsen/logrus"
)

type Action struct {
	ID            string         `xml:"id,attr"`
	Title         string         `xml:"-"`
	PipelineTitle string         `xml:"h3,omitempty"`
	Description   string         `xml:"p,omitempty"`
	Targets       []ActionTarget `xml:"details,omitempty"`
	// using a pointer to avoid empty tag
	PipelineUrl *PipelineURL `xml:"a,omitempty"`
}

type ActionTargetChangelog struct {
	Title       string `xml:"summary,omitempty"`
	Description string `xml:"pre,omitempty"`
}

type PipelineURL struct {
	URL  string `xml:"href,attr"`
	Name string `xml:",chardata"`
}

// String show an action report formatted as a string
func (a *Action) String() string {
	a.sort()
	output, err := xml.MarshalIndent(a, "", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	return string(output[:])
}

func (a *Action) Merge(sourceAction *Action) {

	var c, d []ActionTarget

	switch len(a.Targets) > len(sourceAction.Targets) {
	case true:
		c = a.Targets
		d = sourceAction.Targets
	case false:
		d = a.Targets
		c = sourceAction.Targets
	}

	for i := range d {
		targetFound := false
		for j := range c {
			if d[i].ID == c[j].ID {
				targetFound = true
				c[j].Merge(&d[i])
				break
			}
		}
		if !targetFound {
			c = append(c, d[i])
		}
	}

	a.Targets = c
	a.sort()
}

func (a *Action) sort() {
	sort.Slice(
		a.Targets,
		func(i, j int) bool {
			return a.Targets[i].ID < a.Targets[j].ID
		})

	for id, target := range a.Targets {
		sort.Slice(
			target.Changelogs,
			func(i, j int) bool {
				return target.Changelogs[i].Title < target.Changelogs[j].Title
			})
		a.Targets[id] = target
	}
}

// ToActionsString show an action report formatted as a string
func (a Action) ToActionsString() string {
	output, err := xml.MarshalIndent(
		Actions{
			Actions: []Action{
				a,
			},
		}, "", "    ")
	if err != nil {
		logrus.Errorf("error: %v\n", err)
	}

	return string(output[:])
}

// UpdatePipelineURL analyze the local environment to guess if Updatecli is executed from a CI pipeline
func (a *Action) UpdatePipelineURL() {

	// isGitHubActionWorkflow check if the current execution is running from a GitHub Action
	isGitHubActionWorkflow := func() bool {
		if os.Getenv("GITHUB_ACTION") != "" &&
			os.Getenv("GITHUB_SERVER_URL") != "" &&
			os.Getenv("GITHUB_REPOSITORY") != "" &&
			os.Getenv("GITHUB_RUN_ID") != "" {
			logrus.Debugln("GitHub Action pipeline detected")
			return true
		}
		return false
	}

	// isJenkinsPipeline check if the current execution is running from a Jenkins pipeline
	isJenkinsPipeline := func() bool {
		if os.Getenv("JENKINS_URL") != "" &&
			os.Getenv("BUILD_URL") != "" {
			logrus.Debugln("Jenkins build detected")
			return true
		}
		return false
	}

	// isGitLabCI check if the current execution is running from a GitLab CI pipeline
	isGitLabCI := func() bool {
		if os.Getenv("CI_SERVER_URL") != "" &&
			os.Getenv("CI_JOB_URL") != "" {
			logrus.Debugln("GitLab CI pipeline detected")
			return true
		}
		return false
	}

	if isGitHubActionWorkflow() {
		a.PipelineUrl = &PipelineURL{}
		a.PipelineUrl.Name = "GitHub Action pipeline link"
		a.PipelineUrl.URL = fmt.Sprintf(os.Getenv("GITHUB_SERVER_URL")+"/%s/actions/runs/%s", os.Getenv("GITHUB_REPOSITORY"), os.Getenv("GITHUB_RUN_ID"))
	} else if isJenkinsPipeline() {
		a.PipelineUrl = &PipelineURL{}
		a.PipelineUrl.Name = "Jenkins pipeline link"
		a.PipelineUrl.URL = os.Getenv("BUILD_URL")
	} else if isGitLabCI() {
		a.PipelineUrl = &PipelineURL{}
		a.PipelineUrl.Name = "GitLab CI pipeline link"
		a.PipelineUrl.URL = fmt.Sprintf(os.Getenv("CI_SERVER_URL")+"/%s/-/jobs/%s", os.Getenv("CI_PROJECT_PATH"), os.Getenv("CI_JOB_ID"))
	} else {
		logrus.Debugln("No CI pipeline detected")
	}
}
