package reports

import (
	"encoding/xml"
	"fmt"
	"sort"
)

type htmlReport struct {
	ID          string             `xml:"id,attr"`
	Title       string             `xml:"h2,"`
	Description string             `xml:"p,"`
	Targets     []targetHTMLReport `xml:"details,"`
}

type targetHTMLReport struct {
	ID          string          `xml:"id,attr"`
	Title       string          `xml:"summary,"`
	Description string          `xml:"p,"`
	Changelogs  []HTMLChangelog `xml:"details,"`
}

type HTMLChangelog struct {
	Title       string `xml:"summary,"`
	Description string `xml:"p,"`
}

func (h *htmlReport) String() string {
	output, err := xml.MarshalIndent(h, "", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	return string(output[:])
}

func (h *htmlReport) Merge(a *htmlReport) {
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

func (h *targetHTMLReport) Merge(a *targetHTMLReport) {
	for i := range a.Changelogs {
		changelogFound := false
		for j := range h.Changelogs {
			if a.Changelogs[i].Title == h.Changelogs[j].Title {
				changelogFound = true
				break
			}
		}
		if !changelogFound {
			h.Changelogs = append(h.Changelogs, a.Changelogs[i])
		}
	}
}

// Unmarshal parses the htmlReport string and return a struct
func Unmarshal(input []byte, output *htmlReport) (err error) {
	if err := xml.Unmarshal(input, &output); err != nil {
		return err
	}
	if err := output.Sort(); err != nil {
		return err
	}
	return nil
}

func (h *htmlReport) Sort() error {
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
