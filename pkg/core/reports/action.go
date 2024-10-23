package reports

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"sort"
	"text/template"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/ci"
)

// Action is a struct used to store the result of an action. It is used to generate pullrequest body
type Action struct {
	ID            string         `xml:"id,attr"`
	Title         string         `xml:"-"`
	PipelineTitle string         `xml:"h3,omitempty"`
	Description   string         `xml:"p,omitempty"`
	Targets       []ActionTarget `xml:"details,omitempty"`
	// using a pointer to avoid empty tag
	PipelineUrl *PipelineURL `xml:"a,omitempty"`
}

// ActionTargetChangelog is a struct used to store a target changelog
type ActionTargetChangelog struct {
	Title       string `xml:"summary,omitempty"`
	Description string `xml:"pre,omitempty"`
}

// PipelineURL is a struct used to store a pipeline URL
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

// ToActionsMarkdownString show an action report formatted as a string using markdown
func (a Action) ToActionsMarkdownString() string {
	tmpl, err := template.New("actions").Parse(markdownReportTemplate)
	if err != nil {
		logrus.Errorf("error: %v\n", err)
	}

	manifest := bytes.Buffer{}
	if err := tmpl.Execute(&manifest, a); err != nil {
		logrus.Debugln(err)
		logrus.Errorf("error: %v\n", err)
	}
	return manifest.String()
}

// UpdatePipelineURL analyze the local environment to guess if Updatecli is executed from a CI pipeline
func (a *Action) UpdatePipelineURL() {
	detectedCi, err := ci.New()
	if err != nil {
		logrus.Debugf("No CI pipeline detected (%s)\n", err)
	}

	if detectedCi == nil {
		// No CI pipeline detected
		return
	}

	a.PipelineUrl = &PipelineURL{
		Name: detectedCi.Name(),
		URL:  detectedCi.URL(),
	}
}
