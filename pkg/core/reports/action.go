package reports

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"sort"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/ci"
)

// Action is a struct used to store the result of an action. It is used to generate pullrequest body
type Action struct {
	// ID is the unique identifier of the action
	ID string `xml:"id,attr" json:"id,omitempty"`
	// Title is the title of the action
	Title string `xml:"-" json:"title,omitempty"`
	// PipelineTitle is the title of the pipeline
	PipelineTitle string `xml:"h3,omitempty" json:"pipelineTitle,omitempty"`
	// Description is the description of the action
	Description string `xml:"p,omitempty" json:"description,omitempty"`
	// Targets is the list of targets IDs associated with the action
	Targets []ActionTarget `xml:"details,omitempty" json:"targets,omitempty"`
	// using a pointer to avoid empty tag
	PipelineURL *PipelineURL `xml:"a,omitempty" json:"pipelineURL,omitempty"`
	// Link is the URL of the action
	Link string `xml:"link,omitempty" json:"actionUrl,omitempty"`
}

// ActionTargetChangelog is a struct used to store a target changelog
type ActionTargetChangelog struct {
	// Title is the title of the changelog
	Title string `xml:"summary,omitempty" json:"title,omitempty"`
	// Description is the description of the changelog
	Description string `xml:"pre,omitempty" json:"description,omitempty"`
}

// PipelineURL is a struct used to store a pipeline URL
type PipelineURL struct {
	// URL is the URL of the pipeline
	URL string `xml:"href,attr"`
	// Name is the name of the pipeline
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

func (a *Action) Merge(sourceAction *Action, useDetailsFromSourceAction bool) {
	var c, d []ActionTarget

	useDetailsFromSourceActionTarget := useDetailsFromSourceAction
	switch len(a.Targets) > len(sourceAction.Targets) {
	case true:
		c = a.Targets
		d = sourceAction.Targets
	case false:
		d = a.Targets
		c = sourceAction.Targets
		useDetailsFromSourceActionTarget = !useDetailsFromSourceAction
	}

	for i := range d {
		targetFound := false
		for j := range c {
			if d[i].ID == c[j].ID {
				targetFound = true
				c[j].Merge(&d[i], useDetailsFromSourceActionTarget)
				break
			}
		}
		if !targetFound {
			c = append(c, d[i])
		}
	}

	if useDetailsFromSourceAction {
		a.PipelineTitle = sourceAction.PipelineTitle
		a.Description = sourceAction.Description
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

// updateTargetDescriptions updates descriptions from being console friendly to markdown friendly
func (a *Action) updateTargetDescriptions() {
	for id, target := range a.Targets {
		d := target.Description
		d = strings.Replace(d, "\n\t*", "\n\n*", 1)
		d = strings.ReplaceAll(d, "\n\t*", "\n*")
		a.Targets[id].Description = d
	}
}

// ToActionsString show an action report formatted as a string
func (a Action) ToActionsString() string {
	a.sort()
	a.updateTargetDescriptions()

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
	a.updateTargetDescriptions()

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

	a.PipelineURL = &PipelineURL{
		Name: detectedCi.Name(),
		URL:  detectedCi.URL(),
	}
}
