package github

import (
	"time"

	"github.com/sirupsen/logrus"
)

// RateLimit is a struct that contains GitHub Api limit information
type RateLimit struct {
	Cost      int
	Remaining int
	ResetAt   string
}

// Show display GitHub Api limit usage
// If the remaining credit is 0, it waits until the reset time before returning
func (a *RateLimit) Show() {

	if a.Remaining == 0 {
		resetAtTime, err := time.Parse(time.RFC3339, a.ResetAt)
		if err != nil {
			logrus.Errorf("Parsing GitHub API rate limit reset time: %s", err)
			return
		}
		sleepDuration := time.Until(resetAtTime)

		logrus.Warningf(
			"GitHub API rate limit reached, on hold for %d minute(s) until %s. The process will resume automatically.\n",
			int(sleepDuration.Minutes()),
			resetAtTime.UTC().Format("2006-01-02 15:04:05 UTC"))

		time.Sleep(sleepDuration)
		return
	}

	logrus.Debugf("GitHub API credit used %d, remaining %d (reset at %s)",
		a.Cost, a.Remaining, a.ResetAt)
}
