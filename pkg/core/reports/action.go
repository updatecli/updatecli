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
func (h *Action) String() string {
	if err := h.Sort(); err != nil {
		return ""
	}
	output, err := xml.MarshalIndent(h, "", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	return string(output[:])
}

func (h *Action) Merge(a *Action) {
	for i := range a.Targets {
		targetFound := false
		for j := range h.Targets {
			if a.Targets[i].ID == h.Targets[j].ID {
				targetFound = true
				h.Targets[i].Merge(&a.Targets[j])
				break
			}
		}
		if !targetFound {
			h.Targets = append(h.Targets, a.Targets[i])
		}
	}
}

func (h *Action) Sort() error {
	sort.Slice(
		h.Targets,
		func(i, j int) bool {
			return h.Targets[i].ID < h.Targets[j].ID
		})

	for id, target := range h.Targets {
		sort.Slice(
			target.Changelogs,
			func(i, j int) bool {
				return target.Changelogs[i].Title < target.Changelogs[j].Title
			})
		h.Targets[id] = target
	}
	return nil
}

// Unmarshal parses the htmlReport string and return a struct
func Unmarshal(input []byte, output *Action) (err error) {
	if err := xml.Unmarshal(input, &output); err != nil {
		return err
	}
	if err := output.Sort(); err != nil {
		return err
	}
	return nil
}
