package reports

// ActionTarget holds target data to describe an action report
type ActionTarget struct {
	ID          string                  `xml:"id,attr"`
	Title       string                  `xml:"summary,omitempty"`
	Description string                  `xml:"p,omitempty"`
	Changelogs  []ActionTargetChangelog `xml:"details,omitempty"`
}

func (h *ActionTarget) Merge(a *ActionTarget) {
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
