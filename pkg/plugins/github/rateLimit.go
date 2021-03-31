package github

import "github.com/sirupsen/logrus"

// RateLimit is a struct that contains Github Api limit information
type RateLimit struct {
	Cost      int
	Remaining int
	ResetAt   string
}

// Show display Github Api limit usage
func (a *RateLimit) Show() {
	if (a.Cost * 2) > a.Remaining {
		logrus.Warningf("Running out of Github Api resource, currently used %d remaining %d (reset at %s)",
			a.Cost, a.Remaining, a.ResetAt)
	} else {
		logrus.Debugf("GitHub Api credit used used %d, remaining %d (reset at %s)",
			a.Cost, a.Remaining, a.ResetAt)
	}
}
