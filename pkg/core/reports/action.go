package reports

import (
	"encoding/xml"
	"fmt"
	"sort"
)

type Action struct {
	ID          string         `xml:"id,attr"`
	Title       string         `xml:"h2,omitempty"`
	Description string         `xml:"p,omitempty"`
	Targets     []ActionTarget `xml:"details,omitempty"`
}

type ActionTargetChangelog struct {
	Title       string `xml:"summary,omitempty"`
	Description string `xml:"code,omitempty"`
}

// String show an action report formatted as a string
func (a *Action) String() string {
	if err := a.Sort(); err != nil {
		return ""
	}
	output, err := xml.MarshalIndent(a, "", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	return string(output[:])
}

func (ba *Action) Merge(a *Action) {
	for i := range a.Targets {
		targetFound := false
		for j := range ba.Targets {
			if a.Targets[i].ID == ba.Targets[j].ID {
				targetFound = true
				ba.Targets[j].Merge(&a.Targets[i])
				break
			}
		}
		if !targetFound {
			ba.Targets = append(ba.Targets, a.Targets[i])
		}
	}
}

func (a *Action) Sort() error {
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
	return nil
}
