package reports

import (
	"encoding/xml"
	"fmt"
	"sort"

	"github.com/sirupsen/logrus"
)

type Actions struct {
	Actions []Action `xml:"action"`
}

func (a *Actions) Merge(sourceActions *Actions) {
	var c, d Actions
	useDetailsFromSourceAction := true
	switch len(a.Actions) > len(sourceActions.Actions) {
	case true:
		c = *a
		d = *sourceActions
		useDetailsFromSourceAction = false
	case false:
		d = *a
		c = *sourceActions
	}

	for i := range d.Actions {
		pipelineFound := false
	out:
		for j := range c.Actions {
			if d.Actions[i].ID == c.Actions[j].ID {
				pipelineFound = true
				c.Actions[j].Merge(&d.Actions[i], useDetailsFromSourceAction)
				break out
			}
		}
		if !pipelineFound {
			c.Actions = append(c.Actions, d.Actions[i])
		}
	}
	*a = c
}

// String show an action report formatted as a string
func (a *Actions) String() string {
	a.sort()

	output, err := xml.MarshalIndent(a, "", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	return string(output[:])
}

func MergeFromString(old, new string) string {
	var oldReport Actions
	var newReport Actions

	if old == "" && new != "" {
		return new
	}

	if old != "" && new == "" {
		return old
	}

	err := unmarshal([]byte(old), &oldReport)
	if err != nil {
		logrus.Errorf("failed parsing old report: %s", err)
		// Return the new report by default if something went wrong
		return new
	}

	err = unmarshal([]byte(new), &newReport)
	if err != nil {
		logrus.Errorf("failed parsing new report: %s", err)
		// Return the new report by default
		return new
	}

	newReport.Merge(&oldReport)

	return newReport.String()
}

func (a *Actions) sort() {
	actions := *a
	sort.Slice(
		actions.Actions,
		func(i, j int) bool {
			return actions.Actions[i].ID < actions.Actions[j].ID
		})
}

// unmarshal parses the htmlReport string and return a struct
func unmarshal(input []byte, a *Actions) (err error) {
	if err := xml.Unmarshal(input, a); err != nil {
		return err
	}

	b := *a
	for i := range b.Actions {
		b.Actions[i].sort()
	}
	return nil
}
