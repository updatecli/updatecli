package reports

// ActionTarget holds target data to describe an action report
type ActionTarget struct {
	ID          string                  `xml:"id,attr"`
	Title       string                  `xml:"summary,omitempty"`
	Description string                  `xml:"p,omitempty"`
	Changelogs  []ActionTargetChangelog `xml:"details,omitempty"`
}

func (a *ActionTarget) Merge(sourceActionTarget *ActionTarget, useDetailsFromSourceActionTarget bool) {
	var c, d []ActionTargetChangelog

	switch len(a.Changelogs) > len(sourceActionTarget.Changelogs) {
	case true:
		c = a.Changelogs
		d = sourceActionTarget.Changelogs
	case false:
		d = a.Changelogs
		c = sourceActionTarget.Changelogs
	}

	for i := range d {
		changelogFound := false
		for j := range c {
			if d[i].Title == c[j].Title {
				changelogFound = true
				break
			}
		}
		if !changelogFound {
			c = append(c, d[i])
		}
	}

	if useDetailsFromSourceActionTarget {
		a.Title = sourceActionTarget.Title
		a.Description = sourceActionTarget.Description
	}

	a.Changelogs = c
}
