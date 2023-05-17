package github

import "github.com/sirupsen/logrus"

// RateLimit is a struct that contains GitHub Api limit information
type RateLimit struct {
	Cost      int
	Remaining int
	ResetAt   string
}

// Show display GitHub Api limit usage
func (a *RateLimit) Show() {
	if (a.Cost * 2) > a.Remaining {
		logrus.Warningf("Running out of GitHub Api resource, currently used %d remaining %d (reset at %s)",
			a.Cost, a.Remaining, a.ResetAt)
	} else {
		logrus.Debugf("GitHub Api credit used %d, remaining %d (reset at %s)",
			a.Cost, a.Remaining, a.ResetAt)
	}
}
